package posts_service

import (
	"context"
	"fmt"
)

func (s *PostsService) DeletePost(
	ctx context.Context,
	userID int,
	postID int,
) error {
	if err := s.imagesService.DeleteByPostID(ctx, userID, postID); err != nil {
		return fmt.Errorf("delete post_images: %w", err)
	}
	if err := s.postsRepository.DeletePost(ctx, userID, postID); err != nil {
		return fmt.Errorf("delete post: %w", err)
	}
	return nil
}
