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

const deletePendingByIDsQuery = `
DELETE FROM debatesApp.post_images
WHERE id = ANY($1) AND status = 'PENDING';
`

func (r *PostImagesRepository) DeletePendingByIDs(
	ctx context.Context,
	imageIDs []int,
) error {
	if len(imageIDs) == 0 {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	db := r.dbFromContext(ctx)
	if _, err := db.Exec(ctx, deletePendingByIDsQuery, imageIDs); err != nil {
		return fmt.Errorf("delete pending post_images: %w", err)
	}
	return nil
}
