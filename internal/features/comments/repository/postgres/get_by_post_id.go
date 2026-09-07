package comments_repository

import (
	"context"
	"fmt"

	"github.com/qandoni/debatesApp/internal/core/domain"
)

func (r *CommentsRepository) GetByPostID(
	ctx context.Context,
	postID int,
	limit *int,
	offset *int,
) ([]domain.Comment, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	query := `
		SELECT
			id,
			version,
			post_id,
			parent_comment_id,
			author_id,
			debate_side_id,
			content,
			created_at,
			updated_at
		FROM debatesApp.comments
		WHERE post_id = $1
		ORDER BY created_at ASC, id ASC
		LIMIT $2
		OFFSET $3;
	`

	db := r.dbFromContext(ctx)

	rows, err := db.Query(
		ctx,
		query,
		postID,
		limit,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"select comments: %w",
			err,
		)
	}
	defer rows.Close()

	comments := make([]domain.Comment, 0)

	for rows.Next() {
		var comment domain.Comment

		err := rows.Scan(
			&comment.ID,
			&comment.Version,
			&comment.PostID,
			&comment.ParentCommentID,
			&comment.AuthorID,
			&comment.DebateSideID,
			&comment.Content,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"scan comment: %w",
				err,
			)
		}

		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate comments: %w",
			err,
		)
	}

	return comments, nil
}
