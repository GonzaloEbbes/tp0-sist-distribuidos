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

func main() {
	if err := run(os.Args); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: composer <output_file> <client_count>")
	}

	outputFile := args[1]
	clientCountStr := args[2]

	clientCount, err := strconv.Atoi(clientCountStr)
	if err != nil {
		return fmt.Errorf("invalid client count %q: %w", clientCountStr, err)
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

	addServices(&AppConfig, clientCount)

	rawFileContent, err := generateRawFileContent(AppConfig)
	if err != nil {
		return fmt.Errorf("generate file content: %w", err)
	}

	if err := writeFile(outputFile, rawFileContent); err != nil {
		return fmt.Errorf("write compose file %q: %w", outputFile, err)
	}

	return nil
}

func addServices(appConfig *AppConfig, clientCount int) {
	services := make(map[string]ServiceConfig, clientCount+1) // +1 para el server
	services["server"] = ServiceConfig{
		ContainerName: "server",
		Image:         "server:latest",
		Entrypoint:    "/server",
		Environment:   []string{},
		Networks:      []string{"testing_net"},
		Volumes:       []string{"./server/config.ini:/config.ini"},
		DependsOn: nil,
	}

	for i := 1; i <= clientCount; i++ {
		datasetPath := fmt.Sprintf("/data/agency-%d.csv", i)
		clientService := ServiceConfig{
			ContainerName: fmt.Sprintf("client%d", i),
			Image:         "client:latest",
			Entrypoint:    "/client",
			Environment: []string{
				"CLI_ID=" + fmt.Sprintf("%d", i),
				"CLI_DATASET_FILEPATH=" + datasetPath,
			},
			DependsOn: []string{"server"},
			Networks:  []string{"testing_net"},
			Volumes: []string{
				"./client/config.yaml:/config.yaml",
				fmt.Sprintf("./.data/dataset/agency-%d.csv:%s", i, datasetPath),
			},
		}
		services[fmt.Sprintf("client%d", i)] = clientService
	}

	appConfig.Services = services
}

func generateRawFileContent(appConfig AppConfig) ([]byte, error) {
	yamlContent, err := yaml.Marshal(appConfig)
	if err != nil {
		return nil, err
	}
	return yamlContent, nil
}

func writeFile(filename string, content []byte) error {
	const perm os.FileMode = 0644
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}

	bytes, err := f.Write(content)
	if err != nil {
		_ = f.Close()
		return err
	}
	if bytes != len(content) {
		_ = f.Close()
		if removeErr := os.Remove(filename); removeErr != nil {
			return fmt.Errorf("incomplete write: %w (cleanup failed: %v)", errors.New("incomplete write"), removeErr)
		}
		return errors.New("incomplete write")
	}
	if err = f.Close(); err != nil {
		return err
	}

	return nil
}
