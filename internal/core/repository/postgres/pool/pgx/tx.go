package core_pgx_pool

import (
	"context"

	"github.com/jackc/pgx/v5"
	core_postgres_pool "github.com/qandoni/debatesApp/internal/core/repository/postgres/pool"
)

type Tx struct {
	pgx.Tx
}

func (tx *Tx) Query(
	ctx context.Context,
	sql string,
	args ...any,
) (core_postgres_pool.Rows, error) {

	rows, err := tx.Tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	return pgxRows{rows}, nil
}

func (tx *Tx) Exec(
	ctx context.Context,
	sql string,
	args ...any,
) (core_postgres_pool.CommandTag, error) {

	tag, err := tx.Tx.Exec(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	return pgxCommandTag{tag}, nil
}

func (tx *Tx) QueryRow(
	ctx context.Context,
	sql string,
	args ...any,
) core_postgres_pool.Row {
	row := tx.Tx.QueryRow(ctx, sql, args...)
	return row
}
