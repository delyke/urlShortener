package app

import (
	"github.com/delyke/urlShortener/internal/handler"
	"github.com/delyke/urlShortener/internal/logger"
	"github.com/go-chi/chi/v5"
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
	})
	return r
}

func gzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isPerfectContentType := strings.Contains(r.Header.Get("Content-Type"), "application/json") ||
			strings.Contains(r.Header.Get("Content-Type"), "text/html")
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") && isPerfectContentType {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				http.Error(w, "failed to decompress request", http.StatusInternalServerError)
				return
			}
			defer cr.Close()
			r.Body = cr
		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") || !isPerfectContentType {
			next.ServeHTTP(w, r)
			return
		}

		gzw := newCompressWriter(w)
		defer gzw.Close()

		next.ServeHTTP(gzw, r)
	})
}
