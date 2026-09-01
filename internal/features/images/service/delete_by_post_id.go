package images_service

import (
	"context"
	"fmt"

	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
)

func (s *ImagesService) DeleteByPostID(
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

	images, err := s.postImagesRepository.GetByPostID(ctx, postID)
	if err != nil {
		return fmt.Errorf("get post images: %w", err)
	}

	for _, image := range images {
		if err := s.storage.Delete(ctx, image.ImageURL); err != nil {
			return fmt.Errorf("delete image %q from storage: %w",
				image.ImageURL,
				err,
			)
		}
	}

	if err := s.postImagesRepository.DeleteByPostID(ctx, postID); err != nil {
		return fmt.Errorf("delete post images: %w", err)
	}
	return nil
}
