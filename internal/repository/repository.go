package repository

import "github.com/delyke/urlShortener/internal/model"

type URLRepository interface {
	Save(originalURL string, shortedURL string, userID int64) (string, error)
	GetOriginalLink(shortedURL string) (string, *bool, error)
	SaveBatch(records []model.URL) error
	Ping() error
	GetShortURLByOriginal(originalURL string) (string, error)
	CreateUser() (int64, error)
	GetURLsByUserID(userID int64) (*[]model.URL, error)
	DeleteURLsByUser(userID int64, URLs []string) error
}
