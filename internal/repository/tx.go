package repository

import (
	"context"
	"fmt"
)

func (r *Repo) WithTx(ctx context.Context, fn func(r *Repo) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := fn(&Repo{db: tx, pool: r.pool}); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
