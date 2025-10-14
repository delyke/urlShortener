package handler

import (
	"encoding/json"
	"github.com/delyke/urlShortener/internal/app/appctx"
	"github.com/delyke/urlShortener/internal/config"
	"github.com/delyke/urlShortener/internal/mocks"
	"github.com/delyke/urlShortener/internal/model"
	"github.com/delyke/urlShortener/internal/repository"
	"github.com/delyke/urlShortener/internal/service"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHandler_HandleAPIUserURLs_OK(t *testing.T) {
	cfg := &config.Config{
		RunAddr:  ":8080",
		BaseAddr: "http://localhost:8080",
		LogLevel: "debug",
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockURLRepository(ctrl)
	svc := service.NewURLService(repo, cfg, 5*time.Second, 100)
	h := NewHandler(svc, cfg)

	const uid int64 = 1

	urls := &[]model.URL{
		{
			UUID:        "svn3-adwad2-awfasf",
			ShortURL:    "abc123",
			OriginalURL: "https://vk.com",
			UserID:      uid,
		},
	}
	repo.EXPECT().
		GetURLsByUserID(uid).
		Return(urls, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	req = req.WithContext(appctx.SetUserID(req.Context(), uid))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	hh := http.HandlerFunc(h.HandleAPIUserURLs)
	hh.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Contains(t, res.Header.Get("Content-Type"), "application/json")

	var body []respUserURLs
	_ = json.NewDecoder(res.Body).Decode(&body)
	assert.Len(t, body, 1)
	assert.Equal(t, "https://vk.com", body[0].OriginalURL)
	assert.Equal(t, "http://localhost:8080/abc123", body[0].ShortURL)
}

func TestHandler_HandleAPIUserURLs_NoContent(t *testing.T) {
	cfg := &config.Config{
		RunAddr:  ":8080",
		BaseAddr: "http://localhost:8080",
		LogLevel: "debug",
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockURLRepository(ctrl)
	svc := service.NewURLService(repo, cfg, 5*time.Second, 100)
	h := NewHandler(svc, cfg)

	const uid int64 = 1

	repo.EXPECT().
		GetURLsByUserID(uid).
		Return(nil, repository.ErrRecordNotFound)

	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	req = req.WithContext(appctx.SetUserID(req.Context(), uid))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	hh := http.HandlerFunc(h.HandleAPIUserURLs)
	hh.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusNoContent, res.StatusCode)
}
