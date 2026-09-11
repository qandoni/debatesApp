package comments_ratings_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/qandoni/debatesApp/internal/core/domain"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
)

const updateCommentRatingQuery = `
UPDATE debatesApp.comment_ratings
SET
	score = $1,
	version = version + 1,
	updated_at = $2
WHERE id = $3
  AND version = $4
RETURNING
	id,
	version,
	comment_id,
	user_id,
	score,
	created_at,
	updated_at
`

func (r *CommentRatingsRepository) UpdateCommentRating(
	ctx context.Context,
	rating domain.CommentRating,
) (domain.CommentRating, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	db := r.dbFromContext(ctx)

	var result domain.CommentRating

	err := db.QueryRow(
		ctx,
		updateCommentRatingQuery,
		rating.Rating,
		rating.UpdatedAt,
		rating.ID,
		rating.Version,
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
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.CommentRating{}, core_errors.ErrConflict
		}

		return domain.CommentRating{}, fmt.Errorf(
			"update comment rating: %w",
			err,
		)
	}

	return result, nil
}
