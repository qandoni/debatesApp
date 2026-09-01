package posts_service

import (
	"context"
	"fmt"

	"github.com/qandoni/debatesApp/internal/core/domain"
	posts_contracts "github.com/qandoni/debatesApp/internal/features/posts/contracts"
)

func (s *PostsService) CreatePost(
	ctx context.Context,
	input posts_contracts.CreatePostInput,
) (domain.Post, error) {
	post := domain.NewPostUninitialized(input.AuthorID, input.Content, input.IsDebate)

	post, err := s.postsRepository.CreatePost(ctx, post)
	if err != nil {
		return domain.Post{}, fmt.Errorf("create post in repository: %w", err)
	}
	return post, nil
}
