package migrations

import (
	"github.com/delyke/urlShortener/internal/config/db"
	"github.com/delyke/urlShortener/internal/model"
	"log"
)

func Run() {
	err := db.Get().AutoMigrate(&model.URL{})
	if err != nil {
		panic(err)
	}

	log.Println("Миграции выполнены")
}
