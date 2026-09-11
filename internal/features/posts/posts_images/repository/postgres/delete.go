package post_images_repository

import (
	"context"
	"fmt"
)

const deleteByPostIDQuery = `
DELETE FROM debatesApp.post_images
WHERE post_id=$1;
`

func (r *PostImagesRepository) DeleteByPostID(
	ctx context.Context,
	postID int,
) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	db := r.dbFromContext(ctx)
	_, err := db.Exec(ctx, deleteByPostIDQuery, postID)
	if err != nil {
		return fmt.Errorf("exec query: %w", err)
	}
	return nil
}
