package comments_service

import (
	"context"
	"fmt"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
)

func (s *CommentsService) SetAuthorLike(
	ctx context.Context,
	userID int,
	commentID int,
	liked bool,
) (domain.Comment, error) {
	comment, err := s.commentsRepository.GetByID(ctx, commentID)
	if err != nil {
		return domain.Comment{}, fmt.Errorf("get comment: %w", err)
	}

	if comment.DebateSideID == nil {
		return domain.Comment{}, fmt.Errorf(
			"author like is available only for arguments: %w",
			core_errors.ErrInvalidArgument,
		)
	}

	if comment.ParentCommentID != nil {
		return domain.Comment{}, fmt.Errorf(
			"author like is available only for root arguments: %w",
			core_errors.ErrInvalidArgument,
		)
	}

	debate, err := s.debatesRepository.GetByPostID(ctx, comment.PostID)
	if err != nil {
		return domain.Comment{}, fmt.Errorf("get debate: %w", err)
	}

	if debate.Status != "OPEN" {
		return domain.Comment{}, fmt.Errorf(
			"cannot change author like after debate finished: %w",
			core_errors.ErrConflict,
		)
	}

	authorID, err := s.debatesRepository.GetAuthorID(ctx, debate.ID)
	if err != nil {
		return domain.Comment{}, fmt.Errorf("get debate author: %w", err)
	}

	if authorID != userID {
		return domain.Comment{}, fmt.Errorf(
			"only debate author can set author like: %w",
			core_errors.ErrAccessForbidden,
		)
	}

	updatedComment, err := s.commentsRepository.SetAuthorLike(
		ctx,
		commentID,
		liked,
	)
	if err != nil {
		return domain.Comment{}, fmt.Errorf("set author like: %w", err)
	}

	return updatedComment, nil
}
