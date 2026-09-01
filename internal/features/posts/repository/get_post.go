package posts_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
	core_postgres_pool "github.com/qandoni/debatesApp/internal/core/repository/postgres/pool"
)

func (r *PostsRepository) GetPost(
	ctx context.Context,
	postID int,
) (domain.Post, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	query := `
	SELECT *
	FROM debatesApp.posts
	WHERE id=$1
	AND deleted_at IS null;
	`
	db := r.dbFromContext(ctx)
	row := db.QueryRow(ctx, query, postID)
	var postModel PostModel

	err := row.Scan(
		&postModel.ID,
		&postModel.Version,
		&postModel.AuthorID,
		&postModel.Content,
		&postModel.IsDebate,
		&postModel.CreatedAt,
		&postModel.UpdatedAt,
		&postModel.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return domain.Post{}, fmt.Errorf("post with id='%d': %w", postID, core_errors.ErrNotFound)
		}
		return domain.Post{}, fmt.Errorf("scan error: %w", err)
	}

	postDomain := domain.NewPost(
		postModel.ID,
		postModel.Version,
		postModel.AuthorID,
		postModel.Content,
		postModel.IsDebate,
		postModel.CreatedAt,
		postModel.UpdatedAt,
		postModel.DeletedAt,
	)
	return postDomain, nil
}
