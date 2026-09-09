package statistics_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/qandoni/debatesApp/internal/core/domain"
)

func (r *DebateStatisticsRepository) getTopArgument(
	ctx context.Context,
	postID int,
	sideID int,
) (*domain.TopArgument, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	db := r.dbFromContext(ctx)

	query := `
		SELECT
			c.id,
			c.author_id,
			c.content,
			COALESCE(AVG(cr.score), 0) AS average_rating,
			COUNT(cr.id) AS ratings_count
		FROM debatesapp.comments c
		LEFT JOIN debatesapp.comment_ratings cr
			ON cr.comment_id = c.id
		WHERE c.post_id = $1
		  AND c.debate_side_id = $2
		  AND c.parent_comment_id IS NULL
		GROUP BY
			c.id,
			c.author_id,
			c.content,
			c.created_at
		ORDER BY
			average_rating DESC,
			ratings_count DESC,
			c.created_at ASC
		LIMIT 1
	`

	var argument domain.TopArgument

	err := db.QueryRow(ctx, query, postID, sideID).Scan(
		&argument.ID,
		&argument.AuthorID,
		&argument.Content,
		&argument.AverageRating,
		&argument.RatingsCount,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("get top argument: %w", err)
	}

	return &argument, nil
}
