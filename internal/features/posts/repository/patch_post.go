package posts_repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
	core_postgres_pool "github.com/qandoni/debatesApp/internal/core/repository/postgres/pool"
)

func (r *PostsRepository) PatchPost(
	ctx context.Context,
	userID int,
	postID int,
	postPatch domain.PostPatch,
) (domain.Post, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	query := `
	UPDATE debatesApp.posts
	SET 
		content=$1,
		version=version+1,
		updated_at=$2
	WHERE id=$3
	AND author_id=$4
	AND deleted_at IS null
	RETURNING
		id,
		version,
		author_id,
		content,
		is_debate,
		created_at,
		updated_at,
		deleted_at;
	`

	db := r.dbFromContext(ctx)
	row := db.QueryRow(
		ctx,
		query,
		*postPatch.Content.Value,
		time.Now(),
		postID,
		userID,
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
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return domain.Post{}, fmt.Errorf(
				"post with id='%d' not found: %w",
				postID,
				core_errors.ErrNotFound,
			)
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
