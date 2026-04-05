package balance

import (
	"context"
	"errors"
	"testing"

	"github.com/Arturikou/internal/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBalanceService_Withdraw(t *testing.T) {
	ctx := context.Background()
	withdrawal := models.Withdrawal{
		UserID: 1,
		Amount: decimal.NewFromFloat(100.50),
	}
	errTest := errors.New("db error")

	t.Run("success", func(t *testing.T) {
		repo := NewMockBalanceRepo(t)
		service := NewBalanceService(repo)

		repo.
			On("WithTx", ctx, mock.AnythingOfType("func(context.Context) error")).
			Return(nil).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			err := fn(ctx)
			assert.NoError(t, err)
		})

		repo.
			On("Withdraw", ctx, withdrawal.UserID, withdrawal.Amount).
			Return(nil)

		repo.
			On("AddWithdrawal", ctx, withdrawal).
			Return(nil)

		err := service.Withdraw(ctx, withdrawal)
		assert.NoError(t, err)
	})

	t.Run("withdraw error - rollback", func(t *testing.T) {
		repo := NewMockBalanceRepo(t)
		service := NewBalanceService(repo)

		repo.
			On("WithTx", ctx, mock.AnythingOfType("func(context.Context) error")).
			Return(errTest).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		})

		repo.
			On("Withdraw", ctx, withdrawal.UserID, withdrawal.Amount).
			Return(errTest)

		err := service.Withdraw(ctx, withdrawal)
		assert.ErrorIs(t, err, errTest)
	})

	t.Run("add withdrawal error - rollback", func(t *testing.T) {
		repo := NewMockBalanceRepo(t)
		service := NewBalanceService(repo)

		repo.
			On("WithTx", ctx, mock.AnythingOfType("func(context.Context) error")).
			Return(errTest).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		})

		repo.
			On("Withdraw", ctx, withdrawal.UserID, withdrawal.Amount).
			Return(nil)

		repo.
			On("AddWithdrawal", ctx, withdrawal).
			Return(errTest)

		err := service.Withdraw(ctx, withdrawal)
		assert.ErrorIs(t, err, errTest)
	})
}

func TestBalanceService_GetUserBalance(t *testing.T) {
	ctx := context.Background()
	userID := 1
	expectedBalance := models.Balance{
		Current:   decimal.NewFromInt(500),
		Withdrawn: decimal.NewFromInt(100),
	}
	errTest := errors.New("db error")

	t.Run("success", func(t *testing.T) {
		repo := NewMockBalanceRepo(t)
		service := NewBalanceService(repo)

		repo.On("GetUserBalance", ctx, userID).Return(expectedBalance, nil)

		balance, err := service.GetUserBalance(ctx, userID)
		assert.NoError(t, err)
		assert.Equal(t, expectedBalance, balance)
	})

	t.Run("repo error", func(t *testing.T) {
		repo := NewMockBalanceRepo(t)
		service := NewBalanceService(repo)

		repo.On("GetUserBalance", ctx, userID).Return(models.Balance{}, errTest)

		_, err := service.GetUserBalance(ctx, userID)
		assert.ErrorIs(t, err, errTest)
	})
}

func TestBalanceService_GetWithdrawals(t *testing.T) {
	ctx := context.Background()
	userID := 1
	expectedList := []models.Withdrawal{
		{UserID: 1, Amount: decimal.NewFromInt(50)},
	}
	errTest := errors.New("db error")

	t.Run("success", func(t *testing.T) {
		repo := NewMockBalanceRepo(t)
		service := NewBalanceService(repo)

		repo.On("GetWithdrawals", ctx, userID).Return(expectedList, nil)

		res, err := service.GetWithdrawals(ctx, userID)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
		assert.Equal(t, expectedList, res)
	})

	t.Run("repo error", func(t *testing.T) {
		repo := NewMockBalanceRepo(t)
		service := NewBalanceService(repo)

		repo.On("GetWithdrawals", ctx, userID).Return(nil, errTest)

		_, err := service.GetWithdrawals(ctx, userID)
		assert.ErrorIs(t, err, errTest)
	})
}

func TestBalanceService_UpdateUserBalance(t *testing.T) {
	ctx := context.Background()
	userID := 1
	amount := decimal.NewFromInt(1000)
	errTest := errors.New("db error")

	t.Run("success", func(t *testing.T) {
		repo := NewMockBalanceRepo(t)
		service := NewBalanceService(repo)

		repo.On("UpdateUserBalance", ctx, userID, amount).Return(nil)

		err := service.UpdateUserBalance(ctx, userID, amount)
		assert.NoError(t, err)
	})

	t.Run("repo error", func(t *testing.T) {
		repo := NewMockBalanceRepo(t)
		service := NewBalanceService(repo)

		repo.On("UpdateUserBalance", ctx, userID, amount).Return(errTest)

		err := service.UpdateUserBalance(ctx, userID, amount)
		assert.ErrorIs(t, err, errTest)
	})
}
