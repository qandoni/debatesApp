package users_repository

import (
	"context"
	"time"

	core_postgres "github.com/qandoni/debatesApp/internal/core/repository/postgres"
)

type UsersRepository struct {
	db      core_postgres.DB
	timeout time.Duration
}

func NewUsersRepository(
	db core_postgres.DB,
	timeout time.Duration,
) *UsersRepository {
	return &UsersRepository{
		db:      db,
		timeout: timeout,
	}
}

func (r *UsersRepository) dbFromContext(
	ctx context.Context,
) core_postgres.DB {
	db := core_postgres.DBFromContext(ctx)
	if db != nil {
		return db
	}
	return r.db
}
