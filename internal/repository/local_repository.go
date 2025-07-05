package repository

import (
	"errors"
	"sync"
)

type LocalRepository struct {
	data map[string]string
	mu   sync.RWMutex
}

func NewLocalRepository() *LocalRepository {
	return &LocalRepository{
		data: make(map[string]string),
	}
}

func (repo *LocalRepository) Save(originalURL string, shortedURL string) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	repo.data[shortedURL] = originalURL
	return nil
}

var ErrorRecordNotFound = errors.New("record not found")

func (repo *LocalRepository) GetOriginalLink(shortedURL string) (string, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	originalURL, isSuccess := repo.data[shortedURL]
	if !isSuccess {
		return "", ErrorRecordNotFound
	}
	return originalURL, nil
}
