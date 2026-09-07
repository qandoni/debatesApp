package comments_service

import (
	"context"
	"fmt"

	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
	comments_dto "github.com/qandoni/debatesApp/internal/features/comments/transport/http/dto"
)

func (s *CommentsService) GetByPostID(
	ctx context.Context,
	postID int,
	limit *int,
	offset *int,
) ([]comments_dto.CommentDTOResponse, error) {
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
	return comments_dto.BuildCommentTree(comments), nil
}
