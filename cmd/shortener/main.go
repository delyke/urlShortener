package main

import (
	"fmt"
	"github.com/delyke/urlShortener/internal/app"
	"github.com/delyke/urlShortener/internal/config"
	"github.com/delyke/urlShortener/internal/config/db"
	"github.com/delyke/urlShortener/internal/handler"
	"github.com/delyke/urlShortener/internal/repository"
	"github.com/delyke/urlShortener/internal/service"
	"github.com/delyke/urlShortener/migrations"
	"github.com/joho/godotenv"
	"net/http"
	"reflect"
)

func main() {
	//repo := repository.NewMySQLRepository(db.Get())
	config.ParseFlags()
	repo := repository.NewLocalRepository()
	if reflect.TypeOf(repo).String() == "*repository.MySQLRepository" {
		err := godotenv.Load(".env")
		if err != nil {
			panic(err)
		}
		db.Init()
		migrations.Run()
	}
	svc := service.NewURLService(repo)
	h := handler.NewHandler(svc)
	fmt.Println("Running server on", config.FlagRunAddr)
	err := http.ListenAndServe(config.FlagRunAddr, app.NewRouter(h))
	if err != nil {
		panic(err)
	}
}
