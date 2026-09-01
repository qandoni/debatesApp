package post_images_repository

import (
	"context"
	"fmt"

	"github.com/qandoni/debatesApp/internal/core/domain"
)

func (r *PostImagesRepository) GetByPostID(
	ctx context.Context,
	postID int,
) ([]domain.PostImage, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	query := `
	SELECT *
	FROM debatesApp.post_images
	WHERE post_id=$1
	ORDER BY display_order
	`

	db := r.dbFromContext(ctx)
	rows, err := db.Query(ctx, query, postID)
	if err != nil {
		return []domain.PostImage{}, fmt.Errorf("select post_images: %w", err)
	}
	defer rows.Close()

	var postImagesModels []PostImagesModel
	for rows.Next() {
		var postImagesModel PostImagesModel
		err := rows.Scan(
			&postImagesModel.ID,
			&postImagesModel.PostID,
			&postImagesModel.ImageURL,
			&postImagesModel.DisplayOrder,
			&postImagesModel.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan post_images: %w", err)
		}
		postImagesModels = append(postImagesModels, postImagesModel)
	}
	postImagesDomains := postImagesDomainsFromModels(postImagesModels)
	return postImagesDomains, nil
}
