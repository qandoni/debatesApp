package posts_repository

import (
	"context"
	"fmt"
	"time"

	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
)

const deletePostQuery = `
UPDATE debatesApp.posts
SET
	deleted_at=$1
WHERE id=$2
AND author_id=$3;
`

func (r *PostsRepository) DeletePost(
	ctx context.Context,
	userID int,
	postID int,
) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	db := r.dbFromContext(ctx)
	cmdTag, err := db.Exec(ctx, deletePostQuery, time.Now(), postID, userID)
	if err != nil {
		return fmt.Errorf("exec query: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("post with id='%d': %w", postID, core_errors.ErrNotFound)
	}
	return nil
}
