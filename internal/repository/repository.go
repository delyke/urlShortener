package repository

import "github.com/delyke/urlShortener/internal/model"

type URLRepository interface {
	Save(originalURL string, shortedURL string) (string, error)
	GetOriginalLink(shortedURL string) (string, error)
	SaveBatch(records []model.URL) error
	Ping() error
	GetShortURLByOriginal(originalURL string) (string, error)
}
