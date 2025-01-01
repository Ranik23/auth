package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

type GRPCServerConfig struct {
	GRPCServerAddress string `json:"grpc_server_address" yaml:"grpc_server_address" validate:"required"`
	GRPCServerPort    int    `json:"grpc_server_port" yaml:"grpc_server_port" validate:"required,min=1,max=65535"`
	GRPCServerTLS     bool   `json:"grpc_server_tls" yaml:"grpc_server_tls"`
}

type DataBaseConfig struct {
	DatabaseHost     string `json:"database_host" yaml:"database_host" validate:"required"`
	DatabasePort     int    `json:"database_port" yaml:"database_port" validate:"required,min=1,max=65535"`
	DatabaseUser     string `json:"database_user" yaml:"database_user" validate:"required"`
	DatabasePassword string `json:"database_password" yaml:"database_password" validate:"required"`
	DatabaseName     string `json:"database_name" yaml:"database_name" validate:"required"`
}

type SMTPConfig struct {
	Email string 		`yaml:"email"`
	Password string 	`yaml:"password"`
}

type Config struct {
	GRPCServerConfig `json:"grpc_server_config" yaml:"grpc_server_config"`
	DataBaseConfig	 `json:"database_config" yaml:"database_config"`
	SMTPConfig		 `yaml:"smtp"`
}


func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return &config, nil
}