package debate_votes_repository

import (
	"context"
	"time"

	core_postgres "github.com/qandoni/debatesApp/internal/core/repository/postgres"
)

func NewDebateVotesRepository(
	db core_postgres.DB,
	timeout time.Duration,
) *DebateVotesRepository {
	return &DebateVotesRepository{
		db,
		timeout,
	}
}

type DebateVotesRepository struct {
	db      core_postgres.DB
	timeout time.Duration
}

func (r *DebateVotesRepository) dbFromContext(
	ctx context.Context,
) core_postgres.DB {
	db := core_postgres.DBFromContext(ctx)
	if db != nil {
		return db
	}
	return r.db
}
