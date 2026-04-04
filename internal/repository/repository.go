package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
}

type txKey struct{}

type Repo struct {
	pool *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repo {
	return &Repo{pool: db}
}

func (r *Repo) conn(ctx context.Context) querier {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return r.pool
}

func (r *Repo) exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return r.conn(ctx).Exec(ctx, sql, args...)
}

func (r *Repo) queryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return r.conn(ctx).QueryRow(ctx, sql, args...)
}

func (r *Repo) query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return r.conn(ctx).Query(ctx, sql, args...)
}

func (r *Repo) sendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	return r.conn(ctx).SendBatch(ctx, b)
}
