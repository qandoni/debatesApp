package comments_service

import (
	"context"
	"fmt"

	"github.com/qandoni/debatesApp/internal/core/domain"
)

func (s *CommentsService) CreateArgument(
	ctx context.Context,
	userID int,
	postID int,
	debateSideID int,
	content string,
) (domain.Comment, error) {
	if err := s.validateArgumentCreation(ctx, userID, postID, debateSideID); err != nil {
		return domain.Comment{}, err
	}

	argument := domain.NewCommentUninitialized(
		postID,
		userID,
		nil,
		&debateSideID,
		content,
	)

	createdArgument, err := s.commentsRepository.CreateComment(ctx, argument)
	if err != nil {
		return domain.Comment{}, fmt.Errorf("create argument: %w", err)
	}
	return createdArgument, nil
}

func (s *CommentsService) GetArgumentsWithReplies(
	ctx context.Context,
	postID int,
	limit *int,
	offset *int,
) ([]domain.CommentWithRating, []domain.Comment, error) {
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
