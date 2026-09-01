package posts_repository

import (
	"context"
	"time"

	core_postgres "github.com/qandoni/debatesApp/internal/core/repository/postgres"
)

func NewPostsRepository(
	db core_postgres.DB,
	timeout time.Duration,
) *PostsRepository {
	return &PostsRepository{
		db,
		timeout,
	}
}

type PostsRepository struct {
	db      core_postgres.DB
	timeout time.Duration
}

func (r *PostsRepository) dbFromContext(
	ctx context.Context,
) core_postgres.DB {
	db := core_postgres.DBFromContext(ctx)
	if db != nil {
		return db
	}
	return r.db
}
