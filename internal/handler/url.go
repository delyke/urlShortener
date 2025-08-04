package handler

import (
	"encoding/json"
	"errors"
	"github.com/delyke/urlShortener/internal/config"
	"github.com/delyke/urlShortener/internal/model"
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

	shortenURL, err := h.service.ShortenURL(request.URL)
	if err != nil {
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
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
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

	respItems, err := h.service.ShortenBatch(reqItems)
	if err != nil {
		http.Error(w, "failed to shorten URLs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(respItems)
}
