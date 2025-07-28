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
	repo, err := repository.NewPostgresRepository(cfg.DatabaseDSN)
	if err != nil {
		log.Fatal("Failed to initialize repo: ", err)
	}
	svc := service.NewURLService(repo)
	h := handler.NewHandler(svc, cfg)
	log.Println("Running server on", cfg.RunAddr)
	l, err := logger.Initialize(cfg.LogLevel)
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer l.Sync()
	err = http.ListenAndServe(cfg.RunAddr, app.NewRouter(h, l))
	if err != nil {
		log.Fatal("Failed listen and serve:", err)
	}
}
