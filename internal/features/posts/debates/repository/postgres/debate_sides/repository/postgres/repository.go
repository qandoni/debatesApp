package debate_sides_repository

import (
	"context"
	"time"

	core_postgres "github.com/qandoni/debatesApp/internal/core/repository/postgres"
)

func NewDebateSidesRepository(
	db core_postgres.DB,
	timeout time.Duration,
) *DebateSidesRepository {
	return &DebateSidesRepository{
		db,
		timeout,
	}
}

type DebateSidesRepository struct {
	db      core_postgres.DB
	timeout time.Duration
}

func (r *DebateSidesRepository) dbFromContext(
	ctx context.Context,
) core_postgres.DB {
	db := core_postgres.DBFromContext(ctx)
	if db != nil {
		return db
	}
	return r.db
}
