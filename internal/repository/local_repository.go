package repository

import (
	"errors"
	"log"
	"sync"
	"time"

	"github.com/delyke/urlShortener/internal/model"
)

// LocalRepository stores data in memory for development/testing
type LocalRepository struct {
	data struct {
		urls  []model.URL
		users []model.User
	}
	mu *sync.Mutex
}

// NewLocalRepository creates an in-memory repository instance.
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

// Save stores a new URL mapping in memory.
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
		IsDeleted:   false,
	}
	repo.data.urls = append(repo.data.urls, newURL)
	return shortedURL, nil
}

// ErrRecordNotFound is returned when a record cannot be found in storage.
var ErrRecordNotFound = errors.New("record not found")

// DeleteURLsByUser marks URLs as deleted for specified user.
func (repo *LocalRepository) DeleteURLsByUser(userID int64, URLs []string) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	if len(URLs) == 0 {
		return nil
	}

	toDelete := make(map[string]struct{}, len(URLs))
	for _, u := range URLs {
		if u == "" {
			continue
		}
		toDelete[u] = struct{}{}
	}

	var touched int
	for i := range repo.data.urls {
		u := &repo.data.urls[i]
		if u.UserID == userID {
			u.IsDeleted = true
			touched++
		}
	}

	log.Printf("[soft delete][file] user=%d, updated=%d", userID, touched)
	return nil
}

// GetOriginalLink - returns the original URL and deletion flag for a shortened URL.
func (repo *LocalRepository) GetOriginalLink(shortedURL string) (string, *bool, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	originalURL := ""
	var isDeleted bool
	for _, url := range repo.data.urls {
		if url.ShortURL == shortedURL {
			originalURL = url.OriginalURL
			isDeleted = url.IsDeleted
			break
		}
	}
	if originalURL == "" {
		return "", nil, ErrRecordNotFound
	}
	return originalURL, &isDeleted, nil
}

// Ping checks repository availability
func (repo *LocalRepository) Ping() error {
	return nil
}

// SaveBatch stores a batch of URL records.
func (repo *LocalRepository) SaveBatch(records []model.URL) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	repo.data.urls = append(repo.data.urls, records...)
	return nil
}

// GetShortURLByOriginal - finds a short URL by its original URL.
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

// CreateUser - creates a new user record and returns its ID.
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

// GetURLsByUserID returns all URLs for the specified user.
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
