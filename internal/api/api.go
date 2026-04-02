package api

import (
	"context"
	"log/slog"

	"github.com/Arturikou/internal/models"
	"github.com/go-playground/validator"
)

type AuthService interface {
	Register(ctx context.Context, login string, password string) (int, error)
	Login(ctx context.Context, login string, password string) (int, error)
}

type TokenManager interface {
	BuildJWTString(userID int) (string, error)
	GetUserID(tokenString string) (int, error)
}

type BalanceService interface {
	Withdraw(ctx context.Context, withdraw models.Withdrawal) error
	GetUserBalance(ctx context.Context, userID int) (models.Balance, error)
	GetWithdrawals(ctx context.Context, userID int) ([]models.Withdrawal, error)
}

type Handler struct {
	logger         *slog.Logger
	authService    AuthService
	tokenManager   TokenManager
	balanceService BalanceService
	validate       *validator.Validate
}

func New(
	logger *slog.Logger,
	authService AuthService,
	tokenManager TokenManager,
	balanceService BalanceService,
) *Handler {
	return &Handler{
		logger:         logger,
		authService:    authService,
		tokenManager:   tokenManager,
		balanceService: balanceService,
		validate:       validator.New(),
	}
}
