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

func TestHandler_HandleApiShorten(t *testing.T) {
	type want struct {
		code        int
		contentType string
		body        []byte
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
			name:        "Api Post Shorten Wrong Content Type",
			body:        []byte(`{"url":"http://www.google.com"}`),
			contentType: "text/plain",
			method:      http.MethodPost,
			request:     "/api/shorten",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "application/json",
			},
		},
		{
			name:        "Api Post Empty Body",
			body:        nil,
			contentType: "application/json",
			method:      http.MethodPost,
			request:     "/api/shorten",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "application/json",
			},
		},
		{
			name:        "Api Post Wrong JSON Body",
			body:        []byte(`{"location":"http://www.google.com"}`),
			contentType: "application/json",
			method:      http.MethodPost,
			request:     "/api/shorten",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "application/json",
			},
		},
		{
			name:        "Api Post Shorten Success",
			body:        []byte(`{"url":"http://www.google.com"}`),
			contentType: "application/json",
			method:      http.MethodPost,
			request:     "/api/shorten",
			want: want{
				code:        http.StatusCreated,
				contentType: "application/json",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.request, bytes.NewReader(tt.body))
			request.Header.Set("Content-Type", tt.contentType)
			w := httptest.NewRecorder()
			cfg := &config.Config{
				RunAddr:         ":8080",
				BaseAddr:        "http://localhost:8080",
				LogLevel:        "debug",
				FileStoragePath: "/storage.json",
			}
			repo := repository.NewFileRepository(cfg.FileStoragePath)
			svc := service.NewURLService(repo)

			h := NewHandler(svc, cfg)

			hh := http.HandlerFunc(h.HandleAPIShorten)
			hh(w, request)

			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, result.StatusCode, tt.want.code)
			assert.Equal(t, result.Header.Get("Content-Type"), tt.want.contentType)
			if tt.want.code == http.StatusCreated {
				assert.NotEmpty(t, w.Body.String())
			}
		})
	}
}
