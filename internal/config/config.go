package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
)

type Config struct {
	RunAddr         string `env:"SERVER_ADDRESS"`
	BaseAddr        string `env:"BASE_URL"`
	LogLevel        string `env:"LOG_LEVEL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
}

func GetConfig() (*Config, error) {
	runAddr := flag.String("a", ":8080", "Run server address")
	baseAddr := flag.String("b", "http://localhost:8080", "Base server address")
	logLevel := flag.String("l", "info", "Log level")
	fileStoragePath := flag.String("f", "storage.json", "File storage path")
	databaseDSN := flag.String("d", "", "Database DSN")

	flag.Parse()

	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {

		return nil, err
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

	if cfg.FileStoragePath == "" {
		cfg.FileStoragePath = *fileStoragePath
	}

	if cfg.DatabaseDSN == "" {
		cfg.DatabaseDSN = *databaseDSN
	}

	return &cfg, nil
}
