package images_service

import (
	"context"
	"fmt"

	"github.com/qandoni/debatesApp/internal/core/domain"
)

func (s *ImagesService) GetByPostID(
	ctx context.Context,
	postID int,
) ([]domain.PostImage, error) {
	_, err := s.postsRepository.GetPost(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("get post: %w", err)
	}
	images, err := s.postImagesRepository.GetByPostID(ctx, postID)
	if err != nil {
		return []domain.PostImage{}, fmt.Errorf("get post images: %w", err)
	}
	return images, nil
}
