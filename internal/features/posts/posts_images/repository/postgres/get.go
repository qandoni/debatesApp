package post_images_repository

import (
	"context"
	"fmt"
	"time"

	"github.com/qandoni/debatesApp/internal/core/domain"
)

const getByPostIDQuery = `
SELECT
	id,
	post_id,
	image_url,
	display_order,
	created_at
FROM debatesapp.post_images
WHERE post_id = $1
  AND status = 'ACTIVE'
ORDER BY display_order
`

const getByPostIDsQuery = `
SELECT
	id,
	post_id,
	image_url,
	display_order,
	created_at
FROM debatesapp.post_images
WHERE post_id = ANY($1)
  AND status = 'ACTIVE'
ORDER BY post_id, display_order
`

func (r *PostImagesRepository) GetByPostID(
	ctx context.Context,
	postID int,
) ([]domain.PostImage, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	db := r.dbFromContext(ctx)
	rows, err := db.Query(ctx, getByPostIDQuery, postID)
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

	db := r.dbFromContext(ctx)

	rows, err := db.Query(ctx, getByPostIDsQuery, postIDs)
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

const getPendingQuery = `
SELECT
id,
post_id,
image_url,
display_order,
created_at
FROM debatesapp.post_images
WHERE status = 'PENDING'
`

const getPendingOlderThanQuery = `
SELECT
id,
post_id,
image_url,
display_order,
created_at
FROM debatesapp.post_images
WHERE status = 'PENDING'
  AND created_at < $1
`

const getImagesOfDeletedPostsQuery = `
SELECT
pi.id,
pi.post_id,
pi.image_url,
pi.display_order,
pi.created_at
FROM debatesapp.post_images pi
JOIN debatesapp.posts p ON p.id = pi.post_id
WHERE p.deleted_at IS NOT NULL
`

func (r *PostImagesRepository) GetPending(
	ctx context.Context,
) ([]domain.PostImage, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	db := r.dbFromContext(ctx)
	rows, err := db.Query(ctx, getPendingQuery)
	if err != nil {
		return []domain.PostImage{}, fmt.Errorf("select pending post_images: %w", err)
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
			return nil, fmt.Errorf("scan pending post_images: %w", err)
		}
		postImagesModels = append(postImagesModels, postImagesModel)
	}
	return postImagesDomainsFromModels(postImagesModels), nil
}

func (r *PostImagesRepository) GetPendingOlderThan(
	ctx context.Context,
	staleAfter time.Duration,
) ([]domain.PostImage, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	cutoff := time.Now().Add(-staleAfter)

	db := r.dbFromContext(ctx)
	rows, err := db.Query(ctx, getPendingOlderThanQuery, cutoff)
	if err != nil {
		return []domain.PostImage{}, fmt.Errorf("select stale pending post_images: %w", err)
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
			return nil, fmt.Errorf("scan stale pending post_images: %w", err)
		}
		postImagesModels = append(postImagesModels, postImagesModel)
	}
	return postImagesDomainsFromModels(postImagesModels), nil
}

func (r *PostImagesRepository) GetImagesOfDeletedPosts(
	ctx context.Context,
) ([]domain.PostImage, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	db := r.dbFromContext(ctx)
	rows, err := db.Query(ctx, getImagesOfDeletedPostsQuery)
	if err != nil {
		return []domain.PostImage{}, fmt.Errorf("select images of deleted posts: %w", err)
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
			return nil, fmt.Errorf("scan images of deleted posts: %w", err)
		}
		postImagesModels = append(postImagesModels, postImagesModel)
	}
	return postImagesDomainsFromModels(postImagesModels), nil
}
