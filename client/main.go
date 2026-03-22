package main

import (
	"os"
	"os/signal"
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
		Bet: domain.BetRequest{
			Agency:    v.GetString("id"),
			FirstName: os.Getenv("NOMBRE"),
			LastName:  os.Getenv("APELLIDO"),
			Document:  os.Getenv("DOCUMENTO"),
			Birthdate: os.Getenv("NACIMIENTO"),
			Number:    os.Getenv("NUMERO"),
		},
	}, nil
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
		"action: config | result: success | client_id: %s | server_address: %s | log_level: %s",
		config.ID,
		config.ServerAddress,
		config.LogLevel,
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

	printConfig(config)

	betClient := client.New(
		config.ServerAddress,
		config.ID,
		protocol.NewBetMessageEncoder(),
		protocol.NewServerResponseDecoder(),
	)
	setTerminateHandler(betClient)

	response, err := betClient.SendBet(config.Bet)
	if err != nil {
		log.Errorf(
			"action: apuesta_enviada | result: fail | dni: %s | numero: %s | error: %v",
			config.Bet.Document,
			config.Bet.Number,
			err,
		)
		return
	}

	if !response.IsSuccess() {
		log.Errorf(
			"action: apuesta_enviada | result: fail | dni: %s | numero: %s | error_code: %s | error_message: %s",
			config.Bet.Document,
			config.Bet.Number,
			response.Code,
			response.Message,
		)
		return
	}

	log.Infof(
		"action: apuesta_enviada | result: success | dni: %s | numero: %s",
		config.Bet.Document,
		config.Bet.Number,
	)
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
