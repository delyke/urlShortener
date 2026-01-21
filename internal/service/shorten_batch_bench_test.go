package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/delyke/urlShortener/internal/config"
	"github.com/delyke/urlShortener/internal/model"
	"github.com/delyke/urlShortener/internal/repository"
)

func BenchmarkURLService_ShortenBatch(b *testing.B) {
	repo, err := repository.NewLocalRepository()
	if err != nil {
		b.Fatalf("failed to create repository: %v", err)
	}
	cfg := &config.Config{
		BaseAddr: "http://localhost:8080",
	}
	svc := NewURLService(repo, cfg, time.Second, 100)
	items := make([]model.BatchRequestItem, 100)
	for i := range items {
		items[i] = model.BatchRequestItem{
			CorrelationID: fmt.Sprintf("id-%d", i),
			OriginalURL:   fmt.Sprintf("http://example.com/%d", i),
		}
	}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := svc.ShortenBatch(items, 1); err != nil {
			b.Fatalf("shorent batch failed: %v", err)
		}
	}
}

/*
goos: darwin
goarch: amd64
pkg: github.com/delyke/urlShortener/internal/service
cpu: VirtualApple @ 2.50GHz
BenchmarkURLService_ShortenBatch
BenchmarkURLService_ShortenBatch-11    	    1059	  11578880 ns/op	   66849 B/op	     416 allocs/op
PASS

Process finished with the exit code 0
*/
