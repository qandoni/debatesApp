package post_images_repository

import (
	"context"
	"fmt"

	"github.com/qandoni/debatesApp/internal/core/domain"
)

func (r *PostImagesRepository) CreatePostImage(
	ctx context.Context,
	image domain.PostImage,
) (domain.PostImage, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	query := `
	INSERT INTO debatesApp.post_images(post_id, image_url, display_order, created_at)
	VALUES($1, $2, $3, $4)
	RETURNING
		id,
		post_id,
		image_url,
		display_order,
		created_at;
	`
	db := r.dbFromContext(ctx)
	row := db.QueryRow(ctx, query, image.PostID, image.ImageURL, image.DisplayOrder, image.CreatedAt)
	var postImagesModel PostImagesModel
	err := row.Scan(
		&postImagesModel.ID,
		&postImagesModel.PostID,
		&postImagesModel.ImageURL,
		&postImagesModel.DisplayOrder,
		&postImagesModel.CreatedAt,
	)
	if err != nil {
		return domain.PostImage{}, fmt.Errorf(
			"create post image: %w",
			err,
		)
	}
	postImageDomain := domain.NewPostImage(
		postImagesModel.ID,
		postImagesModel.PostID,
		postImagesModel.ImageURL,
		postImagesModel.DisplayOrder,
		postImagesModel.CreatedAt,
	)
	return postImageDomain, nil
}
