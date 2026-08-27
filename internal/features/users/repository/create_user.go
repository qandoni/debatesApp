package users_repository

import (
	"context"
	"fmt"

	"github.com/qandoni/debatesApp/internal/core/domain"
)

func (r *UsersRepository) CreateUser(
	ctx context.Context,
	user domain.User,
) (domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	query := `INSERT INTO debatesApp.users(username, email, password_hash, avatar_url, bio, created_at, updated_at)
	VALUES($1, $2, $3, $4, $5, $6, $7)
	RETURNING id, version, username, email, password_hash, avatar_url, bio, created_at, updated_at
	`

	db := r.dbFromContext(ctx)
	row := db.QueryRow(
		ctx,
		query,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.AvatarURL,
		user.Bio,
		user.CreatedAt,
		user.UpdatedAt,
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
