package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/delyke/urlShortener/internal/app/appctx"
	"github.com/delyke/urlShortener/internal/config"
	"github.com/delyke/urlShortener/internal/model"
	"github.com/delyke/urlShortener/internal/repository"
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

	userID, ok := appctx.UserID(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	shortedURL, err := h.service.ShortenURL(originalURL, userID)
	if err != nil {
		if errors.Is(err, service.ErrCanNotCreateURL) {
			log.Println("URL shortening error:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		var conflict *repository.ConflictError
		if errors.As(err, &conflict) {
			existingURL, _ := url.JoinPath(h.config.BaseAddr, conflict.ShortURL)
			log.Println("URL shortening conflict:", existingURL)
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte(existingURL))
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
			log.Println(err)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	var resp []respUserURLs
	for _, url := range *urls {
		resp = append(resp, respUserURLs{ShortURL: fmt.Sprintf("%s/%s", h.config.BaseAddr, url.ShortURL), OriginalURL: url.OriginalURL})
	}
	b, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(b)
}

func (h *Handler) HandleGet(w http.ResponseWriter, r *http.Request) {
	shortedURL := chi.URLParam(r, "shortURL")
	originalURL, err := h.service.GetOriginalURL(shortedURL)
	if err == nil {
		http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
		return
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
			log.Println(err)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		_, err = w.Write(b)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		b, err := json.Marshal(ShortenURLErrorResponse{Error: "Internal Server Error #0"})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write(b)
		log.Printf("failed to read the request body: %v", err)
		return
	}
	var request ShortenURLRequest
	err = json.Unmarshal(body, &request)
	if err != nil {
		b, err := json.Marshal(ShortenURLErrorResponse{Error: "JSON parse error"})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		_, err = w.Write(b)
		log.Printf("failed to unmarshal the request body: %v", err)
		return
	}

	if request.URL == "" {
		b, err := json.Marshal(ShortenURLErrorResponse{Error: "URL can't be empty"})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		_, err = w.Write(b)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}
		log.Printf("URL cannot be empty.")
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
			log.Println("URL shortening conflict:", existingURL)
			w.WriteHeader(http.StatusConflict)
			b, err := json.Marshal(ShortenURLSuccessResponse{Result: existingURL})
			if err != nil {
				b, err := json.Marshal(ShortenURLErrorResponse{Error: "Internal Server Error #2"})
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					log.Println(err)
					return
				}
				w.WriteHeader(http.StatusInternalServerError)
				_, err = w.Write(b)
				log.Println(err)
				return
			}
			_, err = w.Write(b)
			if err != nil {
				b, err := json.Marshal(ShortenURLErrorResponse{Error: "Internal Server Error #3"})
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					log.Println(err)
					return
				}
				w.WriteHeader(http.StatusInternalServerError)
				_, err = w.Write(b)
				log.Println(err)
				return
			}
			return
		}
		log.Printf("shorten url error: %v", err)
		b, err := json.Marshal(ShortenURLErrorResponse{Error: "Failed to shorten URL"})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write(b)
		log.Printf("failed to short url: %v", err)
		return
	}

	w.WriteHeader(http.StatusCreated)

	shortedURL, err := url.JoinPath(h.config.BaseAddr, shortenURL)
	if err != nil {
		b, err := json.Marshal(ShortenURLErrorResponse{Error: "Internal Server Error #1"})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write(b)
		log.Println(err)
		return
	}

	b, err := json.Marshal(ShortenURLSuccessResponse{Result: shortedURL})
	if err != nil {
		b, err := json.Marshal(ShortenURLErrorResponse{Error: "Internal Server Error #2"})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write(b)
		log.Println(err)
		return
	}
	_, err = w.Write(b)
	if err != nil {
		b, err := json.Marshal(ShortenURLErrorResponse{Error: "Internal Server Error #3"})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write(b)
		log.Println(err)
		return
	}
}

func (h *Handler) HandlePing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := h.service.PingDatabase()
	if err == nil {
		w.WriteHeader(http.StatusOK)
		return
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
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
			log.Println(err)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		_, err = w.Write(b)
		log.Println(err)
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
			log.Println(err)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write(b)
		log.Println(err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(respItems)
	if err != nil {
		log.Printf("Error encode: %v", err)
	}
}
