package main

import (
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

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			h.HandleGet(w, r)
			return
		} else if r.Method == "POST" && r.URL.Path == "/" {
			h.HandlePost(w, r)
			return
		} else {
			http.Error(w, "Bad Request", http.StatusBadRequest)
		}
	})
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
