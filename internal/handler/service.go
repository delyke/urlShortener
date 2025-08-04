package handler

import "github.com/delyke/urlShortener/internal/model"

type ShortenURLService interface {
	ShortenURL(originalURL string) (string, error)
	GetOriginalURL(shortenURL string) (string, error)
	ShortenBatch(items []model.BatchRequestItem) ([]model.BatchResponseItem, error)
	PingDatabase() error
}
