package posts_service

import (
	"context"
	"fmt"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
)

func (s *PostsService) GetPosts(
	ctx context.Context,
	limit *int,
	offset *int,
) ([]domain.Post, error) {
	if limit != nil && *limit < 0 {
		return nil, fmt.Errorf(
			"limit must be non-negative: %w",
			core_errors.ErrInvalidArgument,
		)
	}
	if offset != nil && *offset < 0 {
		return nil, fmt.Errorf(
			"offset must be non-negative: %w",
			core_errors.ErrInvalidArgument,
		)
	}
	posts, err := s.postsRepository.GetPosts(ctx, limit, offset)
	if err != nil {
		return []domain.Post{}, fmt.Errorf("get posts from repository: %w", err)
	}
	for i := range posts {
		images, err := s.imagesService.GetByPostID(ctx, posts[i].ID)
		if err != nil {
			return []domain.Post{}, fmt.Errorf("get images for post with id: '%d': %w", posts[i].ID, err)
		}
		posts[i].Images = images
	}
	// TODO передалать цикл для производительности под GetByPostIDs(
	//     ctx context.Context,
	//     postIDs []int,
	// ) (map[int][]domain.PostImage, error)
	return posts, nil
}
