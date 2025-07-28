package main

import (
	"github.com/delyke/urlShortener/internal/app"
	"github.com/delyke/urlShortener/internal/config"
	"github.com/delyke/urlShortener/internal/handler"
	"github.com/delyke/urlShortener/internal/logger"
	"github.com/delyke/urlShortener/internal/repository"
	"github.com/delyke/urlShortener/internal/service"
	"log"
	"net/http"
)

func main() {
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal("Failed to initialize config: ", err)
	}
	l, err := logger.Initialize(cfg.LogLevel)
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer l.Sync()
	var repo repository.URLRepository
	if cfg.DatabaseDSN != "" {
		l.Info("Using database Postgres: ", cfg.DatabaseDSN)
		repo, err = repository.NewPostgresRepository(cfg.DatabaseDSN)
	} else if cfg.FileStoragePath != "" {
		l.Info("Using infile database: ", cfg.FileStoragePath)
		repo, err = repository.NewFileRepository(cfg.FileStoragePath)
	} else {
		l.Info("Using inmemory database")
		repo, err = repository.NewLocalRepository()
	}

	if err != nil {
		l.Fatal("Failed to initialize repo: ", err)
	}
	svc := service.NewURLService(repo)
	h := handler.NewHandler(svc, cfg)
	l.Info("Running server on", cfg.RunAddr)

	err = http.ListenAndServe(cfg.RunAddr, app.NewRouter(h, l))
	if err != nil {
		l.Fatal("Failed listen and serve:", err)
	}
}
