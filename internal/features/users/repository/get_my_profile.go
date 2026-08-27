package users_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
	core_postgres_pool "github.com/qandoni/debatesApp/internal/core/repository/postgres/pool"
)

func (r *UsersRepository) GetMyProfile(ctx context.Context, userID int) (domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	query := `
	SELECT id, version, username, email, password_hash, avatar_url, bio, created_at, updated_at
	FROM debatesApp.users
	WHERE id=$1;
	`
	db := r.dbFromContext(ctx)
	row := db.QueryRow(ctx, query, userID)
	var userModel UserModel

	err := row.Scan(
		&userModel.ID,
		&userModel.Version,
		&userModel.Username,
		&userModel.Email,
		&userModel.PasswordHash,
		&userModel.AvatarURL,
		&userModel.Bio,
		&userModel.CreatedAt,
		&userModel.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return domain.User{}, fmt.Errorf("user with id='%v': %w", userID, core_errors.ErrNotFound)
		}
		return domain.User{}, fmt.Errorf("scan error: %w", err)
	}

	userDomain := domain.NewUser(
		userModel.ID,
		userModel.Version,
		userModel.Username,
		userModel.Email,
		userModel.PasswordHash,
		userModel.AvatarURL,
		userModel.Bio,
		userModel.CreatedAt,
		userModel.UpdatedAt,
	)
	return userDomain, nil
}
