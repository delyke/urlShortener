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

func (repo *LocalRepository) Save(originalUrl string, shortedUrl string) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	repo.data[shortedUrl] = originalUrl
	return nil
}

func (repo *LocalRepository) GetOriginalLink(shortedUrl string) (string, bool) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	originalUrl, ok := repo.data[shortedUrl]
	return originalUrl, ok
}
