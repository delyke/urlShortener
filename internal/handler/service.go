package handler

import "github.com/delyke/urlShortener/internal/model"

type ShortenURLService interface {
	ShortenURL(originalURL string, userID int64) (string, error)
	GetOriginalURL(shortenURL string) (string, error)
	ShortenBatch(items []model.BatchRequestItem, userID int64) ([]model.BatchResponseItem, error)
	PingDatabase() error
	GetFreeShortURL() (string, error)
	GetURLsByUser(userID int64) (*[]model.URL, error)
}
