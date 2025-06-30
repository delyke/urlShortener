package service

type ShortenURLService interface {
	ShortenURL(originalURL string) (string, error)
	GetOriginalURL(shortenURL string) (string, error)
}
