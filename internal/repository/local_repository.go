package repository

import (
	"errors"
	"github.com/delyke/urlShortener/internal/model"
	"sync"
	"time"
)

type LocalRepository struct {
	data struct {
		urls  []model.URL
		users []model.User
	}
	mu *sync.Mutex
}

func NewLocalRepository() (*LocalRepository, error) {
	return &LocalRepository{
		data: struct {
			urls  []model.URL
			users []model.User
		}{
			urls:  []model.URL{},
			users: []model.User{},
		},
		mu: &sync.Mutex{},
	}, nil
}

func (repo *LocalRepository) Save(originalURL string, shortedURL string, userID int64) (string, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	for _, url := range repo.data.urls {
		if url.OriginalURL == originalURL {
			return url.ShortURL, NewConflictError(url.ShortURL)
		}
	}

	newURL := model.URL{
		OriginalURL: originalURL,
		ShortURL:    shortedURL,
		UserID:      userID,
	}
	repo.data.urls = append(repo.data.urls, newURL)
	return shortedURL, nil
}

var ErrRecordNotFound = errors.New("record not found")

func (repo *LocalRepository) GetOriginalLink(shortedURL string) (string, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	originalURL := ""
	for _, url := range repo.data.urls {
		if url.ShortURL == shortedURL {
			originalURL = url.OriginalURL
			break
		}
	}
	if originalURL == "" {
		return "", ErrRecordNotFound
	}
	return originalURL, nil
}

func (repo *LocalRepository) Ping() error {
	return nil
}

func (repo *LocalRepository) SaveBatch(records []model.URL) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	repo.data.urls = append(repo.data.urls, records...)
	return nil
}

func (repo *LocalRepository) GetShortURLByOriginal(originalURL string) (string, error) {
	var shortURL string
	repo.mu.Lock()
	defer repo.mu.Unlock()
	for _, url := range repo.data.urls {
		if url.OriginalURL == originalURL {
			shortURL = url.ShortURL
			break
		}
	}
	if shortURL == "" {
		return "", ErrRecordNotFound
	}
	return shortURL, nil
}

func (repo *LocalRepository) CreateUser() (int64, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	user := model.User{
		ID:        int64(len(repo.data.users) + 1),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.data.users = append(repo.data.users, user)
	return user.ID, nil
}

func (repo *LocalRepository) GetURLsByUserID(userID int64) (*[]model.URL, error) {
	var urls []model.URL
	repo.mu.Lock()
	defer repo.mu.Unlock()
	for _, url := range repo.data.urls {
		if url.UserID == userID {
			urls = append(urls, url)
		}
	}
	return &urls, nil
}
