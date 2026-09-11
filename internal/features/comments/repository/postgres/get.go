package comments_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/qandoni/debatesApp/internal/core/domain"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
)

const getByIDQuery = `
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

const getByPostIDQuery = `
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

const getByPostIDWithoutPaginationQuery = `
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

const getArgumentsWithRepliesQuery = `
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

const hasUserArgumentInPostQuery = `
SELECT EXISTS (
	SELECT 1
	FROM debatesApp.comments
	WHERE author_id = $1
	  AND post_id = $2
	  AND debate_side_id IS NOT NULL
	  AND parent_comment_id IS NULL
);
`

func (r *CommentsRepository) GetByID(
	ctx context.Context,
	commentID int,
) (domain.Comment, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	db := r.dbFromContext(ctx)

	var comment domain.Comment

	err := db.QueryRow(
		ctx,
		getByIDQuery,
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

func (r *CommentsRepository) GetByPostID(
	ctx context.Context,
	postID int,
	limit *int,
	offset *int,
) ([]domain.Comment, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	db := r.dbFromContext(ctx)

	rows, err := db.Query(
		ctx,
		getByPostIDQuery,
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

func (r *CommentsRepository) GetByPostIDWithoutPagination(
	ctx context.Context,
	postID int,
) ([]domain.Comment, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	db := r.dbFromContext(ctx)

	rows, err := db.Query(ctx, getByPostIDWithoutPaginationQuery, postID)
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

func (r *CommentsRepository) GetArgumentsWithReplies(
	ctx context.Context,
	postID int,
	limit *int,
	offset *int,
) ([]domain.CommentWithRating, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	db := r.dbFromContext(ctx)

	rows, err := db.Query(ctx, getArgumentsWithRepliesQuery, postID, limit, offset)
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

func (r *CommentsRepository) HasUserArgumentInPost(
	ctx context.Context,
	userID int,
	postID int,
) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	db := r.dbFromContext(ctx)

	var exists bool

	err := db.QueryRow(
		ctx,
		hasUserArgumentInPostQuery,
		userID,
		postID,
	).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf(
			"check user argument: %w",
			err,
		)
	}

	return exists, nil
}
