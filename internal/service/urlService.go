package service

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"github.com/delyke/urlShortener/internal/repository"
	"strings"
)

type URLService struct {
	repo repository.URLRepository
}

func NewURLService(repo repository.URLRepository) *URLService {
	return &URLService{repo: repo}
}

var errNotFound = errors.New("url not found")

func (s *URLService) ShortenURL(originalUrl string) (string, error) {
beginShortUrl:
	shortenURL := generateShortenURL()
	_, errExist := s.GetOriginalURL(shortenURL)

	if errExist == nil {
		goto beginShortUrl
	}

	err := s.repo.Save(originalUrl, shortenURL)
	if err != nil {
		return "", err
	}
	return shortenURL, nil
}

func (s *URLService) GetOriginalURL(shortenURL string) (string, error) {
	url, ok := s.repo.GetOriginalLink(shortenURL)
	if !ok {
		return "", errNotFound
	}
	return url, nil
}

func generateShortenURL() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return strings.TrimRight(base64.URLEncoding.EncodeToString(b), "=")
}
