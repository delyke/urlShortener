package service

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/delyke/urlShortener/internal/repository"
	"strings"
)

type URLService struct {
	repo *repository.MySQLRepository
}

func NewURLService(repo *repository.MySQLRepository) *URLService {
	return &URLService{repo: repo}
}

var errNotFound = errors.New("url not found")

func (s *URLService) ShortenUrl(originalUrl string) (string, error) {
	fmt.Println(originalUrl)
beginShortUrl:
	shortenUrl := generateShortenUrl()
	_, errExist := s.GetOriginalUrl(shortenUrl)

	if errExist == nil {
		goto beginShortUrl
	}

	err := s.repo.Save(originalUrl, shortenUrl)
	if err != nil {
		return "", err
	}
	return shortenUrl, nil
}

func (s *URLService) GetOriginalUrl(shortenUrl string) (string, error) {
	url, ok := s.repo.GetOriginalLink(shortenUrl)
	if ok != true {
		return "", errNotFound
	}
	return url, nil
}

func generateShortenUrl() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return strings.TrimRight(base64.URLEncoding.EncodeToString(b), "=")
}
