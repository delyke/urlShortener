package service

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"github.com/delyke/urlShortener/internal/repository"
	"log"
	"strings"
)

type URLService struct {
	repo repository.URLRepository
}

func NewURLService(repo repository.URLRepository) *URLService {
	return &URLService{repo: repo}
}

var ErrNotFound = errors.New("url not found")
var ErrCanNotCreateURL = errors.New("url cannot be created")

func (s *URLService) ShortenURL(originalURL string) (string, error) {
	var shortenURL string
	for i := 0; i < 3; i++ {
		shortenURL = generateShortenURL()
		_, err := s.GetOriginalURL(shortenURL)
		if err == nil {
			shortenURL = ""
			continue
		} else {
			break
		}
	}
	if shortenURL == "" {
		return "", ErrCanNotCreateURL
	}

	err := s.repo.Save(originalURL, shortenURL)
	if err != nil {
		return "", err
	}
	return shortenURL, nil
}

func (s *URLService) GetOriginalURL(shortenURL string) (string, error) {
	log.Println("GetOriginalURL: ", shortenURL)
	url, err := s.repo.GetOriginalLink(shortenURL)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return "", ErrNotFound
		} else {
			return "", err
		}
	}
	return url, nil
}

func (s *URLService) PingDatabase() error {
	return s.repo.Ping()
}

func generateShortenURL() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return strings.TrimRight(base64.URLEncoding.EncodeToString(b), "=")
}
