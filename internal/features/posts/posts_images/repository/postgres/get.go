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
SELECT
	id,
	post_id,
	image_url,
	display_order,
	created_at
FROM debatesapp.post_images
WHERE post_id = $1
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

func (r *PostImagesRepository) GetByPostIDs(
	ctx context.Context,
	postIDs []int,
) (map[int][]domain.PostImage, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	if len(postIDs) == 0 {
		return map[int][]domain.PostImage{}, nil
	}

	query := `
		SELECT
			id,
			post_id,
			image_url,
			display_order,
			created_at
		FROM debatesapp.post_images
		WHERE post_id = ANY($1)
		ORDER BY post_id, display_order
	`

	db := r.dbFromContext(ctx)

	rows, err := db.Query(ctx, query, postIDs)
	if err != nil {
		return nil, fmt.Errorf("select post_images: %w", err)
	}
	defer rows.Close()

	postImagesModels := make([]PostImagesModel, 0)

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

		postImagesModels = append(
			postImagesModels,
			postImagesModel,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate post_images: %w", err)
	}

	postImagesDomains := postImagesDomainsFromModels(postImagesModels)

	result := make(map[int][]domain.PostImage, len(postIDs))

	for _, postID := range postIDs {
		result[postID] = []domain.PostImage{}
	}

	for _, image := range postImagesDomains {
		result[image.PostID] = append(
			result[image.PostID],
			image,
		)
	}

	return result, nil
}
