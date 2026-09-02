package debates_repository

import (
	"context"
	"time"

	core_postgres "github.com/qandoni/debatesApp/internal/core/repository/postgres"
)

func NewDebatesRepository(
	db core_postgres.DB,
	timeout time.Duration,
) *DebatesRepository {
	return &DebatesRepository{
		db,
		timeout,
	}
}

type DebatesRepository struct {
	db      core_postgres.DB
	timeout time.Duration
}

func (r *DebatesRepository) dbFromContext(
	ctx context.Context,
) core_postgres.DB {
	db := core_postgres.DBFromContext(ctx)
	if db != nil {
		return db
	}
	return r.db
}
