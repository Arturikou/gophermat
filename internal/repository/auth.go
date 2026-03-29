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

func (r *Repo) AddUser(ctx context.Context, login string, password string) (int, error) {
	var userID int

	query := `INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id`

	err := r.db.QueryRow(ctx, query, login, password).Scan(&userID)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == pgerrcode.UniqueViolation {
			return 0, models.ErrLoginAlreadyExists
		}

		return 0, fmt.Errorf("add user: %w", err)
	}

	return userID, nil
}

func (r *Repo) GetUserByLogin(ctx context.Context, login string) (models.User, error) {
	var user models.User

	query := `SELECT id, login, password_hash FROM users WHERE login = $1`

	err := r.db.QueryRow(ctx, query, login).Scan(&user.ID, &user.Login, &user.PasswordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, models.ErrLoginNotFound
		}

		return models.User{}, fmt.Errorf("get user: %w", err)
	}

	return user, nil
}
