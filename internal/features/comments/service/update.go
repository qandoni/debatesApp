package comments_service

import (
	"context"
	"fmt"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_enum "github.com/qandoni/debatesApp/internal/core/enum"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
)

func (s *CommentsService) UpdateComment(
	ctx context.Context,
	userID int,
	commentID int,
	content string,
) (domain.Comment, error) {
	comment, err := s.commentsRepository.GetByID(
		ctx, commentID,
	)
	if err != nil {
		return domain.Comment{}, fmt.Errorf("get by comment id: %w", err)
	}

	if comment.AuthorID != userID {
		return domain.Comment{}, core_errors.ErrAccessForbidden
	}

	if comment.DebateSideID != nil {
		debate, err := s.debatesRepository.GetByPostID(ctx, comment.PostID)
		if err != nil {
			return domain.Comment{}, fmt.Errorf("get debate: %w", err)
		}
		if debate.Status != core_enum.DebateStatusOpen {
			return domain.Comment{}, core_errors.ErrConflict
		}
	}

	comment.Content = content

	updatedComment, err := s.commentsRepository.UpdateComment(ctx, comment)
	if err != nil {
		return domain.Comment{}, fmt.Errorf("update comment: %w", err)
	}

	return updatedComment, nil

}
