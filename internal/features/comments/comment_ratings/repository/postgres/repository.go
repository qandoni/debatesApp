package comments_ratings_repository

import (
	"context"
	"time"

	core_postgres "github.com/qandoni/debatesApp/internal/core/repository/postgres"
)

func NewCommentRatingsRepository(
	db core_postgres.DB,
	timeout time.Duration,
) *CommentRatingsRepository {
	return &CommentRatingsRepository{
		db,
		timeout,
	}
}

type CommentRatingsRepository struct {
	db      core_postgres.DB
	timeout time.Duration
}

func (r *CommentRatingsRepository) dbFromContext(
	ctx context.Context,
) core_postgres.DB {
	db := core_postgres.DBFromContext(ctx)
	if db != nil {
		return db
	}
	return r.db
}
