package handler

import (
	"fmt"
	"github.com/delyke/urlShortener/internal/config"
	"github.com/delyke/urlShortener/internal/service"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"strings"
)

type Handler struct {
	service service.ShortenURLService
}

func NewHandler(service service.ShortenURLService) *Handler {
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

	shortedURL = fmt.Sprintf("%s/%s", config.FlagBaseAddr, shortedURL)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(shortedURL))
}

func (h *Handler) HandleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	shortedURL := chi.URLParam(r, "shortURL")

	originalURL, err := h.service.GetOriginalURL(shortedURL)
	if err == nil {
		http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
	} else {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}
