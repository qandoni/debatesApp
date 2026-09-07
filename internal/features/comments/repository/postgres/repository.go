package comments_repository

import (
	"context"
	"time"

	core_postgres "github.com/qandoni/debatesApp/internal/core/repository/postgres"
)

func NewCommentsRepository(
	db core_postgres.DB,
	timeout time.Duration,
) *CommentsRepository {
	return &CommentsRepository{
		db,
		timeout,
	}
}

type CommentsRepository struct {
	db      core_postgres.DB
	timeout time.Duration
}

func (r *CommentsRepository) dbFromContext(
	ctx context.Context,
) core_postgres.DB {
	db := core_postgres.DBFromContext(ctx)
	if db != nil {
		return db
	}
	return r.db
}
