package comments_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/qandoni/debatesApp/internal/core/domain"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
)

func (r *CommentsRepository) SetAuthorLike(
	ctx context.Context,
	commentID int,
	liked bool,
) (domain.Comment, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	db := r.dbFromContext(ctx)

	query := `
		UPDATE debatesapp.comments
		SET
			author_liked = $1,
			version = version + 1,
			updated_at = NOW()
		WHERE id = $2
		RETURNING
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
	`

	var comment domain.Comment

	err := db.QueryRow(
		ctx,
		query,
		liked,
		commentID,
	).Scan(
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
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Comment{}, core_errors.ErrNotFound
		}

		return domain.Comment{}, fmt.Errorf("update author like: %w", err)
	}

	return comment, nil
}
