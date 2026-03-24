package main

import (
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/op/go-logging"
	"gopkg.in/ini.v1"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/server/common"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/server/internal/infrastructure/protocol"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/server/internal/infrastructure/repository"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/server/internal/usecase"
)

var log = logging.MustGetLogger("log")

type configParams struct {
	port             int
	listenBacklog    int
	loggingLevel     string
	expectedAgencies int
}

const DEFAULT_EXPECTED_AGENCIES = 5

func initializeConfig() (configParams, error) {
	cfg := ini.Empty()
	_ = cfg.Append("./config.ini")

	defaultSection := cfg.Section("DEFAULT")

	portValue := os.Getenv("SERVER_PORT")
	if portValue == "" {
		portValue = defaultSection.Key("SERVER_PORT").String()
	}

	listenBacklogValue := os.Getenv("SERVER_LISTEN_BACKLOG")
	if listenBacklogValue == "" {
		listenBacklogValue = defaultSection.Key("SERVER_LISTEN_BACKLOG").String()
	}

	loggingLevelValue := os.Getenv("LOGGING_LEVEL")
	if loggingLevelValue == "" {
		loggingLevelValue = defaultSection.Key("LOGGING_LEVEL").String()
	}

	expectedAgenciesValue := os.Getenv("SERVER_EXPECTED_AGENCIES")
	if expectedAgenciesValue == "" {
		expectedAgenciesValue = defaultSection.Key("SERVER_EXPECTED_AGENCIES").MustString(strconv.Itoa(DEFAULT_EXPECTED_AGENCIES))
	}

	port, err := strconv.Atoi(portValue)
	if err != nil {
		return configParams{}, err
	}

	listenBacklog, err := strconv.Atoi(listenBacklogValue)
	if err != nil {
		return configParams{}, err
	}

	expectedAgencies, err := strconv.Atoi(expectedAgenciesValue)
	if err != nil {
		return configParams{}, err
	}

	return configParams{
		port:             port,
		listenBacklog:    listenBacklog,
		loggingLevel:     loggingLevelValue,
		expectedAgencies: expectedAgencies,
	}, nil
}

func initializeLog(loggingLevel string) error {
	baseBackend := logging.NewLogBackend(os.Stdout, "", 0)
	format := logging.MustStringFormatter(
		`%{time:2006-01-02 15:04:05} %{level:.5s}     %{message}`,
	)
	backendFormatter := logging.NewBackendFormatter(baseBackend, format)

	backendLeveled := logging.AddModuleLevel(backendFormatter)
	logLevelCode, err := logging.LogLevel(loggingLevel)
	if err != nil {
		return err
	}
	backendLeveled.SetLevel(logLevelCode, "")

	logging.SetBackend(backendLeveled)
	return nil
}

func main() {
	config, err := initializeConfig()
	if err != nil {
		log.Criticalf("%s", err)
		return
	}

	if err := initializeLog(config.loggingLevel); err != nil {
		log.Criticalf("%s", err)
		return
	}

	log.Debugf("action: config | result: success | port: %d | listen_backlog: %d | logging_level: %s | expected_agencies: %d",
		config.port,
		config.listenBacklog,
		config.loggingLevel,
		config.expectedAgencies,
	)

	drawState := usecase.NewDrawState(config.expectedAgencies)
	betRepository := repository.NewBetRepository()
	registerBet := usecase.NewRegisterBet(betRepository)
	finishAgency := usecase.NewFinishAgency(drawState)
	queryWinners := usecase.NewQueryWinners(drawState, betRepository)
	server, err := common.NewServer(
		config.port,
		config.listenBacklog,
		protocol.NewMessageProcessor(registerBet, finishAgency, queryWinners),
		protocol.NewResponseEncoder(),
	)
	if err != nil {
		log.Criticalf("%s", err)
		return
	}

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, syscall.SIGTERM)

	go func() {
		<-terminate
		stopServer(server)
	}()

	server.Run()
}

func stopServer(server *common.Server) {
	log.Info("Termination signal received. Stopping server...")
	server.Stop()
}
