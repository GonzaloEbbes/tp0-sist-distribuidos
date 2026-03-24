package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/op/go-logging"
	"github.com/spf13/viper"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/internal/client"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/internal/domain"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/internal/protocol"
)

var log = logging.MustGetLogger("log")

type clientConfig struct {
	ID            string
	ServerAddress string
	LogLevel      string
	DatasetPath   string
	MaxBatchSize  int
	Bet           domain.BetRequest
}

func initConfig() (*viper.Viper, error) {
	v := viper.New()
	v.AutomaticEnv()
	v.SetEnvPrefix("cli")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.BindEnv("id")
	v.BindEnv("server.address")
	v.BindEnv("log.level")
	v.BindEnv("dataset.filepath")
	v.BindEnv("batch.maxAmount")

	v.SetConfigFile("./config.yaml")
	if err := v.ReadInConfig(); err != nil {
		log.Warning("Configuration could not be read from config file. Using env variables instead")
	}

	return v, nil
}

func loadClientConfig() (clientConfig, error) {
	v, err := initConfig()
	if err != nil {
		return clientConfig{}, err
	}

	return clientConfig{
		ID:            v.GetString("id"),
		ServerAddress: v.GetString("server.address"),
		LogLevel:      v.GetString("log.level"),
		DatasetPath:   resolveDatasetPath(v.GetString("dataset.filepath"), v.GetString("id")),
		MaxBatchSize:  v.GetInt("batch.maxAmount"),
	}, nil
}

func resolveDatasetPath(path string, clientID string) string {
	if path == "" {
		return fmt.Sprintf("./agency-%s.csv", clientID)
	}
	if strings.Contains(path, "%s") {
		return fmt.Sprintf(path, clientID)
	}
	return path
}

func loadBetsFromCSV(csvPath string, agencyID string) (domain.BetRequest, error) {
	file, err := os.Open(filepath.Clean(csvPath))
	if err != nil {
		return domain.BetRequest{}, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return domain.BetRequest{}, err
	}

	bets := make([]domain.Bet, 0, len(records))
	for index, record := range records {
		if len(record) != 5 {
			return domain.BetRequest{}, fmt.Errorf("invalid csv row %d: expected 5 fields, got %d", index+1, len(record))
		}
		bets = append(bets, domain.Bet{
			Agency:    agencyID,
			FirstName: record[0],
			LastName:  record[1],
			Document:  record[2],
			Birthdate: record[3],
			Number:    record[4],
		})
	}

	return newBetRequest(bets), nil
}

func newBetRequest(bets []domain.Bet) domain.BetRequest {
	return domain.BetRequest{
		Type:   domain.MessageTypeBetBatch,
		Agency: bets[0].Agency,
		Bets:   bets,
	}
}

func initLogger(logLevel string) error {
	baseBackend := logging.NewLogBackend(os.Stdout, "", 0)
	format := logging.MustStringFormatter(
		`%{time:2006-01-02 15:04:05} %{level:.5s}     %{message}`,
	)
	backendFormatter := logging.NewBackendFormatter(baseBackend, format)

	backendLeveled := logging.AddModuleLevel(backendFormatter)
	logLevelCode, err := logging.LogLevel(logLevel)
	if err != nil {
		return err
	}
	backendLeveled.SetLevel(logLevelCode, "")

	logging.SetBackend(backendLeveled)
	return nil
}

func printConfig(config clientConfig) {
	log.Infof(
		"action: config | result: success | client_id: %s | server_address: %s | log_level: %s | dataset_path: %s | batch_max_amount: %d",
		config.ID,
		config.ServerAddress,
		config.LogLevel,
		config.DatasetPath,
		config.MaxBatchSize,
	)
}

func main() {
	config, err := loadClientConfig()
	if err != nil {
		log.Criticalf("%s", err)
		return
	}

	if err := initLogger(config.LogLevel); err != nil {
		log.Criticalf("%s", err)
		return
	}

	config.Bet, err = loadBetsFromCSV(config.DatasetPath, config.ID)
	if err != nil {
		log.Criticalf("%s", err)
		return
	}

	printConfig(config)

	betClient := client.New(
		config.ServerAddress,
		config.ID,
		config.MaxBatchSize,
		protocol.NewBetMessageEncoder(),
		protocol.NewServerResponseDecoder(),
	)
	setTerminateHandler(betClient)

	response, err := betClient.SendBet(config.Bet)
	if err != nil {
		log.Errorf(
			"action: apuesta_enviada | result: fail | cantidad: %d | error: %v",
			len(config.Bet.Bets),
			err,
		)
		return
	}

	if !response.IsSuccess() {
		log.Errorf(
			"action: apuesta_enviada | result: fail | cantidad: %d | error_code: %s | error_message: %s",
			len(config.Bet.Bets),
			response.Code,
			response.Message,
		)
		return
	}

	log.Infof(
		"action: apuesta_enviada | result: success | cantidad: %d",
		len(config.Bet.Bets),
	)

	finishResponse, err := betClient.SendRequest(domain.BetRequest{
		Type:   domain.MessageTypeFinish,
		Agency: config.ID,
	})
	if err != nil {
		log.Errorf("action: fin_agencia | result: fail | agency: %s | error: %v", config.ID, err)
		return
	}
	if !finishResponse.IsSuccess() {
		log.Errorf("action: fin_agencia | result: fail | agency: %s | error_code: %s | error_message: %s", config.ID, finishResponse.Code, finishResponse.Message)
		return
	}
	log.Infof("action: fin_agencia | result: success | agency: %s", config.ID)

	winnersResponse, err := betClient.QueryWinnersUntilReady(domain.BetRequest{
		Type:   domain.MessageTypeWinnersQuery,
		Agency: config.ID,
	})
	if err != nil {
		log.Errorf("action: consulta_ganadores | result: fail | error: %v", err)
		return
	}
	if !winnersResponse.IsSuccess() {
		log.Errorf(
			"action: consulta_ganadores | result: fail | error_code: %s | error_message: %s",
			winnersResponse.Code,
			winnersResponse.Message,
		)
		return
	}

	log.Infof(
		"action: consulta_ganadores | result: success | cant_ganadores: %d",
		countWinners(winnersResponse.Message),
	)
}

func countWinners(message string) int {
	if message == "" {
		return 0
	}
	return len(strings.Split(message, ";"))
}

func setTerminateHandler(client *client.Client) {
	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, syscall.SIGTERM)

	go func() {
		<-terminate
		log.Info("Termination signal received. Stopping client...")
		client.Close()
		os.Exit(0)
	}()
}
