package repository

import (
	"github.com/delyke/urlShortener/internal/model"
	"gorm.io/gorm"
)

type URLRepository interface {
	Save(originalUrl string, shortedUrl string) error
	GetOriginalLink(shortedUrl string) (string, bool)
}

type MySQLRepository struct {
	db *gorm.DB
}

func NewMySQLRepository(db *gorm.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

func (r *MySQLRepository) Save(originalUrl string, shortedUrl string) error {
	url := model.URL{OriginalUrl: originalUrl, ShortedUrl: shortedUrl}
	return r.db.Create(&url).Error
}

func (r *MySQLRepository) GetOriginalLink(shortedUrl string) (string, bool) {
	var url model.URL
	err := r.db.Where("shorted_url = ?", shortedUrl).First(&url).Error
	if err != nil {
		return "", false
	}
	return url.OriginalUrl, true
}
