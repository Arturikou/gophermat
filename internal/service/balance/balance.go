package balance

import (
	"context"

	"github.com/Arturikou/internal/models"
	"github.com/shopspring/decimal"
)

//go:generate mockery --name=BalanceRepo --filename=mock_balance_repo_test.go --inpackage --disable-version-string
type BalanceRepo interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
	Withdraw(ctx context.Context, userID int, amount decimal.Decimal) error
	AddWithdrawal(ctx context.Context, withdrawal models.Withdrawal) error
	GetUserBalance(ctx context.Context, userID int) (models.Balance, error)
	GetWithdrawals(ctx context.Context, userID int) ([]models.Withdrawal, error)
	UpdateUserBalance(ctx context.Context, userID int, amount decimal.Decimal) error
}

type BalanceService struct {
	repo BalanceRepo
}

func NewBalanceService(
	repo BalanceRepo,
) *BalanceService {
	return &BalanceService{
		repo: repo,
	}
}

func (b *BalanceService) Withdraw(ctx context.Context, withdrawal models.Withdrawal) error {
	return b.repo.WithTx(ctx, func(ctx context.Context) error {
		if err := b.repo.Withdraw(ctx, withdrawal.UserID, withdrawal.Amount); err != nil {
			return err
		}

		return b.repo.AddWithdrawal(ctx, withdrawal)
	})
}

func (b *BalanceService) GetUserBalance(ctx context.Context, userID int) (models.Balance, error) {
	return b.repo.GetUserBalance(ctx, userID)
}

func (b *BalanceService) GetWithdrawals(ctx context.Context, userID int) ([]models.Withdrawal, error) {
	return b.repo.GetWithdrawals(ctx, userID)
}

func (b *BalanceService) UpdateUserBalance(ctx context.Context, userID int, amount decimal.Decimal) error {
	return b.repo.UpdateUserBalance(ctx, userID, amount)
}
