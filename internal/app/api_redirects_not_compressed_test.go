package app

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"github.com/delyke/urlShortener/internal/config"
	"github.com/delyke/urlShortener/internal/handler"
	"github.com/delyke/urlShortener/internal/logger"
	"github.com/delyke/urlShortener/internal/mocks"
	"github.com/delyke/urlShortener/internal/repository"
	"github.com/delyke/urlShortener/internal/service"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGzipExpand(t *testing.T) {
	cfg := &config.Config{
		RunAddr:         ":8080",
		BaseAddr:        "http://localhost:8080",
		LogLevel:        "debug",
		FileStoragePath: "/storage.json",
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockURLRepository(ctrl)

	svc := service.NewURLService(repo, cfg)
	h := handler.NewHandler(svc, cfg)
	l, err := logger.Initialize(cfg.LogLevel)
	require.NoError(t, err)
	router := NewRouter(h, l)

	originalURL := "https://vk.com"
	requestBody := map[string]string{"url": originalURL}
	jsonBody, _ := json.Marshal(requestBody)

	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, err = zw.Write(jsonBody)
	require.NoError(t, err)
	require.NoError(t, zw.Close())

	repo.EXPECT().
		Save(originalURL, gomock.Any()).
		Return(nil)

	repo.EXPECT().
		GetOriginalLink(gomock.Any()).
		Return("", repository.ErrRecordNotFound)

	req := httptest.NewRequest(http.MethodPost, "/api/shorten", &buf)

	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")

	w := httptest.NewRecorder()

	repo.EXPECT().
		GetOriginalLink(gomock.Any()).
		Return(originalURL, nil)

	router.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusCreated, res.StatusCode)
	require.Equal(t, "gzip", res.Header.Get("Content-Encoding"))

	zr, err := gzip.NewReader(res.Body)
	require.NoError(t, err)
	bodyBytes, err := io.ReadAll(zr)
	require.NoError(t, err)
	require.NoError(t, zr.Close())

	type shortenResponse struct {
		Result string `json:"result"`
	}

	var resp shortenResponse
	err = json.Unmarshal(bodyBytes, &resp)
	require.NoError(t, err)
	require.NotEmpty(t, resp.Result)

	require.True(t, strings.HasPrefix(resp.Result, cfg.BaseAddr), "unexpected result URL prefix")
	shortKey := strings.TrimPrefix(resp.Result, cfg.BaseAddr+"/")
	require.NotEmpty(t, shortKey)

	log.Println("SHORT KEY:", shortKey)
	getReq := httptest.NewRequest(http.MethodGet, "/"+shortKey, nil)
	getReq.Header.Set("Accept-Encoding", "gzip")
	wGet := httptest.NewRecorder()

	router.ServeHTTP(wGet, getReq)
	getRes := wGet.Result()
	defer getRes.Body.Close()

	require.Equal(t, http.StatusTemporaryRedirect, getRes.StatusCode)
	location := getRes.Header.Get("Location")
	require.Equal(t, originalURL, location)
}
