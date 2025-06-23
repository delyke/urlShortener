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

	originalUrl := string(body)
	originalUrl = strings.TrimSpace(originalUrl)
	if originalUrl == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortedUrl, err := h.service.ShortenUrl(originalUrl)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	shortedUrl = "http://localhost:8080/" + shortedUrl
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "text/plain")
	_, _ = w.Write([]byte(shortedUrl))
}

func (h *Handler) HandleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	shortedUrl := strings.TrimPrefix(r.URL.Path, "/")
	if shortedUrl == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	originalUrl, err := h.service.GetOriginalUrl(shortedUrl)
	if err == nil {
		http.Redirect(w, r, originalUrl, http.StatusTemporaryRedirect)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}
