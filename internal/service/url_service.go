package service

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"github.com/delyke/urlShortener/internal/config"
	"github.com/delyke/urlShortener/internal/model"
	"github.com/delyke/urlShortener/internal/repository"
	"log"
	"strings"
)

type URLService struct {
	repo repository.URLRepository
	cfg  *config.Config
}

func NewURLService(repo repository.URLRepository, config *config.Config) *URLService {
	return &URLService{repo: repo, cfg: config}
}

var ErrNotFound = errors.New("url not found")
var ErrCanNotCreateURL = errors.New("url cannot be created")

func (s *URLService) GetFreeShortURL() (string, error) {
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
	return shortenURL, nil
}

func (s *URLService) ShortenURL(originalURL string) (string, error) {
	shortenURL, err := s.GetFreeShortURL()
	if err != nil {
		return "", err
	}
	shortenURL, err = s.repo.Save(originalURL, shortenURL)
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

func (s *URLService) ShortenBatch(items []model.BatchRequestItem) ([]model.BatchResponseItem, error) {
	var records []model.URL
	var responses []model.BatchResponseItem

	for _, item := range items {
		short, err := s.GetFreeShortURL()
		if err != nil {
			return nil, err
		}
		records = append(records, model.URL{
			OriginalURL: item.OriginalURL,
			ShortURL:    short,
		})
		responses = append(responses, model.BatchResponseItem{
			CorrelationID: item.CorrelationID,
			ShortURL:      s.cfg.BaseAddr + "/" + short,
		})
	}

	if err := s.repo.SaveBatch(records); err != nil {
		return nil, err
	}

	return responses, nil
}

func generateShortenURL() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return strings.TrimRight(base64.URLEncoding.EncodeToString(b), "=")
}
