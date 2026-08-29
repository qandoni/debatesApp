package users_repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_postgres_pool "github.com/qandoni/debatesApp/internal/core/repository/postgres/pool"
)

func (r *UsersRepository) EditProfile(ctx context.Context, userID int, user domain.User) (domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	query := `
	UPDATE debatesApp.users
	SET
		username=$1,
		email=$2,
		password_hash=$3,
		bio=$4,
		updated_at=$5,
		version=version+1
	WHERE id=$6
	AND version=$7
	RETURNING
		id,
		version,
		username,
		email,
		password_hash,
		avatar_url,
		bio,
		created_at,
		updated_at;
	`
	db := r.dbFromContext(ctx)
	row := db.QueryRow(
		ctx,
		query,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.Bio,
		time.Now(),
		userID,
		user.Version,
	)
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
			return domain.User{}, fmt.Errorf("user with id='%d' concurrently accessed: %w",
				userID,
				core_postgres_pool.ErrConflict)
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
