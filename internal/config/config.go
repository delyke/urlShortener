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
	AuthCookieName  string `env:"AUTH_COOKIE_NAME"`
	HMACSecret      string `env:"HMAC_SECRET"`
	AuditFile       string `env:"AUDIT_FILE"`
	AuditURL        string `env:"AUDIT_URL"`
	PprofEnabled    bool   `env:"PPROF_ENABLED"`
}

func GetConfig() (*Config, error) {
	runAddr := flag.String("a", ":8080", "Run server address")
	baseAddr := flag.String("b", "http://localhost:8080", "Base server address")
	logLevel := flag.String("l", "info", "Log level")
	fileStoragePath := flag.String("f", "", "File storage path")
	databaseDSN := flag.String("d", "", "Database DSN")
	authCookieName := flag.String("c", "Authorization", "Auth cookie name")
	hmacSecret := flag.String("s", "default-secret", "HMAC secret")
	auditFile := flag.String("audit-file", "", "Audit file path")
	auditURL := flag.String("audit-url", "", "Audit server URL")
	pprofEnabled := flag.Bool("pprof", false, "Enable pprof endpoint")

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

	if cfg.AuthCookieName == "" {
		cfg.AuthCookieName = *authCookieName
	}

	if cfg.HMACSecret == "" {
		cfg.HMACSecret = *hmacSecret
	}

	if cfg.AuditFile == "" {
		cfg.AuditFile = *auditFile
	}

	if cfg.AuditURL == "" {
		cfg.AuditURL = *auditURL
	}

	if !cfg.PprofEnabled {
		cfg.PprofEnabled = *pprofEnabled
	}

	return &cfg, nil
}
