package users_repository

import (
	"context"
	"fmt"
	"time"
)

func (r *UsersRepository) UpdateAvatarURL(
	ctx context.Context,
	userID int,
	avatarURL string,
) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	query := `
	UPDATE debatesApp.users
	SET
		avatar_url=$1,
		updated_at=$2,
		version=version+1
	WHERE id=$3;
	`

	db := r.dbFromContext(ctx)

	_, err := db.Exec(
		ctx,
		query,
		avatarURL,
		time.Now(),
		userID,
	)
	if err != nil {
		return fmt.Errorf("update avatar url: %w", err)
	}

	return nil
}
