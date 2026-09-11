package comments_ratings_repository

import (
	"context"
	"fmt"

	"github.com/qandoni/debatesApp/internal/core/domain"
)

func (r *CommentRatingsRepository) CreateCommentRating(
	ctx context.Context,
	rating domain.CommentRating,
) (domain.CommentRating, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	db := r.dbFromContext(ctx)

	query := `
		INSERT INTO debatesApp.comment_ratings (
			version,
			comment_id,
			user_id,
			score,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING
			id,
			version,
			comment_id,
			user_id,
			score,
			created_at,
			updated_at
	`

	var result domain.CommentRating

	err := db.QueryRow(
		ctx,
		query,
		rating.Version,
		rating.CommentID,
		rating.UserID,
		rating.Rating,
		rating.CreatedAt,
		rating.UpdatedAt,
	).Scan(
		&result.ID,
		&result.Version,
		&result.CommentID,
		&result.UserID,
		&result.Rating,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		return domain.CommentRating{}, fmt.Errorf(
			"create comment rating: %w",
			err,
		)
	}

	return result, nil
}
