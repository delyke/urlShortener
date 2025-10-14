package repository

import (
	"errors"
	"github.com/delyke/urlShortener/internal/model"
	"log"
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

func (repo *LocalRepository) GetOriginalLink(shortedURL string) (string, *bool, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	originalURL := ""
	var isDeleted bool
	for _, url := range repo.data.urls {
		if url.ShortURL == shortedURL {
			originalURL = url.OriginalURL
			isDeleted = true
			break
		}
	}
	if originalURL == "" {
		return "", nil, ErrRecordNotFound
	}
	return originalURL, &isDeleted, nil
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
