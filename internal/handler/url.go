package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/delyke/urlShortener/internal/app/appctx"
	"github.com/delyke/urlShortener/internal/audit"
	"github.com/delyke/urlShortener/internal/config"
	"github.com/delyke/urlShortener/internal/logger"
	"github.com/delyke/urlShortener/internal/model"
	"github.com/delyke/urlShortener/internal/repository"
	"github.com/delyke/urlShortener/internal/service"
	"github.com/go-chi/chi/v5"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Handler struct {
	service ShortenURLService
	config  *config.Config
	l       *logger.Logger
}

func NewHandler(service ShortenURLService, cfg *config.Config, l *logger.Logger) *Handler {
	return &Handler{service: service, config: cfg, l: l}
}

func (h *Handler) sendAuditEvent(ctx context.Context, action string, userID int64, hasUser bool, originalURL string) {
	event := audit.Event{
		TS:     time.Now().Unix(),
		Action: action,
		URL:    originalURL,
	}
	if hasUser {
		event.UserID = strconv.FormatInt(userID, 10)
	}
	if err := h.service.NotifyAudit(ctx, event); err != nil {
		h.l.Errorf("failed to notify audit event: %v", err)
	}
}

func (h *Handler) HandlePost(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.l.Errorf("failed to read the request body: %v", err)
		return
	}

	originalURL := string(body)
	originalURL = strings.TrimSpace(originalURL)
	if originalURL == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID, ok := appctx.UserID(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	shortedURL, err := h.service.ShortenURL(originalURL, userID)
	if err != nil {
		if errors.Is(err, service.ErrCanNotCreateURL) {
			h.l.Errorf("URL shortening error:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		var conflict *repository.ConflictError
		if errors.As(err, &conflict) {
			existingURL, _ := url.JoinPath(h.config.BaseAddr, conflict.ShortURL)
			h.l.Errorf("URL shortening conflict:", existingURL)
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte(existingURL))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		h.l.Errorf("failed to shorten URL: %v", err)
		return
	}
	shortedURL, err = url.JoinPath(h.config.BaseAddr, shortedURL)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.l.Errorf("failed to join path: %v", err)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(shortedURL))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.l.Errorf("failed to write body: %v", err)
		return
	}
	h.sendAuditEvent(r.Context(), "shorten", userID, true, originalURL)
}

