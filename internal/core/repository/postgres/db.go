package core_postgres

import (
	"context"

	core_postgres_pool "github.com/qandoni/debatesApp/internal/core/repository/postgres/pool"
)

type DB interface {
	Query(ctx context.Context, sql string, args ...any) (core_postgres_pool.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) core_postgres_pool.Row
	Exec(ctx context.Context, sql string, args ...any) (core_postgres_pool.CommandTag, error)
}
