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

func gzipMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w

		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGZIP := strings.Contains(acceptEncoding, "gzip")

		contentType := r.Header.Get("Content-Type")
		contentTypeAcceptEncoding := strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/html")
		if supportsGZIP && contentTypeAcceptEncoding {
			cw := newCompressWriter(w)
			ow = cw
			defer cw.Close()
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGZIP := strings.Contains(contentEncoding, "gzip")
		if sendsGZIP && contentTypeAcceptEncoding {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer r.Body.Close()
		}
		h.ServeHTTP(ow, r)
	})
}
