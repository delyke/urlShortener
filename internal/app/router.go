package app

import (
	"compress/gzip"
	"encoding/json"
	"github.com/delyke/urlShortener/internal/handler"
	"github.com/delyke/urlShortener/internal/logger"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"strings"
)

func NewRouter(h *handler.Handler, l *logger.Logger) chi.Router {
	r := chi.NewRouter()
	r.Use(gzipMiddleware)
	r.Route("/", func(r chi.Router) {
		r.With(l.RequestLogger).Post("/", h.HandlePost)
		r.With(l.RequestLogger).Get("/", h.HandleGet)
		r.With(l.RequestLogger).Get("/{shortURL}", h.HandleGet)
		r.With(l.RequestLogger).Post("/api/shorten", h.HandleAPIShorten)
		r.With(l.RequestLogger).Post("/api/shorten/batch", h.HandleAPIShortenBatch)
		r.With(l.RequestLogger).Get("/ping", h.HandlePing)
	})
	return r
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func gzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				b, err := json.Marshal(ErrorResponse{Error: "Invalid gzip body"})
				if err != nil {
					log.Println("Ошибка кодировки ответа в json", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusBadRequest)
				_, err = w.Write(b)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					log.Println("Ошибка возврата данных: ", err)
					return
				}
			}
			defer gz.Close()
			r.Body = gz
		}

		originalWriter := w
		clientAcceptGzip := strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
		if clientAcceptGzip {
			isRelevantContentType := strings.Contains(r.Header.Get("Content-Type"), "application/json") ||
				strings.Contains(r.Header.Get("Content-Type"), "text/html")
			if isRelevantContentType {
				gzWriter := gzip.NewWriter(w)
				defer gzWriter.Close()
				w.Header().Set("Content-Encoding", "gzip")
				originalWriter = &gzipWriter{
					ResponseWriter: w,
					Writer:         gzWriter,
				}
			}
		}
		next.ServeHTTP(originalWriter, r)
	})
}
