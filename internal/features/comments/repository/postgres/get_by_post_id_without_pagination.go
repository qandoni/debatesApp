package comments_repository

import (
	"context"
	"fmt"

	"github.com/qandoni/debatesApp/internal/core/domain"
)

func (r *CommentsRepository) GetByPostIDWithoutPagination(
	ctx context.Context,
	postID int,
) ([]domain.Comment, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	db := r.dbFromContext(ctx)

	query := `
        SELECT
            id,
            version,
            post_id,
            parent_comment_id,
            author_id,
            debate_side_id,
            content,
            author_liked,
            created_at,
            updated_at
        FROM debatesapp.comments
        WHERE post_id = $1
        ORDER BY created_at ASC
    `

	rows, err := db.Query(ctx, query, postID)
	if err != nil {
		return nil, fmt.Errorf("query comments by post id: %w", err)
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
			&comment.AuthorLiked,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan comment: %w", err)
		}

		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate comments: %w", err)
	}

	return comments, nil
}
