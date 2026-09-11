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

func (r *PostsRepository) GetPosts(
	ctx context.Context,
	limit *int,
	offset *int,
) ([]domain.Post, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	query := `
SELECT id, version, author_id, content, is_debate, created_at, updated_at, deleted_at
FROM debatesApp.posts
WHERE deleted_at IS null
ORDER BY id ASC
LIMIT $1
OFFSET $2;
`

	db := r.dbFromContext(ctx)
	rows, err := db.Query(ctx, query, limit, offset)
	if err != nil {
		return []domain.Post{}, fmt.Errorf("select posts: %w", err)
	}
	defer rows.Close()
	var postsModels []PostModel
	for rows.Next() {
		var postModel PostModel
		err := rows.Scan(
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
			return nil, fmt.Errorf("scan posts: %w", err)
		}
		postsModels = append(postsModels, postModel)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("next rows: %w", err)
	}
	postsDomains := postDomainsFromModels(postsModels)
	return postsDomains, nil
}
