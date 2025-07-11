package main

import (
	"github.com/delyke/urlShortener/internal/app"
	"github.com/delyke/urlShortener/internal/config"
	"github.com/delyke/urlShortener/internal/handler"
	"github.com/delyke/urlShortener/internal/repository"
	"github.com/delyke/urlShortener/internal/service"
	"log"
	"net/http"
)

func main() {
	cfg := config.GetConfig()
	repo := repository.NewLocalRepository()
	svc := service.NewURLService(repo)
	h := handler.NewHandler(svc, cfg)
	log.Println("Running server on", cfg.RunAddr)
	err := http.ListenAndServe(cfg.RunAddr, app.NewRouter(h))
	if err != nil {
		log.Fatal(err)
	}
}
