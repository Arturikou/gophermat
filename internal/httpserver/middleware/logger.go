package middleware

import (
	"log/slog"
	"net/http"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
)

func RequestLogger(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)

			t := time.Now()
			defer func() {
				logger.Info("http request",
					slog.String("path", r.URL.Path),
					slog.String("method", r.Method),
					slog.Int("status", ww.Status()),
					slog.Duration("duration", time.Since(t)),
					slog.Int("size", ww.BytesWritten()),
					slog.String("request_id", chimw.GetReqID(r.Context())),
				)
			}()
			next.ServeHTTP(ww, r)
		})
	}
}
