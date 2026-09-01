package post_images_repository

import (
	"context"
	"fmt"

	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
)

func (r *PostImagesRepository) DeleteByPostID(
	ctx context.Context,
	postID int,
) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	query := `
	DELETE FROM debatesApp.post_images
	WHERE post_id=$1;
	`
	db := r.dbFromContext(ctx)
	cmdTag, err := db.Exec(ctx, query, postID)
	if err != nil {
		return fmt.Errorf("exec query: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("post_images with post_id='%d': %w", postID, core_errors.ErrNotFound)
	}
	return nil
}
