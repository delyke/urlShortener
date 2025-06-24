package repository

import (
	"github.com/delyke/urlShortener/internal/model"
	"gorm.io/gorm"
)

type MySQLRepository struct {
	db *gorm.DB
}

func NewMySQLRepository(db *gorm.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

func (r *MySQLRepository) Save(originalURL string, shortedURL string) error {
	url := model.URL{OriginalURL: originalURL, ShortedURL: shortedURL}
	return r.db.Create(&url).Error
}

func (r *MySQLRepository) GetOriginalLink(shortedURL string) (string, bool) {
	var url model.URL
	err := r.db.Where("shorted_url = ?", shortedURL).First(&url).Error
	if err != nil {
		return "", false
	}
	return url.OriginalURL, true
}
