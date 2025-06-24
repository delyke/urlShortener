package handler

import (
	"github.com/delyke/urlShortener/internal/service"
	"io"
	"net/http"
	"strings"
)

type Handler struct {
	service *service.URLService
}

func NewHandler(service *service.URLService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) HandlePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	originalURL := string(body)
	originalURL = strings.TrimSpace(originalURL)
	if originalURL == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortedURL, err := h.service.ShortenURL(originalURL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	shortedURL = "http://localhost:8080/" + shortedURL
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(shortedURL))
}

func (h *Handler) HandleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	shortedURL := strings.TrimPrefix(r.URL.Path, "/")
	if shortedURL == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	originalURL, err := h.service.GetOriginalURL(shortedURL)
	if err == nil {
		http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}
