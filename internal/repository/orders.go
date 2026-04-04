package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Arturikou/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (r *Repo) AddOrder(ctx context.Context, orderNumber string, userID int) error {
	query := `INSERT INTO orders(number, user_id) VALUES ($1, $2)`

	_, err := r.exec(ctx, query, orderNumber, userID)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == pgerrcode.UniqueViolation {
			return models.ErrDuplicateOrder
		}

		return fmt.Errorf("add order number: %w", err)
	}

	return nil
}

func (r *Repo) GetUserIDByOrderNumber(ctx context.Context, number string) (int, error) {
	query := `SELECT user_id FROM orders WHERE number = $1`
	var userID int

	err := r.queryRow(ctx, query, number).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("get user by order number: %w", err)
	}

	return userID, nil
}

func (r *Repo) GetOrdersByUserID(ctx context.Context, userID int) ([]models.Order, error) {
	query := `
		SELECT number, status, accrual, uploaded_at 
		FROM orders WHERE user_id = $1 
		ORDER BY uploaded_at DESC
		`

	rows, err := r.query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("get orders by user id: %w", err)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		if err = rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.UploadedAt); err != nil {
			return nil, fmt.Errorf("get orders by user id: %w", err)
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func (r *Repo) GetOrderNumbersInStatus(ctx context.Context, status []models.OrderStatus) ([]string, error) {
	query := `
		SELECT number FROM orders 
		WHERE status = ANY($1) 
		ORDER BY uploaded_at 
		LIMIT 100;
	`
	statusStrings := make([]string, len(status))
	for i, s := range status {
		statusStrings[i] = string(s)
	}

	rows, err := r.query(ctx, query, statusStrings)
	if err != nil {
		return nil, fmt.Errorf("get order numbers: %w", err)
	}
	defer rows.Close()
	var numbers []string
	for rows.Next() {
		var number string
		if err = rows.Scan(&number); err != nil {
			return nil, fmt.Errorf("get order numbers: %w", err)
		}
		numbers = append(numbers, number)
	}

	return numbers, nil
}

func (r *Repo) UpdateOrders(ctx context.Context, orders []models.Order) error {
	query := `
		UPDATE orders
		SET status = $1, accrual = $2
		WHERE number = $3 AND (status IS DISTINCT FROM $1)
	`

	batch := &pgx.Batch{}
	for _, order := range orders {
		batch.Queue(query, order.Status, order.Accrual, order.Number)
	}

	br := r.sendBatch(ctx, batch)
	defer br.Close()

	for range orders {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("update order: %w", err)
		}
	}

	return nil
}
