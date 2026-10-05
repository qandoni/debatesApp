package comments_service

import (
	"context"
	"fmt"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_realtime "github.com/qandoni/debatesApp/internal/core/realtime"
)

func (s *CommentsService) CreateArgument(
	ctx context.Context,
	userID int,
	postID int,
	debateSideID int,
	content string,
) (domain.Comment, error) {
	var createdArgument domain.Comment

	if err := s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := s.validateArgumentCreation(txCtx, userID, postID, debateSideID); err != nil {
			return err
		}

		argument := domain.NewCommentUninitialized(
			postID,
			userID,
			nil,
			&debateSideID,
			content,
		)

		var err error
		createdArgument, err = s.commentsRepository.CreateComment(txCtx, argument)
		if err != nil {
			return fmt.Errorf("create argument: %w", err)
		}
		return nil
	}); err != nil {
		return domain.Comment{}, err
	}

	if err := s.publishEvent(
		createdArgument.PostID,
		core_realtime.EventArgumentCreated,
		core_realtime.NewCommentCreatedData(createdArgument),
	); err != nil {
		return domain.Comment{}, fmt.Errorf("publish argument created event: %w", err)
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
