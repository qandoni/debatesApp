package images_service

import (
	"context"
	"fmt"

	"github.com/qandoni/debatesApp/internal/core/domain"
)

func (s *ImagesService) GetByPostIDs(
	ctx context.Context,
	postIDs []int,
) (map[int][]domain.PostImage, error) {
	images, err := s.postImagesRepository.GetByPostIDs(ctx, postIDs)
	if err != nil {
		return nil, fmt.Errorf("get images by post ids: %w", err)
	}

	return images, nil
}
