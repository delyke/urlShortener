package handler

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/delyke/urlShortener/internal/app/appctx"
	"github.com/delyke/urlShortener/internal/config"
	"github.com/delyke/urlShortener/internal/logger"
	"github.com/delyke/urlShortener/internal/mocks"
	"github.com/delyke/urlShortener/internal/repository"
	"github.com/delyke/urlShortener/internal/service"
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
			request = request.WithContext(appctx.SetUserID(request.Context(), 1))
			w := httptest.NewRecorder()
			cfg := &config.Config{
				RunAddr:  ":8080",
				BaseAddr: "http://localhost:8080",
				LogLevel: "debug",
			}

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockURLRepository(ctrl)
			svc := service.NewURLService(repo, cfg, 5*time.Second, 100)
			l, err := logger.Initialize(cfg.LogLevel)
			if err != nil {
				log.Fatal("Failed to initialize logger:", err)
			}
			h := NewHandler(svc, cfg, l)

			if tt.name == "Positive Test" {
				repo.EXPECT().
					GetOriginalLink(gomock.Any()).
					Return("", nil, repository.ErrRecordNotFound)

				repo.EXPECT().
					Save("https://vk.com", gomock.Any(), gomock.Any()).
					Return("abc123", nil)
			}

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
