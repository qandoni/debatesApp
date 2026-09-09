package comments_service

import (
	"context"
	"fmt"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
)

func (s *CommentsService) GetArgumentsWithReplies(
	ctx context.Context,
	postID int,
	limit *int,
	offset *int,
) ([]domain.CommentWithRating, []domain.Comment, error) {
	if limit != nil && *limit < 0 {
		return nil, nil, fmt.Errorf(
			"limit must be non-negative: %w",
			core_errors.ErrInvalidArgument,
		)
	}
	if offset != nil && *offset < 0 {
		return nil, nil, fmt.Errorf(
			"offset must be non-negative: %w",
			core_errors.ErrInvalidArgument,
		)
	}

	arguments, err := s.commentsRepository.GetArgumentsWithReplies(
		ctx,
		postID,
		limit,
		offset,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("get arguments: %w", err)
	}

	comments, err := s.commentsRepository.GetByPostIDWithoutPagination(
		ctx,
		postID,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("get comments: %w", err)
	}

	return arguments, comments, nil
}
