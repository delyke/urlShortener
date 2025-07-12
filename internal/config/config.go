package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"log"
)

type Config struct {
	RunAddr  string `env:"SERVER_ADDRESS"`
	BaseAddr string `env:"BASE_URL"`
	LogLevel string `env:"LOG_LEVEL"`
}

func GetConfig() *Config {
	runAddr := flag.String("a", ":8080", "Run server address")
	baseAddr := flag.String("b", "http://localhost:8080", "Base server address")
	logLevel := flag.String("l", "info", "Log level")

	flag.Parse()

	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	if cfg.RunAddr == "" {
		cfg.RunAddr = *runAddr
	}

	if cfg.BaseAddr == "" {
		cfg.BaseAddr = *baseAddr
	}

	if cfg.LogLevel == "" {
		cfg.LogLevel = *logLevel
	}

	return &cfg
}
