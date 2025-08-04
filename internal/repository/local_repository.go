package repository

import (
	"errors"
	"fmt"
	"github.com/delyke/urlShortener/internal/model"
	"sync"
)

type LocalRepository struct {
	data map[string]string
	mu   *sync.Mutex
}

func NewLocalRepository() (*LocalRepository, error) {
	return &LocalRepository{
		data: make(map[string]string),
		mu:   &sync.Mutex{},
	}, nil
}

func (repo *LocalRepository) Save(originalURL string, shortedURL string) (string, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	for short, orig := range repo.data {
		if orig == originalURL {
			return short, NewConflictError(short)
		}
	}

	repo.data[shortedURL] = originalURL
	return shortedURL, nil
}

var ErrRecordNotFound = errors.New("record not found")

func (repo *LocalRepository) GetOriginalLink(shortedURL string) (string, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	originalURL, isSuccess := repo.data[shortedURL]
	if !isSuccess {
		return "", fmt.Errorf("%w: %s", ErrRecordNotFound, shortedURL)
	}
	return originalURL, nil
}

func (repo *LocalRepository) Ping() error {
	return nil
}

func (repo *LocalRepository) SaveBatch(records []model.URL) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	for _, record := range records {
		repo.data[record.ShortURL] = record.OriginalURL
	}
	return nil
}

func (repo *LocalRepository) GetShortURLByOriginal(originalURL string) (string, error) {
	if short, exists := repo.data[originalURL]; exists {
		return short, nil
	}
	return "", ErrRecordNotFound
}
