package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	yaml "gopkg.in/yaml.v3"
)

type AppConfig struct {
	Name     string                   `yaml:"name"`
	Services map[string]ServiceConfig `yaml:"services"`
	Networks map[string]NetworkConfig `yaml:"networks"`
}

type ServiceConfig struct {
	ContainerName string   `yaml:"container_name"`
	Image         string   `yaml:"image"`
	Entrypoint    string   `yaml:"entrypoint"`
	Environment   []string `yaml:"environment"`
	DependsOn     []string `yaml:"depends_on,omitempty"`
	Networks      []string `yaml:"networks"`
	Volumes       []string `yaml:"volumes,omitempty"`
}

type NetworkConfig struct {
	Ipam IpamConfig `yaml:"ipam"`
}

type IpamConfig struct {
	Driver string         `yaml:"driver"`
	Config []SubnetConfig `yaml:"config"`
}

type SubnetConfig struct {
	Subnet string `yaml:"subnet"`
}

type BetEnvConfig struct {
	FirstName string
	LastName  string
	Document  string
	Birthdate string
	Number    string
}

var defaultBetEnvConfig = BetEnvConfig{
	FirstName: "Santiago Lionel",
	LastName:  "Lorca",
	Document:  "30904465",
	Birthdate: "1999-03-17",
	Number:    "7574",
}

// TODO: Eliminar los prints y handlear errores
func main() {
	args := os.Args
	if len(args) < 3 {
		//TODO:
		// faltan parámetros
		fmt.Println("Usage: composer <output_file> <client_count> [nombre] [apellido] [documento] [nacimiento] [numero]")
		return
	}

	outputFile := args[1]
	clientCountStr := args[2]
	betEnvConfig := parseBetEnvConfig(args[3:])

	clientCount, err := strconv.Atoi(clientCountStr)
	if err != nil {
		// TODO: manejar error de conversión
		fmt.Printf("Invalid client count: %s\n", clientCountStr)
		return
	}
	AppConfig := AppConfig{
		Name: "tp0",
		Networks: map[string]NetworkConfig{
			"testing_net": {
				Ipam: IpamConfig{
					Driver: "default",
					Config: []SubnetConfig{
						{
							Subnet: "172.25.125.0/24",
						},
					},
				},
			},
		},
	}

	addServices(&AppConfig, clientCount, betEnvConfig)

	rawFileContent, err := generateRawFileContent(AppConfig)
	if err != nil {
		// TODO: manejar error
		fmt.Printf("Error generating file content: %v\n", err)
		return
	}
	err = writeFile(outputFile, rawFileContent)
}

func addServices(appConfig *AppConfig, clientCount int, betEnvConfig BetEnvConfig) {
	services := make(map[string]ServiceConfig, clientCount+1) // +1 para el server
	services["server"] = ServiceConfig{
		ContainerName: "server",
		Image:         "server:latest",
		Entrypoint:    "/server",
		Environment:   []string{},
		Networks:      []string{"testing_net"},
		// TODO: por ahora solo montamos el file, podria servir montar el dir de server
		Volumes:   []string{"./server/config.ini:/config.ini"},
		DependsOn: nil,
	}

	for i := 1; i <= clientCount; i++ {
		clientService := ServiceConfig{
			ContainerName: fmt.Sprintf("client%d", i),
			Image:         "client:latest",
			Entrypoint:    "/client",
			Environment: []string{
				"CLI_ID=" + fmt.Sprintf("%d", i),
				"NOMBRE=" + betEnvConfig.FirstName,
				"APELLIDO=" + betEnvConfig.LastName,
				"DOCUMENTO=" + betEnvConfig.Document,
				"NACIMIENTO=" + betEnvConfig.Birthdate,
				"NUMERO=" + betEnvConfig.Number,
			},
			DependsOn: []string{"server"},
			Networks:  []string{"testing_net"},
			Volumes:   []string{"./client/config.yaml:/config.yaml"},
		}
		services[fmt.Sprintf("client%d", i)] = clientService
	}

	appConfig.Services = services
}

func parseBetEnvConfig(args []string) BetEnvConfig {
	betEnvConfig := defaultBetEnvConfig

	if len(args) > 0 && args[0] != "" {
		betEnvConfig.FirstName = args[0]
	}
	if len(args) > 1 && args[1] != "" {
		betEnvConfig.LastName = args[1]
	}
	if len(args) > 2 && args[2] != "" {
		betEnvConfig.Document = args[2]
	}
	if len(args) > 3 && args[3] != "" {
		betEnvConfig.Birthdate = args[3]
	}
	if len(args) > 4 && args[4] != "" {
		betEnvConfig.Number = args[4]
	}

	return betEnvConfig
}

func generateRawFileContent(appConfig AppConfig) ([]byte, error) {
	yamlContent, err := yaml.Marshal(appConfig)
	if err != nil {
		return nil, err
	}
	return yamlContent, nil
}

func writeFile(filename string, content []byte) error {
	var perm os.FileMode = 0644 // TODO: explicar
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	bytes, err := f.Write(content)
	if err != nil {
		return err
	}
	if bytes != len(content) {
		// TODO: delete file if write was incomplete, then return error
		return errors.New("incomplete write")
	}
	if err = f.Close(); err != nil {
		return err
	}

	return nil
}
