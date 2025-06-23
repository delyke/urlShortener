package db

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"os"
	"sync"
)

var (
	db  *gorm.DB
	mux sync.Mutex
)

// Init Инициализация подключения к БД
func Init() *gorm.DB {
	mux.Lock()
	defer mux.Unlock()
	if db != nil {
		return db
	}

	dsn := os.Getenv("MYSQL_DSN")
	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	return db
}

// Get Получение подключения к БД, если уже инициализирована - выдает его, если нет - инициализирует
func Get() *gorm.DB {
	if db == nil {
		Init()
	}
	return db
}
