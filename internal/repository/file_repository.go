package repository

import (
	"encoding/json"
	"fmt"
	"github.com/delyke/urlShortener/internal/model"
	"io"
	"log"
	"os"
)

type FileRepository struct {
	filename string
	urls     []model.URL
}

func NewFileRepository(filename string) (*FileRepository, error) {
	var urls []model.URL
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
	}, nil
}

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

func (p *Producer) Close() error {
	return p.file.Close()
}

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

func (c *Consumer) Close() error {
	return c.file.Close()
}

func (repo *FileRepository) Save(originalURL string, shortedURL string) error {
	producer, err := newProducer(repo.filename)
	if err != nil {
		return err
	}

	UUID, err := repo.generateUUID()
	if err != nil {
		return err
	}

	url := model.URL{
		UUID:        UUID,
		OriginalURL: originalURL,
		ShortURL:    shortedURL,
	}
	repo.urls = append(repo.urls, url)
	producer.encoder.SetIndent("", "\t")
	return producer.encoder.Encode(repo.urls)
}

func (repo *FileRepository) GetOriginalLink(shortedURL string) (string, error) {
	for _, url := range repo.urls {
		if url.ShortURL == shortedURL {
			return url.OriginalURL, nil
		}
	}
	return "", ErrRecordNotFound
}

func (repo *FileRepository) generateUUID() (string, error) {
	return fmt.Sprintf("%d", len(repo.urls)+1), nil
}

func (repo *FileRepository) Ping() error {
	return nil
}

func (repo *FileRepository) SaveBatch(records []model.URL) error {
	for _, record := range records {
		err := repo.Save(record.OriginalURL, record.ShortURL)
		if err != nil {
			return err
		}
	}
	return nil
}
