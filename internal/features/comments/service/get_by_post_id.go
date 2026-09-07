package comments_service

import (
	"context"
	"fmt"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
)

func (s *CommentsService) GetByPostID(
	ctx context.Context,
	postID int,
	limit *int,
	offset *int,
) ([]domain.Comment, error) {
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
	_, err := s.postsRepository.GetPost(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf(
			"get post: %w", err,
		)
	}
	comments, err := s.commentsRepository.GetByPostID(ctx, postID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get comments: %w", err)
	}
	return comments, nil
}
