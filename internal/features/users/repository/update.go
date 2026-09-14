package users_repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_postgres_pool "github.com/qandoni/debatesApp/internal/core/repository/postgres/pool"
)

const editProfileQuery = `
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

const updateAvatarURLQuery = `
WITH old_avatar AS (
	SELECT avatar_url
	FROM debatesApp.users
	WHERE id = $3
)
UPDATE debatesApp.users
SET
	avatar_url=$1,
	updated_at=$2,
	version=version+1
WHERE id=$3
RETURNING
	(SELECT avatar_url FROM old_avatar) AS old_avatar_url
`

func (r *UsersRepository) EditProfile(ctx context.Context, userID int, user domain.User) (domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	db := r.dbFromContext(ctx)
	row := db.QueryRow(
		ctx,
		editProfileQuery,
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

func (r *UsersRepository) UpdateAvatarURL(
	ctx context.Context,
	userID int,
	avatarURL string,
) (*string, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	db := r.dbFromContext(ctx)

	var oldAvatarURL *string

	err := db.QueryRow(
		ctx,
		updateAvatarURLQuery,
		avatarURL,
		time.Now(),
		userID,
	).Scan(&oldAvatarURL)
	if err != nil {
		return nil, fmt.Errorf("update avatar url: %w", err)
	}

	return oldAvatarURL, nil
}
