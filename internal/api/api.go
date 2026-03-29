package api

import (
	"context"
	"log/slog"

	"github.com/go-playground/validator"
)

type AuthService interface {
	Register(ctx context.Context, login, password string) (int, error)
	Login(ctx context.Context, login string, password string) (int, error)
}

type TokenManager interface {
	BuildJWTString(userID int) (string, error)
}

type Handler struct {
	logger       *slog.Logger
	authService  AuthService
	tokenManager TokenManager
	validate     *validator.Validate
}

func New(logger *slog.Logger, authService AuthService, tokenManager TokenManager) *Handler {
	return &Handler{
		logger:       logger,
		authService:  authService,
		tokenManager: tokenManager,
		validate:     validator.New(),
	}
}
