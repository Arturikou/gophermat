package api

import (
	mw "github.com/Arturikou/internal/httpserver/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (h *Handler) Router() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(mw.RequestLogger(h.logger))
	r.Use(mw.Gzip)

	r.Route("/api", func(r chi.Router) {
		r.Route("/user", func(r chi.Router) {
			r.Post("/register", h.Register)
			r.Post("/login", h.Login)
		})
	})

	return r
}
