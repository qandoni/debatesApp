package posts_service

import (
	"context"
	"fmt"

	"github.com/qandoni/debatesApp/internal/core/domain"
)

func (s *PostsService) PatchPost(
	ctx context.Context,
	userID int,
	postID int,
	postPatch domain.PostPatch,
) (domain.Post, error) {
	patchedPost, err := s.postsRepository.PatchPost(ctx, userID, postID, postPatch)
	if err != nil {
		return domain.Post{}, fmt.Errorf("patch post: %w", err)
	}
	return patchedPost, nil
}
