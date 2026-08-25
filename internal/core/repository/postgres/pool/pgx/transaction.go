package core_pgx_pool

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	core_postgres "github.com/qandoni/debatesApp/internal/core/repository/postgres"
)

type TransactionManager struct {
	pool *Pool
}

func NewTransactionManager(pool *Pool) *TransactionManager {
	return &TransactionManager{
		pool: pool,
	}
}

func (m *TransactionManager) WithinTransaction(
	ctx context.Context,
	fn func(ctx context.Context) error,
) error {
	tx, err := m.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	defer tx.Rollback(ctx)

	ctx = core_postgres.ContextWithDB(
		ctx,
		&Tx{Tx: tx},
	)

	if err := fn(ctx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}
