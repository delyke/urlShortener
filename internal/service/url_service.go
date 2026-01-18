package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"github.com/delyke/urlShortener/internal/audit"
	"github.com/delyke/urlShortener/internal/config"
	"github.com/delyke/urlShortener/internal/model"
	"github.com/delyke/urlShortener/internal/repository"
	"log"
	"strings"
	"sync"
	"time"
)

type URLService struct {
	repo       repository.URLRepository
	cfg        *config.Config
	delURLCh   chan DeleteUserURLs
	delTimeout time.Duration
	batchInt   time.Duration
	wg         sync.WaitGroup
	batchMax   int
	stopCh     chan struct{}
	observers  []audit.Observer
	obsMu      sync.RWMutex
}

func NewURLService(repo repository.URLRepository, config *config.Config, delTimeout time.Duration, bufLen int) *URLService {
	if delTimeout <= 0 {
		delTimeout = time.Second * 5
	}
	return &URLService{
		repo:       repo,
		cfg:        config,
		delTimeout: delTimeout,
		delURLCh:   make(chan DeleteUserURLs, bufLen),
		batchMax:   1000,
		batchInt:   time.Second * 5,
		stopCh:     make(chan struct{}),
	}
}

var ErrNotFound = errors.New("url not found")
var ErrCanNotCreateURL = errors.New("url cannot be created")

type DeleteUserURLs struct {
	UserID       int64
	ShortenedURL string
}

func (s *URLService) StartDeleter() {
	s.wg.Add(1)
	go s.deleteLoop()
}

func (s *URLService) StopDeleter() {
	close(s.stopCh)
	s.wg.Wait()
}

func (s *URLService) RegisterAuditObserver(observer audit.Observer) {
	if observer == nil {
		return
	}
	s.obsMu.Lock()
	defer s.obsMu.Unlock()
	s.observers = append(s.observers, observer)
}

func (s *URLService) NotifyAudit(ctx context.Context, event audit.Event) error {
	s.obsMu.RLock()
	observers := append([]audit.Observer(nil), s.observers...)
	s.obsMu.RUnlock()

	var errs []error
	for _, observer := range observers {
		if err := observer.OnEvent(ctx, event); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func fanIn[T any](chs ...<-chan T) <-chan T {
	out := make(chan T, len(chs))
	var wg sync.WaitGroup
	wg.Add(len(chs))
	for _, ch := range chs {
		ch := ch
		go func(c <-chan T) {
			defer wg.Done()
			for v := range c {
				out <- v
			}
		}(ch)
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func (s *URLService) deleteLoop() {
	defer s.wg.Done()

	merged := fanIn[DeleteUserURLs](s.delURLCh)

	ticker := time.NewTicker(s.batchInt)
	defer ticker.Stop()

	buckets := make(map[int64][]string)

	flushUser := func(uid int64) {
		urls := buckets[uid]
		if len(urls) == 0 {
			return
		}
		if err := s.repo.DeleteURLsByUser(uid, urls); err != nil {
			log.Printf("urls batch delete by user (%d) err: %v", uid, err)
		}
		delete(buckets, uid)
	}

	flushAll := func() {
		for uid := range buckets {
			flushUser(uid)
		}
	}

	for {
		select {
		case cmd, ok := <-merged:
			if !ok {
				flushAll()
				return
			}
			buckets[cmd.UserID] = append(buckets[cmd.UserID], cmd.ShortenedURL)
			if s.batchMax > 0 && len(buckets[cmd.UserID]) >= s.batchMax {
				flushUser(cmd.UserID)
			}
		case <-ticker.C:
			flushAll()
		case <-s.stopCh:
			flushAll()
			return
		}
	}
}

func (s *URLService) EnqueueDelete(userID int64, URLs []string) error {
	if len(URLs) == 0 {
		return errors.New("urls is empty")
	}

	for _, URL := range URLs {
		if URL == "" {
			continue
		}
		chData := DeleteUserURLs{UserID: userID, ShortenedURL: URL}
		select {
		case s.delURLCh <- chData:
		case <-time.After(s.delTimeout):
			return errors.New("enqueue timeout: queue is full")
		}
	}
	return nil
}

func (s *URLService) CreateUser() (int64, error) {
	return s.repo.CreateUser()
}

func (s *URLService) GetURLsByUser(userID int64) (*[]model.URL, error) {
	var urls *[]model.URL
	urls, err := s.repo.GetURLsByUserID(userID)
	if err != nil {
		return nil, err
	}
	return urls, nil
}

func (s *URLService) GetFreeShortURL() (string, error) {
	var shortenURL string
	for i := 0; i < 3; i++ {
		shortenURL = generateShortenURL()
		_, _, err := s.GetOriginalURL(shortenURL)
		if err == nil {
			shortenURL = ""
			continue
		} else {
			break
		}
	}
	if shortenURL == "" {
		return "", ErrCanNotCreateURL
	}
	return shortenURL, nil
}

func (s *URLService) ShortenURL(originalURL string, userID int64) (string, error) {
	shortenURL, err := s.GetFreeShortURL()
	if err != nil {
		return "", err
	}
	shortenURL, err = s.repo.Save(originalURL, shortenURL, userID)
	if err != nil {
		return "", err
	}
	return shortenURL, nil
}

func (s *URLService) GetOriginalURL(shortenURL string) (string, *bool, error) {
	url, isDeleted, err := s.repo.GetOriginalLink(shortenURL)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return "", nil, ErrNotFound
		} else {
			return "", nil, err
		}
	}
	return url, isDeleted, nil
}

func (s *URLService) PingDatabase() error {
	return s.repo.Ping()
}

func (s *URLService) ShortenBatch(items []model.BatchRequestItem, userID int64) ([]model.BatchResponseItem, error) {
	records := make([]model.URL, 0, len(items))
	responses := make([]model.BatchResponseItem, 0, len(items))

	for _, item := range items {
		short, err := s.GetFreeShortURL()
		if err != nil {
			return nil, err
		}
		records = append(records, model.URL{
			OriginalURL: item.OriginalURL,
			ShortURL:    short,
			UserID:      userID,
		})
		responses = append(responses, model.BatchResponseItem{
			CorrelationID: item.CorrelationID,
			ShortURL:      s.cfg.BaseAddr + "/" + short,
		})
	}

	if err := s.repo.SaveBatch(records); err != nil {
		return nil, err
	}

	return responses, nil
}

func generateShortenURL() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return strings.TrimRight(base64.URLEncoding.EncodeToString(b), "=")
}
