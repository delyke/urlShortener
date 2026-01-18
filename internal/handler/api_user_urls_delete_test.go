package handler

import (
	"bytes"
	"encoding/json"
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
	"github.com/delyke/urlShortener/internal/service"
)

func newTestSvc(t *testing.T, cfg *config.Config, delTimeout time.Duration, bufLen int) *service.URLService {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	repo := mocks.NewMockURLRepository(ctrl)
	return service.NewURLService(repo, cfg, delTimeout, bufLen)
}

func TestHandleAPIUserURLsDelete_OK(t *testing.T) {
	cfg := &config.Config{BaseAddr: "http://localhost:8080"}
	svc := newTestSvc(t, cfg, 200*time.Millisecond, 10) // буфер есть → не блокируемся
	l, err := logger.Initialize(cfg.LogLevel)
	if err != nil {
		t.Errorf("Failed to initialize logger: %v", err)
		return
	}
	h := NewHandler(svc, cfg, l)

	body, _ := json.Marshal([]string{"a1", "b2", "c3"})
	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(appctx.SetUserID(req.Context(), 42))

	w := httptest.NewRecorder()
	http.HandlerFunc(h.HandleAPIUserURLsDelete).ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusAccepted, res.StatusCode)
	assert.Contains(t, res.Header.Get("Content-Type"), "application/json")
}

func TestHandleAPIUserURLsDelete_BadContentType(t *testing.T) {
	cfg := &config.Config{}
	svc := newTestSvc(t, cfg, 200*time.Millisecond, 10)
	l, err := logger.Initialize(cfg.LogLevel)
	if err != nil {
		t.Errorf("Failed to initialize logger: %v", err)
		return
	}
	h := NewHandler(svc, cfg, l)

	body, _ := json.Marshal([]string{"a1"})
	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewReader(body))
	req.Header.Set("Content-Type", "text/plain")
	req = req.WithContext(appctx.SetUserID(req.Context(), 1))

	w := httptest.NewRecorder()
	http.HandlerFunc(h.HandleAPIUserURLsDelete).ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(res.Body)
	assert.Contains(t, buf.String(), "Content-Type must be application/json")
}

func TestHandleAPIUserURLsDelete_Unauthorized_NoUser(t *testing.T) {
	cfg := &config.Config{}
	svc := newTestSvc(t, cfg, 200*time.Millisecond, 10)
	l, err := logger.Initialize(cfg.LogLevel)
	if err != nil {
		t.Errorf("Failed to initialize logger: %v", err)
		return
	}
	h := NewHandler(svc, cfg, l)

	body, _ := json.Marshal([]string{"a1"})
	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	http.HandlerFunc(h.HandleAPIUserURLsDelete).ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

func TestHandleAPIUserURLsDelete_EmptyIDs(t *testing.T) {
	cfg := &config.Config{}
	svc := newTestSvc(t, cfg, 200*time.Millisecond, 10)
	l, err := logger.Initialize(cfg.LogLevel)
	if err != nil {
		t.Errorf("Failed to initialize logger: %v", err)
		return
	}
	h := NewHandler(svc, cfg, l)

	body := []byte(`[]`)
	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(appctx.SetUserID(req.Context(), 7))

	w := httptest.NewRecorder()
	http.HandlerFunc(h.HandleAPIUserURLsDelete).ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandleAPIUserURLsDelete_EnqueueTimeout(t *testing.T) {
	cfg := &config.Config{}
	svc := newTestSvc(t, cfg, 10*time.Millisecond, 0)
	l, err := logger.Initialize(cfg.LogLevel)
	if err != nil {
		t.Errorf("Failed to initialize logger: %v", err)
		return
	}
	h := NewHandler(svc, cfg, l)

	body, _ := json.Marshal([]string{"x"})
	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(appctx.SetUserID(req.Context(), 1))

	w := httptest.NewRecorder()
	http.HandlerFunc(h.HandleAPIUserURLsDelete).ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}
