package comments_repository

import (
	"context"
	"fmt"

	"github.com/qandoni/debatesApp/internal/core/domain"
)

func (r *CommentsRepository) CreateComment(
	ctx context.Context,
	comment domain.Comment,
) (domain.Comment, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	query := `
	INSERT INTO debatesApp.comments(
		post_id,
		parent_comment_id,
		author_id,
		debate_side_id,
		content
	)
		VALUES($1, $2, $3, $4, $5)
		RETURNING
			id,
			version,
			post_id,
			parent_comment_id,
			author_id,
			debate_side_id,
			content,
			created_at,
			updated_at;
	`

	db := r.dbFromContext(ctx)
	var result domain.Comment
	err := db.QueryRow(
		ctx,
		query,
		comment.PostID,
		comment.ParentCommentID,
		comment.AuthorID,
		comment.DebateSideID,
		comment.Content,
	).Scan(
		&result.ID,
		&result.Version,
		&result.PostID,
		&result.ParentCommentID,
		&result.AuthorID,
		&result.DebateSideID,
		&result.Content,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		return domain.Comment{}, fmt.Errorf("insert and scan comment: %w", err)
	}
	return result, nil
}
