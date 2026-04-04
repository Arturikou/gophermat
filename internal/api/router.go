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

			r.Group(func(r chi.Router) {
				r.Use(mw.Auth(h.tokenManager))
				r.Get("/balance", h.GetBalance)
				r.Post("/balance/withdraw", h.Withdraw)
				r.Get("/withdrawals", h.GetWithdrawals)

				r.Post("/orders", h.Orders)
				r.Get("/orders", h.GetOrders)
			})
		})
	})

	return r
}
