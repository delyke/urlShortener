package app

import (
	"github.com/delyke/urlShortener/internal/handler"
	"github.com/go-chi/chi/v5"
)

func NewRouter(h *handler.Handler) chi.Router {
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Post("/", h.HandlePost)
		r.Get("/", h.HandleGet)
		r.Get("/{shortURL}", h.HandleGet)
	})
	return r
}
