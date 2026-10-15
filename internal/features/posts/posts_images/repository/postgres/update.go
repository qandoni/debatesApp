package post_images_repository

import (
	"context"
	"fmt"
)

const markActiveByIDsQuery = `
UPDATE debatesApp.post_images
SET status = 'ACTIVE'
WHERE id = ANY($1) AND status = 'PENDING';
`

func (r *PostImagesRepository) MarkActiveByIDs(
	ctx context.Context,
	imageIDs []int,
) error {
	if len(imageIDs) == 0 {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	db := r.dbFromContext(ctx)
	if _, err := db.Exec(ctx, markActiveByIDsQuery, imageIDs); err != nil {
		return fmt.Errorf("mark post_images active: %w", err)
	}
	return nil
}
