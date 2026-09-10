package users_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
)

func (r *UsersRepository) GetAvatarURL(
	ctx context.Context,
	userID int,
) (*string, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	db := r.dbFromContext(ctx)

	query := `
		SELECT avatar_url
		FROM debatesapp.users
		WHERE id = $1
	`

	var avatarURL *string

	err := db.QueryRow(ctx, query, userID).Scan(&avatarURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, core_errors.ErrNotFound
		}

		return nil, fmt.Errorf("get avatar url: %w", err)
	}

	return avatarURL, nil
}
