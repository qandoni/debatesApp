package comments_ratings_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/qandoni/debatesApp/internal/core/domain"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
)

const getByCommentAndUserQuery = `
SELECT
	id,
	version,
	comment_id,
	user_id,
	score,
	created_at,
	updated_at
FROM debatesApp.comment_ratings
WHERE comment_id = $1
  AND user_id = $2
`

func (r *CommentRatingsRepository) GetByCommentAndUser(
	ctx context.Context,
	commentID int,
	userID int,
) (domain.CommentRating, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	db := r.dbFromContext(ctx)

	var rating domain.CommentRating

	err := db.QueryRow(
		ctx,
		getByCommentAndUserQuery,
		commentID,
		userID,
	).Scan(
		&rating.ID,
		&rating.Version,
		&rating.CommentID,
		&rating.UserID,
		&rating.Rating,
		&rating.CreatedAt,
		&rating.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.CommentRating{}, core_errors.ErrNotFound
		}

		return domain.CommentRating{}, fmt.Errorf(
			"get comment rating: %w",
			err,
		)
	}

	return rating, nil
}
