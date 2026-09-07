package comments_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/qandoni/debatesApp/internal/core/domain"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
)

func (r *CommentsRepository) GetByID(
	ctx context.Context,
	commentID int,
) (domain.Comment, error) {
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
		WHERE id = $1;
	`

	db := r.dbFromContext(ctx)

	var comment domain.Comment

	err := db.QueryRow(
		ctx,
		query,
		commentID,
	).Scan(
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
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Comment{}, core_errors.ErrNotFound
		}

		return domain.Comment{}, fmt.Errorf(
			"select and scan comment: %w",
			err,
		)
	}

	return comment, nil
}
