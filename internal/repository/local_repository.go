package repository

import (
	"errors"
	"fmt"
	"sync"
)

type LocalRepository struct {
	data map[string]string
	mu   *sync.Mutex
}

func NewLocalRepository() *LocalRepository {
	return &LocalRepository{
		data: make(map[string]string),
		mu:   &sync.Mutex{},
	}
}

func (repo *LocalRepository) Save(originalURL string, shortedURL string) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	repo.data[shortedURL] = originalURL
	return nil
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
