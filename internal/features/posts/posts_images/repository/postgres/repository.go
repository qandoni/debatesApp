package post_images_repository

import (
	"context"
	"time"

	core_postgres "github.com/qandoni/debatesApp/internal/core/repository/postgres"
)

func NewPostImagesRepository(
	db core_postgres.DB,
	timeout time.Duration,
) *PostImagesRepository {
	return &PostImagesRepository{
		db,
		timeout,
	}
}

type PostImagesRepository struct {
	db      core_postgres.DB
	timeout time.Duration
}

func (r *PostImagesRepository) dbFromContext(
	ctx context.Context,
) core_postgres.DB {
	db := core_postgres.DBFromContext(ctx)
	if db != nil {
		return db
	}
	return r.db
}
