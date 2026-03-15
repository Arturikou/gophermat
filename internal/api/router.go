package api

import (
	mw "github.com/Arturikou/internal/httpserver/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (a *Handler) Router() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(mw.RequestLogger(a.logger))
	r.Use(mw.Gzip)

	return r
}
