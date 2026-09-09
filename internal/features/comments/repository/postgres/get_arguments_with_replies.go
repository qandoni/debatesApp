package comments_repository

import (
	"context"
	"fmt"

	"github.com/qandoni/debatesApp/internal/core/domain"
)

func (r *CommentsRepository) GetArgumentsWithReplies(
	ctx context.Context,
	postID int,
	limit *int,
	offset *int,
) ([]domain.CommentWithRating, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	db := r.dbFromContext(ctx)

	query := `
        SELECT
            c.id,
            c.version,
            c.post_id,
            c.parent_comment_id,
            c.author_id,
            c.debate_side_id,
            c.content,
            c.author_liked,
            c.created_at,
            c.updated_at,
            COALESCE(AVG(cr.score), 0),
            COUNT(cr.id)
        FROM debatesapp.comments c
        LEFT JOIN debatesapp.comment_ratings cr
            ON cr.comment_id = c.id
        WHERE c.post_id = $1
          AND c.parent_comment_id IS NULL
          AND c.debate_side_id IS NOT NULL
        GROUP BY c.id
        ORDER BY
            COALESCE(AVG(cr.score), 0) DESC,
            COUNT(cr.id) DESC,
            c.created_at ASC
        LIMIT $2
        OFFSET $3
    `

	rows, err := db.Query(ctx, query, postID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query arguments: %w", err)
	}
	defer rows.Close()

	arguments := make([]domain.CommentWithRating, 0)

	for rows.Next() {
		var argument domain.CommentWithRating

		err := rows.Scan(
			&argument.Comment.ID,
			&argument.Comment.Version,
			&argument.Comment.PostID,
			&argument.Comment.ParentCommentID,
			&argument.Comment.AuthorID,
			&argument.Comment.DebateSideID,
			&argument.Comment.Content,
			&argument.Comment.AuthorLiked,
			&argument.Comment.CreatedAt,
			&argument.Comment.UpdatedAt,
			&argument.AverageRating,
			&argument.RatingsCount,
		)
		if err != nil {
			return nil, fmt.Errorf("scan argument: %w", err)
		}

		arguments = append(arguments, argument)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate arguments: %w", err)
	}

	return arguments, nil

}
