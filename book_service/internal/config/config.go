package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

const (
	Book = "book_service"
	DB   = "DB"
)

type Config struct {
	BookServer ServerConfig   `yaml:"book_server"`
	DB         DatabaseConfig `yaml:"DB"`
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
	case Book:
		return config.BookServer
	case DB:
		return config.DB
	default:
		return nil
	}
}
