package service

import (
	"github.com/delyke/urlShortener/internal/config"
	"github.com/delyke/urlShortener/internal/repository"
	"testing"
	"time"
)

func BenchmarkURLService_GetOriginalURL(b *testing.B) {
	repo, err := repository.NewLocalRepository()
	if err != nil {
		b.Fatalf("failed to init local database: %v", err)
	}
	cfg := &config.Config{
		BaseAddr: "http://127.0.0.1:8080",
	}
	svc := NewURLService(repo, cfg, time.Second, 100)
	short, err := svc.ShortenURL("http://example.com", 1)
	if err != nil {
		b.Fatalf("failed to shorten url: %v", err)
	}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, _, err := svc.GetOriginalURL(short); err != nil {
			b.Fatalf("failed to shorten url: %v", err)
		}
	}
}

/*
goos: darwin
goarch: amd64
pkg: github.com/delyke/urlShortener/internal/service
cpu: VirtualApple @ 2.50GHz
BenchmarkURLService_GetOriginalURL
BenchmarkURLService_GetOriginalURL-11    	76152399	        15.09 ns/op	       1 B/op	       1 allocs/op
PASS
*/
