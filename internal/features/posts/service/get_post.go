package posts_service

import (
	"context"
	"fmt"

	"github.com/qandoni/debatesApp/internal/core/domain"
)

func (s *PostsService) GetPost(
	ctx context.Context,
	postID int,
) (domain.Post, error) {
	post, err := s.postsRepository.GetPost(ctx, postID)
	if err != nil {
		return domain.Post{}, fmt.Errorf("get post from repository: %w", err)
	}

	images, err := s.imagesService.GetByPostID(ctx, postID)
	if err != nil {
		return domain.Post{}, fmt.Errorf("get post images: %w", err)
	}

	post.Images = images

	return post, nil
}
