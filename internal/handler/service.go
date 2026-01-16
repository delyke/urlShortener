package handler

import (
	"context"
	"github.com/delyke/urlShortener/internal/audit"
	"github.com/delyke/urlShortener/internal/model"
)

type ShortenURLService interface {
	ShortenURL(originalURL string, userID int64) (string, error)
	GetOriginalURL(shortenURL string) (string, *bool, error)
	ShortenBatch(items []model.BatchRequestItem, userID int64) ([]model.BatchResponseItem, error)
	PingDatabase() error
	GetFreeShortURL() (string, error)
	GetURLsByUser(userID int64) (*[]model.URL, error)
	EnqueueDelete(userID int64, URLs []string) error
	NotifyAudit(ctx context.Context, event audit.Event) error
}
