package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

var ErrContextKeyNotFound = errors.New("key not found in context")
var ErrContextValueWrongType = errors.New("context value has unexpected type")

type ContextKey string

const ContextKeyUserID ContextKey = "userID"

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

			ctx := context.WithValue(r.Context(), ContextKeyUserID, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) (int, error) {
	val := ctx.Value(ContextKeyUserID)
	if val == nil {
		return 0, ErrContextKeyNotFound
	}

	id, ok := val.(int)
	if !ok {
		return 0, ErrContextValueWrongType
	}

	return id, nil
}
