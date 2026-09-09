package comment_ratings_service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
)

func (s *CommentRatingsService) Rate(
	ctx context.Context,
	userID int,
	commentID int,
	score int,
) (domain.CommentRating, error) {

	if score < 1 || score > 5 {
		return domain.CommentRating{}, fmt.Errorf(
			"rating must be between 1 and 5: %w",
			core_errors.ErrInvalidArgument,
		)
	}

	comment, err := s.commentsRepository.GetByID(
		ctx,
		commentID,
	)
	if err != nil {
		return domain.CommentRating{}, fmt.Errorf(
			"get comment: %w",
			err,
		)
	}

	if comment.ParentCommentID != nil {
		return domain.CommentRating{}, fmt.Errorf(
			"only root arguments can be rated: %w",
			core_errors.ErrInvalidArgument,
		)
	}

	if comment.DebateSideID == nil {
		return domain.CommentRating{}, fmt.Errorf(
			"only arguments can be rated: %w",
			core_errors.ErrInvalidArgument,
		)
	}

	debate, err := s.debatesRepository.GetByPostID(
		ctx,
		comment.PostID,
	)
	if err != nil {
		return domain.CommentRating{}, fmt.Errorf(
			"get debate: %w",
			err,
		)
	}

	if debate.Status != "OPEN" {
		return domain.CommentRating{}, fmt.Errorf(
			"cannot rate argument after debate is finished: %w",
			core_errors.ErrAccessForbidden,
		)
	}

	existingRating, err := s.commentRatingsRepository.GetByCommentAndUser(
		ctx,
		commentID,
		userID,
	)

	if err != nil {
		if !errors.Is(err, core_errors.ErrNotFound) {
			return domain.CommentRating{}, fmt.Errorf(
				"get existing rating: %w",
				err,
			)
		}

		rating := domain.NewCommentRating(
			0,
			1,
			commentID,
			userID,
			score,
			time.Now(),
			nil,
		)

		createdRating, err := s.commentRatingsRepository.CreateCommentRating(
			ctx,
			rating,
		)
		if err != nil {
			return domain.CommentRating{}, fmt.Errorf(
				"create comment rating: %w",
				err,
			)
		}

		return createdRating, nil
	}

	if existingRating.Rating == score {
		return existingRating, nil
	}

	now := time.Now()

	existingRating.Rating = score
	existingRating.UpdatedAt = &now

	updatedRating, err := s.commentRatingsRepository.UpdateCommentRating(
		ctx,
		existingRating,
	)
	if err != nil {
		return domain.CommentRating{}, fmt.Errorf(
			"update comment rating: %w",
			err,
		)
	}

	return updatedRating, nil
}
