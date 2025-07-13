package app

import (
	"github.com/delyke/urlShortener/internal/handler"
	"github.com/delyke/urlShortener/internal/logger"
	"github.com/go-chi/chi/v5"
)

func NewRouter(h *handler.Handler, l *logger.Logger) chi.Router {
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.With(l.RequestLogger).Post("/", h.HandlePost)
		r.With(l.RequestLogger).Get("/", h.HandleGet)
		r.With(l.RequestLogger).Get("/{shortURL}", h.HandleGet)
		r.With(l.RequestLogger).Post("/api/shorten", h.HandleAPIShorten)
	})
	return r
}
