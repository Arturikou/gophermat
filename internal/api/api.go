package api

import (
	"context"
	"log/slog"

	"github.com/Arturikou/internal/models"
	"github.com/go-playground/validator"
)

//go:generate mockery --name=AuthService --filename=mock_auth_service_test.go --inpackage --disable-version-string
type AuthService interface {
	Register(ctx context.Context, login string, password string) (int, error)
	Login(ctx context.Context, login string, password string) (int, error)
}

//go:generate mockery --name=TokenManager --filename=mock_token_manager_test.go --inpackage --disable-version-string
type TokenManager interface {
	BuildJWTString(userID int) (string, error)
	GetUserID(tokenString string) (int, error)
}

//go:generate mockery --name=BalanceService --filename=mock_balance_service_test.go --inpackage --disable-version-string
type BalanceService interface {
	Withdraw(ctx context.Context, withdraw models.Withdrawal) error
	GetUserBalance(ctx context.Context, userID int) (models.Balance, error)
	GetWithdrawals(ctx context.Context, userID int) ([]models.Withdrawal, error)
}

//go:generate mockery --name=OrderService --filename=mock_order_service_test.go --inpackage --disable-version-string
type OrderService interface {
	AddOrder(ctx context.Context, orderNumber string, userID int) error
	GetUserOrders(ctx context.Context, userID int) ([]models.Order, error)
}

type Handler struct {
	logger         *slog.Logger
	authService    AuthService
	tokenManager   TokenManager
	balanceService BalanceService
	orderService   OrderService
	validate       *validator.Validate
}

func New(
	logger *slog.Logger,
	authService AuthService,
	tokenManager TokenManager,
	balanceService BalanceService,
	orderService OrderService,
) *Handler {
	return &Handler{
		logger:         logger,
		authService:    authService,
		tokenManager:   tokenManager,
		balanceService: balanceService,
		orderService:   orderService,
		validate:       validator.New(),
	}
}
