package main

import (
	"log"
	"net/http"
	"time"

	"github.com/delyke/urlShortener/internal/app"
	"github.com/delyke/urlShortener/internal/audit"
	"github.com/delyke/urlShortener/internal/config"
	"github.com/delyke/urlShortener/internal/handler"
	"github.com/delyke/urlShortener/internal/logger"
	"github.com/delyke/urlShortener/internal/repository"
	"github.com/delyke/urlShortener/internal/service"
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
	svc := service.NewURLService(repo, cfg, 10*time.Second, 100)
	svc.StartDeleter()
	defer svc.StopDeleter()

	if cfg.AuditFile != "" {
		observer, err := audit.NewFileObserver(cfg.AuditFile)
		if err != nil {
			l.Fatal("Failed to initialize audit file observer: ", err)
		}
		svc.RegisterAuditObserver(observer)
	}

	if cfg.AuditURL != "" {
		observer, err := audit.NewHTTPObserver(cfg.AuditURL, nil)
		if err != nil {
			l.Fatal("Failed to initialize audit http observer: ", err)
		}
		svc.RegisterAuditObserver(observer)
	}

	h := handler.NewHandler(svc, cfg, l)
	l.Info("Running server on", cfg.RunAddr)

	err = http.ListenAndServe(cfg.RunAddr, app.NewRouter(h, l, cfg, svc))
	if err != nil {
		l.Fatal("Failed listen and serve:", err)
	}
}
