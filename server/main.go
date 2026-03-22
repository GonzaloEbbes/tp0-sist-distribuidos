package main

import (
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/op/go-logging"
	"gopkg.in/ini.v1"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/server/common"
)

var log = logging.MustGetLogger("log")

type configParams struct {
	port          int
	listenBacklog int
	loggingLevel  string
}

// Parse environment variables or config file to find program config params
//
// This function looks for program configuration parameters in the
// environment first and then in a config file.
// If a required key is missing or a parameter cannot be parsed,
// an error is returned.
// If parsing succeeds, the function returns the parsed config parameters.
func initializeConfig() (configParams, error) {
	cfg := ini.Empty()

	// If config.ini does not exists original config object is not modified
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

	port, err := strconv.Atoi(portValue)
	if err != nil {
		return configParams{}, err
	}

	listenBacklog, err := strconv.Atoi(listenBacklogValue)
	if err != nil {
		return configParams{}, err
	}

	return configParams{
		port:          port,
		listenBacklog: listenBacklog,
		loggingLevel:  loggingLevelValue,
	}, nil
}

// Logging initialization
//
// Current timestamp is added to be able to identify in docker
// compose logs the date when the log has arrived
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

	// Log config parameters at the beginning of the program to verify the configuration
	// of the component
	log.Debugf("action: config | result: success | port: %d | listen_backlog: %d | logging_level: %s",
		config.port,
		config.listenBacklog,
		config.loggingLevel,
	)

	// Initialize server and start server loop
	server, err := common.NewServer(config.port, config.listenBacklog)
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
