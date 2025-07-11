package handler

import (
	"errors"
	"github.com/delyke/urlShortener/internal/config"
	"github.com/delyke/urlShortener/internal/service"
	"github.com/go-chi/chi/v5"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type Handler struct {
	service ShortenURLService
	config  *config.Config
}

func NewHandler(service ShortenURLService, cfg *config.Config) *Handler {
	return &Handler{service: service, config: cfg}
}

func (h *Handler) HandlePost(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("failed to read the request body: %v", err)
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
		if errors.Is(err, service.ErrCanNotCreateURL) {
			log.Println("URL shortening error:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	shortedURL, err = url.JoinPath(h.config.BaseAddr, shortedURL)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(shortedURL))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
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
		if errors.Is(err, service.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}
	}
}