func (h *Handler) HandleAPIUserURLsDelete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Header.Get("Content-Type") != "application/json" {
		b, err := json.Marshal(ShortenURLErrorResponse{Error: "Content-Type must be application/json"})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.l.Debugf("invalid content type: %v", err)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		_, err = w.Write(b)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.l.Debugf("failed to write body: %v", err)
			return
		}
		return
	}

	userID, ok := appctx.UserID(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var ids []string
	if err := json.NewDecoder(r.Body).Decode(&ids); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		h.l.Debugf("failed to decode request body: %v", err)
		return
	}

	if len(ids) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := h.service.EnqueueDelete(userID, ids)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.l.Errorf("Enqueue Delete error: %v", err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

type respUserURLs struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func (h *Handler) HandleAPIUserURLs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := appctx.UserID(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	urls, err := h.service.GetURLsByUser(userID)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			w.WriteHeader(http.StatusNoContent)
			h.l.Debug(err)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		h.l.Errorf("failed get urls by user: %v", err)
		return
	}
	var resp []respUserURLs
	for _, u := range *urls {
		resp = append(resp, respUserURLs{ShortURL: fmt.Sprintf("%s/%s", h.config.BaseAddr, u.ShortURL), OriginalURL: u.OriginalURL})
	}
	b, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.l.Errorf("failed to marshal json: %v", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(b)
}

func (h *Handler) HandleGet(w http.ResponseWriter, r *http.Request) {
	shortedURL := chi.URLParam(r, "shortURL")
	originalURL, isDeleted, err := h.service.GetOriginalURL(shortedURL)
	if err == nil {
		if *isDeleted {
			w.WriteHeader(http.StatusGone)
			return
		}
		http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
		userID, ok := appctx.UserID(r.Context())
		h.sendAuditEvent(r.Context(), "follow", userID, ok, originalURL)
		return
	} else {
		if errors.Is(err, service.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			h.l.Errorf("error get original URL: %v", err)
			return
		}
	}
}

type ShortenURLRequest struct {
	URL string `json:"url"`
}

type ShortenURLSuccessResponse struct {
	Result string `json:"result"`
}

type ShortenURLErrorResponse struct {
	Error string `json:"error"`
}

func (h *Handler) HandleAPIShorten(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Header.Get("Content-Type") != "application/json" {
		b, err := json.Marshal(ShortenURLErrorResponse{Error: "Content-Type must be application/json"})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.l.Debugf("content type not application/json: %v", err)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		_, err = w.Write(b)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.l.Errorf("failed to write body: %v", err)
			return
		}
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		b, err := json.Marshal(ShortenURLErrorResponse{Error: "Internal Server Error #0"})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.l.Debugf("failed to marshal json body: %v", err)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write(b)
		h.l.Errorf("failed to write response body: %v", err)
		return
	}
	var request ShortenURLRequest
	err = json.Unmarshal(body, &request)
	if err != nil {
		b, err := json.Marshal(ShortenURLErrorResponse{Error: "JSON parse error"})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.l.Errorf("failed to marshal error response body: %v", err)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		_, err = w.Write(b)
		h.l.Errorf("failed to unmarshal the request body: %v", err)
		return
	}

	if request.URL == "" {
		b, err := json.Marshal(ShortenURLErrorResponse{Error: "URL can't be empty"})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.l.Errorf("failed to marshal error response body: %v", err)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		_, err = w.Write(b)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.l.Errorf("failed to write response body: %v", err)
			return
		}
		return
	}
	userID, ok := appctx.UserID(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	shortenURL, err := h.service.ShortenURL(request.URL, userID)
	if err != nil {

		var conflict *repository.ConflictError
		if errors.As(err, &conflict) {
			existingURL, _ := url.JoinPath(h.config.BaseAddr, conflict.ShortURL)
			w.WriteHeader(http.StatusConflict)
			b, err := json.Marshal(ShortenURLSuccessResponse{Result: existingURL})
			if err != nil {
				h.l.Errorf("failed to write success response body: %v", err)
				b, err := json.Marshal(ShortenURLErrorResponse{Error: "Internal Server Error #2"})
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					h.l.Errorf("failed to marshal error response body: %v", err)
					return
				}
				w.WriteHeader(http.StatusInternalServerError)
				_, err = w.Write(b)
				if err != nil {
					h.l.Errorf("failed to write response body: %v", err)
				}
				return
			}
			_, err = w.Write(b)
			if err != nil {
				b, err := json.Marshal(ShortenURLErrorResponse{Error: "Internal Server Error #3"})
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					h.l.Errorf("failed to marshal error response body: %v", err)
					return
				}
				w.WriteHeader(http.StatusInternalServerError)
				_, err = w.Write(b)
				h.l.Errorf("failed to write response body: %v", err)
				return
			}
			return
		}
		log.Printf("shorten url error: %v", err)
		b, err := json.Marshal(ShortenURLErrorResponse{Error: "Failed to shorten URL"})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.l.Errorf("failed to marshal error response body: %v", err)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write(b)
		h.l.Errorf("failed to write response body: %v", err)
		return
	}

	w.WriteHeader(http.StatusCreated)

	shortedURL, err := url.JoinPath(h.config.BaseAddr, shortenURL)
	if err != nil {
		b, err := json.Marshal(ShortenURLErrorResponse{Error: "Internal Server Error #1"})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.l.Errorf("failed to marshal error response body: %v", err)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write(b)
		h.l.Errorf("failed to write response body: %v", err)
		return
	}

	b, err := json.Marshal(ShortenURLSuccessResponse{Result: shortedURL})
	if err != nil {
		b, err := json.Marshal(ShortenURLErrorResponse{Error: "Internal Server Error #2"})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.l.Errorf("failed to marshal error response body: %v", err)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write(b)
		h.l.Errorf("failed to write response body: %v", err)
		return
	}
	_, err = w.Write(b)
	if err != nil {
		b, err := json.Marshal(ShortenURLErrorResponse{Error: "Internal Server Error #3"})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.l.Errorf("failed to marshal error response body: %v", err)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write(b)
		h.l.Errorf("failed to write response body: %v", err)
		return
	}
	h.sendAuditEvent(r.Context(), "shorten", userID, true, request.URL)
}

func (h *Handler) HandlePing(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := h.service.PingDatabase()
	if err == nil {
		w.WriteHeader(http.StatusOK)
		return
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		h.l.Errorf("failed ping: %v", err)
		return
	}
}

func (h *Handler) HandleAPIShortenBatch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var reqItems []model.BatchRequestItem
	if err := json.NewDecoder(r.Body).Decode(&reqItems); err != nil || len(reqItems) == 0 {
		b, err := json.Marshal(ShortenURLErrorResponse{Error: "Invalid or empty request"})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.l.Errorf("failed to marshal error response body: %v", err)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		_, err = w.Write(b)
		h.l.Errorf("failed to write response body: %v", err)
		return
	}
	userID, ok := appctx.UserID(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	respItems, err := h.service.ShortenBatch(reqItems, userID)
	if err != nil {
		b, err := json.Marshal(ShortenURLErrorResponse{Error: "Failed to shorten URL"})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.l.Errorf("failed to marshal error response body: %v", err)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write(b)
		h.l.Errorf("failed to write response body: %v", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(respItems)
	if err != nil {
		h.l.Errorf("Error encode: %v", err)
	}
}
