package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/delyke/urlShortener/internal/app/appctx"
	"github.com/delyke/urlShortener/internal/audit"
	"github.com/delyke/urlShortener/internal/config"
	"github.com/delyke/urlShortener/internal/handler"
	"github.com/delyke/urlShortener/internal/logger"
	"github.com/delyke/urlShortener/internal/model"
	"github.com/go-chi/chi/v5"
	"net/http"
	"net/http/httptest"
	"strings"
)

type exampleService struct {
	baseAddr string
}

func (s exampleService) ShortenURL(originalURL string, userID int64) (string, error) {
	return "abc123", nil
}

func (s exampleService) GetOriginalURL(shortenURL string) (string, *bool, error) {
	deleted := false
	return "https://example.com", &deleted, nil
}

func (s exampleService) ShortenBatch(items []model.BatchRequestItem, userID int64) ([]model.BatchResponseItem, error) {
	responses := make([]model.BatchResponseItem, 0, len(items))
	for _, item := range items {
		responses = append(responses, model.BatchResponseItem{
			CorrelationID: item.CorrelationID,
			ShortURL:      s.baseAddr + "/" + item.CorrelationID,
		})
	}
	return responses, nil
}

func (s exampleService) PingDatabase() error {
	return nil
}

func (s exampleService) GetFreeShortURL() (string, error) {
	return "abc123", nil
}

func (s exampleService) GetURLsByUser(_ int64) (*[]model.URL, error) {
	urls := []model.URL{{
		ShortURL:    "abc123",
		OriginalURL: "https://example.com",
	}}
	return &urls, nil
}

func (s exampleService) EnqueueDelete(_ int64, _ []string) error {
	return nil
}

func (s exampleService) NotifyAudit(_ context.Context, _ audit.Event) error {
	return nil
}

func newExampleHandler() *handler.Handler {
	cfg := &config.Config{
		BaseAddr: "http://localhost:8080",
	}
	log, _ := logger.Initialize("error")
	return handler.NewHandler(exampleService{baseAddr: cfg.BaseAddr}, cfg, log)
}

func ExampleHandler_HandleAPIShorten() {
	h := newExampleHandler()
	payload := []byte(`{"url": "https://example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(appctx.SetUserID(req.Context(), int64(123)))
	rec := httptest.NewRecorder()
	h.HandleAPIShorten(rec, req)
	fmt.Printf("status=%d body=%s\n", rec.Code, strings.TrimSpace(rec.Body.String()))
	// Output:
	// status=201 body={"result":"http://localhost:8080/abc123"}
}

func ExampleHandler_HandlePost() {
	h := newExampleHandler()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com"))
	req = req.WithContext(appctx.SetUserID(req.Context(), int64(123)))
	rec := httptest.NewRecorder()
	h.HandlePost(rec, req)
	fmt.Printf("status=%d body=%s\n", rec.Code, strings.TrimSpace(rec.Body.String()))
	// Output:
	// status=201 body=http://localhost:8080/abc123
}

func ExampleHandler_HandleGet() {
	h := newExampleHandler()
	req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("shortenURL", "abc123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(appctx.SetUserID(req.Context(), int64(123)))
	rec := httptest.NewRecorder()

	h.HandleGet(rec, req)
	fmt.Printf("status=%d body=%s\n", rec.Code, strings.TrimSpace(rec.Body.String()))
	// Output:
	// status=307 body=<a href="https://example.com">Temporary Redirect</a>.
}

func ExampleHandler_HandleAPIShortenBatch() {
	h := newExampleHandler()
	items := []model.BatchRequestItem{{
		CorrelationID: "req-1",
		OriginalURL:   "https://example.com",
	}}
	payload, _ := json.Marshal(items)
	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(appctx.SetUserID(req.Context(), int64(123)))
	rec := httptest.NewRecorder()

	h.HandleAPIShortenBatch(rec, req)

	fmt.Printf("status=%d body=%s\n", rec.Code, strings.TrimSpace(rec.Body.String()))
	// Output:
	// status=201 body=[{"correlation_id":"req-1","short_url":"http://localhost:8080/req-1"}]
}

func ExampleHandler_HandleAPIUserURLs() {
	h := newExampleHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	req = req.WithContext(appctx.SetUserID(req.Context(), int64(123)))
	rec := httptest.NewRecorder()

	h.HandleAPIUserURLs(rec, req)
	fmt.Printf("status=%d body=%s\n", rec.Code, strings.TrimSpace(rec.Body.String()))
	// Output:
	// status=200 body=[{"short_url":"http://localhost:8080/abc123","original_url":"https://example.com"}]
}

func ExampleHandler_HandleAPIUserURLsDelete() {
	h := newExampleHandler()
	payload := []byte(`["abc123"]`)
	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(appctx.SetUserID(req.Context(), int64(123)))
	rec := httptest.NewRecorder()

	h.HandleAPIUserURLsDelete(rec, req)
	fmt.Printf("status=%d\n", rec.Code)
	// Output:
	// status=202
}

func ExampleHandler_HandlePing() {
	h := newExampleHandler()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	h.HandlePing(rec, req)
	fmt.Printf("status=%d\n", rec.Code)
	// Output:
	// status=200
}
