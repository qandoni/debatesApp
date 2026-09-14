package posts_service

import (
	"context"
	"fmt"

	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
)

func (s *PostsService) DeletePost(
	ctx context.Context,
	userID int,
	postID int,
) error {
	post, err := s.postsRepository.GetPost(ctx, postID)
	if err != nil {
		return fmt.Errorf("get post: %w", err)
	}

	if post.AuthorID != userID {
		return core_errors.ErrAccessForbidden
	}

	images, err := s.imagesService.GetByPostID(ctx, postID)
	if err != nil {
		return fmt.Errorf("get post images: %w", err)
	}

	if err := s.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		if err := s.imagesService.DeleteRecordsByPostID(ctx, postID); err != nil {
			return err
		}

		if err := s.postsRepository.DeletePost(ctx, userID, postID); err != nil {
			return fmt.Errorf("delete post: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("delete post transaction: %w", err)
	}
	if err := s.imagesService.DeleteFromStorage(ctx, images); err != nil {
		return fmt.Errorf("delete post images from storage: %w", err)
	}

	return nil
}
