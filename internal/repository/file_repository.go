package repository

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/delyke/urlShortener/internal/model"
)

// FileRepository stores URL and user data in a JSON file.
type FileRepository struct {
	filename string
	urls     []model.URL
	users    []model.User
}

// NewFileRepository loads data from a JSON file and returns a repository.
func NewFileRepository(filename string) (*FileRepository, error) {
	var urls []model.URL
	var users []model.User
	consumer, err := newConsumer(filename)
	if err != nil {
		return nil, err
	}
	if err := consumer.decoder.Decode(&urls); err != nil {
		if err != io.EOF {
			log.Printf("Error reading from file: %v", err)
			return nil, err
		}
	}
	defer consumer.file.Close()
	return &FileRepository{
		filename: filename,
		urls:     urls,
		users:    users,
	}, nil
}

// Producer writes JSON records to a file.
type Producer struct {
	file    *os.File
	encoder *json.Encoder
}

func newProducer(filename string) (*Producer, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	return &Producer{file: file, encoder: json.NewEncoder(file)}, nil
}

// Close releases the underlying file descriptor
func (p *Producer) Close() error {
	return p.file.Close()
}

// Consumer - reads JSON records from a file.
type Consumer struct {
	file    *os.File
	decoder *json.Decoder
}

func newConsumer(filename string) (*Consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	return &Consumer{file: file, decoder: json.NewDecoder(file)}, nil
}

// Close releases the underlying file descriptor.
func (c *Consumer) Close() error {
	return c.file.Close()
}

// DeleteURLsByUser - marks URLs as deleted for the specified user.
func (repo *FileRepository) DeleteURLsByUser(userID int64, URLs []string) error {
	if len(URLs) == 0 {
		return nil
	}
	producer, err := newProducer(repo.filename)
	if err != nil {
		return err
	}

	toDelete := make(map[string]struct{}, len(URLs))
	for _, u := range URLs {
		if u == "" {
			continue
		}
		toDelete[u] = struct{}{}
	}

	var touched int
	for i := range repo.urls {
		u := &repo.urls[i]
		if u.UserID == userID {
			u.IsDeleted = true
			touched++
		}
	}

	if touched == 0 {
		return nil
	}

	producer.encoder.SetIndent("", "\t")
	if err := producer.encoder.Encode(repo.urls); err != nil {
		return err
	}
	log.Printf("[soft delete][file] user=%d, updated=%d", userID, touched)
	return nil
}

// Save stores a new URL mapping and persists it to disk.
func (repo *FileRepository) Save(originalURL string, shortedURL string, userID int64) (string, error) {
	for _, u := range repo.urls {
		if u.OriginalURL == originalURL {
			return u.ShortURL, NewConflictError(u.ShortURL)
		}
	}

	producer, err := newProducer(repo.filename)
	if err != nil {
		return "", err
	}

	UUID, err := repo.generateUUID()
	if err != nil {
		return "", err
	}

	url := model.URL{
		UUID:        UUID,
		OriginalURL: originalURL,
		ShortURL:    shortedURL,
		UserID:      userID,
		IsDeleted:   false,
	}
	repo.urls = append(repo.urls, url)

	producer.encoder.SetIndent("", "\t")
	if err := producer.encoder.Encode(repo.urls); err != nil {
		return "", err
	}

	return shortedURL, nil
}

// GetOriginalLink returns the original URL and deletion flag for a shortened URL.
func (repo *FileRepository) GetOriginalLink(shortedURL string) (string, *bool, error) {
	for _, url := range repo.urls {
		if url.ShortURL == shortedURL {
			return url.OriginalURL, &url.IsDeleted, nil
		}
	}
	return "", nil, ErrRecordNotFound
}

func (repo *FileRepository) generateUUID() (string, error) {
	return fmt.Sprintf("%d", len(repo.urls)+1), nil
}

// Ping checks repository availability.
func (repo *FileRepository) Ping() error {
	return nil
}

// SaveBatch stores a batch of URL records.
func (repo *FileRepository) SaveBatch(records []model.URL) error {
	for _, record := range records {
		_, err := repo.Save(record.OriginalURL, record.ShortURL, record.UserID)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetShortURLByOriginal finds a short URL by its original URL.
func (repo *FileRepository) GetShortURLByOriginal(originalURL string) (string, error) {
	for _, url := range repo.urls {
		if url.OriginalURL == originalURL {
			return url.ShortURL, nil
		}
	}
	return "", ErrRecordNotFound
}

// CreateUser creates a new user record and returns its ID.
func (repo *FileRepository) CreateUser() (int64, error) {
	producer, err := newProducer(repo.filename)
	if err != nil {

		return 0, err
	}

	user := model.User{
		ID:        int64(len(repo.users) + 1),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.users = append(repo.users, user)

	producer.encoder.SetIndent("", "\t")
	if err := producer.encoder.Encode(repo.users); err != nil {
		return 0, err
	}

	return user.ID, nil
}

// GetURLsByUserID returns all URLs for the specified user.
func (repo *FileRepository) GetURLsByUserID(userID int64) (*[]model.URL, error) {
	var urls []model.URL
	for _, url := range repo.urls {
		if url.UserID == userID {
			urls = append(urls, url)
		}
	}
	return &urls, nil
}
