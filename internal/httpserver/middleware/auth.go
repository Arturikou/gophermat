package middleware

import (
	"net/http"
	"strings"

	"github.com/Arturikou/internal/ctxkeys"
)

type TokenManager interface {
	GetUserID(tokenString string) (int, error)
}

func Auth(m TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			userID, err := m.GetUserID(tokenString)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r.WithContext(ctxkeys.SetUserID(r.Context(), userID)))
		})
	}
}
