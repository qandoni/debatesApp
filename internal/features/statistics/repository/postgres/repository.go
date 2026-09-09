package statistics_repository

import (
	"context"
	"time"

	core_postgres "github.com/qandoni/debatesApp/internal/core/repository/postgres"
)

func NewDebateStatisticsRepository(
	db core_postgres.DB,
	timeout time.Duration,
) *DebateStatisticsRepository {
	return &DebateStatisticsRepository{
		db,
		timeout,
	}
}

type DebateStatisticsRepository struct {
	db      core_postgres.DB
	timeout time.Duration
}

func (r *DebateStatisticsRepository) dbFromContext(
	ctx context.Context,
) core_postgres.DB {
	db := core_postgres.DBFromContext(ctx)
	if db != nil {
		return db
	}
	return r.db
}
