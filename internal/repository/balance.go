package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Arturikou/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shopspring/decimal"
)

func (r *Repo) Withdraw(ctx context.Context, userID int, amount decimal.Decimal) error {
	query := `
		UPDATE bonus_balances
		SET current_amount=current_amount - $1,
    		total_withdrawn = total_withdrawn + $1,
    		updated_at = NOW()
		WHERE user_id = $2
    `

	_, err := r.exec(ctx, query, amount, userID)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == pgerrcode.CheckViolation {
			return models.ErrInsufficientFunds
		}

		return fmt.Errorf("withdraw balance: %w", err)
	}

	return nil
}

func (r *Repo) AddWithdrawal(ctx context.Context, withdraw models.Withdrawal) error {
	query := `
	INSERT INTO bonus_withdrawals (user_id, order_number, amount)
	VALUES ($1, $2, $3)
	`
	_, err := r.exec(ctx, query, withdraw.UserID, withdraw.OrderNumber, withdraw.Amount)
	if err != nil {
		return fmt.Errorf("add withdrawal: %w", err)
	}

	return nil
}

func (r *Repo) GetUserBalance(ctx context.Context, userID int) (models.Balance, error) {
	var balance models.Balance

	query := `SELECT current_amount, total_withdrawn FROM bonus_balances WHERE user_id = $1`

	err := r.queryRow(ctx, query, userID).Scan(&balance.Current, &balance.Withdrawn)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Balance{}, nil
		}

		return models.Balance{}, fmt.Errorf("get user balance: %w", err)
	}

	return balance, nil
}

func (r *Repo) GetWithdrawals(ctx context.Context, userID int) ([]models.Withdrawal, error) {
	query := `SELECT order_number, amount, processed_at FROM bonus_withdrawals WHERE user_id = $1`

	rows, err := r.query(ctx, query, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []models.Withdrawal{}, nil
		}

		return nil, fmt.Errorf("get user withdrawals: %w", err)
	}
	defer rows.Close()

	var withdraws []models.Withdrawal
	for rows.Next() {
		var withdraw models.Withdrawal
		if err = rows.Scan(&withdraw.OrderNumber, &withdraw.Amount, &withdraw.ProcessedAt); err != nil {
			return nil, fmt.Errorf("get user withdrawals: %w", err)
		}
		withdraws = append(withdraws, withdraw)
	}

	return withdraws, nil
}
