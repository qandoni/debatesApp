package comments_repository

import (
	"context"
	"fmt"
)

func (r *CommentsRepository) HasUserArgumentInPost(
	ctx context.Context,
	userID int,
	postID int,
) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	query := `
		SELECT EXISTS (
			SELECT 1
			FROM debatesApp.comments
			WHERE author_id = $1
			  AND post_id = $2
			  AND debate_side_id IS NOT NULL
			  AND parent_comment_id IS NULL
		);
	`

	db := r.dbFromContext(ctx)

	var exists bool

	err := db.QueryRow(
		ctx,
		query,
		userID,
		postID,
	).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf(
			"check user argument: %w",
			err,
		)
	}

	return exists, nil
}
