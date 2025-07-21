package handler

import (
	"bytes"
	"github.com/delyke/urlShortener/internal/config"
	"github.com/delyke/urlShortener/internal/repository"
	"github.com/delyke/urlShortener/internal/service"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_HandlePost(t *testing.T) {
	type want struct {
		code        int
		contentType string
	}
	tests := []struct {
		name        string
		body        []byte
		contentType string
		method      string
		request     string
		want        want
	}{
		{
			name:        "Positive Test",
			body:        []byte(`https://vk.com`),
			contentType: "text/plain",
			method:      "POST",
			request:     "/",
			want: want{
				code:        http.StatusCreated,
				contentType: "text/plain",
			},
		},
		{
			name:        "Test With Empty Body",
			body:        []byte(``),
			contentType: "text/plain",
			method:      "POST",
			request:     "/",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.request, bytes.NewReader(tt.body))
			w := httptest.NewRecorder()
			cfg := &config.Config{
				RunAddr:         ":8080",
				BaseAddr:        "http://localhost:8080",
				LogLevel:        "debug",
				FileStoragePath: "/storage.json",
			}

			repo, err := repository.NewFileRepository(cfg.FileStoragePath)
			if err != nil {
				t.Errorf("Failed to initialize repo: %v", err)
				return
			}
			svc := service.NewURLService(repo)
			h := NewHandler(svc, cfg)

			hh := http.HandlerFunc(h.HandlePost)
			hh(w, request)

			result := w.Result()
			result.Body.Close()

			assert.Equal(t, tt.want.code, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
			assert.NotEmpty(t, result.Body)
		})
	}
}
