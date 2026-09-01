package posts_repository

import (
	"context"
	"fmt"
	"time"

	"github.com/qandoni/debatesApp/internal/core/domain"
)

func (r *PostsRepository) CreatePost(
	ctx context.Context,
	post domain.Post,
) (domain.Post, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	query := `INSERT INTO debatesApp.posts(author_id, content, is_debate, created_at)
	VALUES($1, $2, $3, $4)
	RETURNING id, version, author_id, content, is_debate, created_at, updated_at, deleted_at;
	`

	db := r.dbFromContext(ctx)
	row := db.QueryRow(
		ctx,
		query,
		post.AuthorID,
		post.Content,
		post.IsDebate,
		time.Now(),
	)
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
