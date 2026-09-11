package comments_service

import (
	"context"
	"fmt"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_enum "github.com/qandoni/debatesApp/internal/core/enum"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
)

func (s *CommentsService) CreateArgument(
	ctx context.Context,
	userID int,
	postID int,
	debateSideID int,
	content string,
) (domain.Comment, error) {
	_, err := s.postsRepository.GetPost(ctx, postID)
	if err != nil {
		return domain.Comment{}, fmt.Errorf(
			"get post: %w",
			err,
		)
	}

	debate, err := s.debatesRepository.GetByPostID(ctx, postID)
	if err != nil {
		return domain.Comment{}, fmt.Errorf("get debate: %w", err)
	}
	if debate.Status != core_enum.DebateStatusOpen {
		return domain.Comment{}, fmt.Errorf("debates are closed: %w", core_errors.ErrConflict)
	}
	sides, err := s.debateSidesRepository.GetByDebateID(ctx, debate.ID)
	if err != nil {
		return domain.Comment{}, fmt.Errorf(
			"get debate sides: %w",
			err,
		)
	}

	var sideExists bool

	for _, side := range sides {
		if side.ID == debateSideID {
			sideExists = true
			break
		}
	}
	if !sideExists {
		return domain.Comment{}, core_errors.ErrNotFound
	}

	vote, err := s.debateVotesRepository.GetByDebateAndUser(ctx, debate.ID, userID)
	if err != nil {
		return domain.Comment{}, fmt.Errorf("get user vote: %w", err)
	}
	if vote.DebateSideID != debateSideID {
		return domain.Comment{}, fmt.Errorf("user voted for other side: '%d': %w", vote.DebateSideID, core_errors.ErrConflict)
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
