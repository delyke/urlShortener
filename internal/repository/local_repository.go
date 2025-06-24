package repository

import "sync"

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

func (repo *LocalRepository) GetOriginalLink(shortedURL string) (string, bool) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	originalURL, ok := repo.data[shortedURL]
	return originalURL, ok
}
