package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

const (
	Client = "client"
	Server = "http_server"
	Book   = "book_service"
)

type Config struct {
	Client     ClientConfig `yaml:"client"`
	HTTPServer ServerConfig `yaml:"http_server"`
	BookServer ServerConfig `yaml:"book_server"`
}

var config *Config

func InitReadConfig() {
	configFile, err := os.ReadFile("internal/config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		log.Fatalf("Failed to unmarshal config file: %v", err)
	}
}

func ReadConfig(key string) interface{} {

	switch key {
	case Client:
		return config.Client
	case Server:
		return config.HTTPServer
	case Book:
		return config.BookServer
	default:
		return nil
	}
}
