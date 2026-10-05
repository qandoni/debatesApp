package comment_ratings_service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
	core_realtime "github.com/qandoni/debatesApp/internal/core/realtime"
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

	var (
		result        domain.CommentRating
		commentPostID int
		shouldPublish bool
		isNew         bool
	)

	if err := s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		comment, err := s.commentsRepository.GetByID(
			txCtx,
			commentID,
		)
		if err != nil {
			return fmt.Errorf(
				"get comment: %w",
				err,
			)
		}

		if comment.ParentCommentID != nil {
			return fmt.Errorf(
				"only root arguments can be rated: %w",
				core_errors.ErrInvalidArgument,
			)
		}

		if comment.DebateSideID == nil {
			return fmt.Errorf(
				"only arguments can be rated: %w",
				core_errors.ErrInvalidArgument,
			)
		}

		// Блокировка строки дебата сериализует проверку статуса
		// с параллельным FinishDebate: рейтинг не применяется
		// после финиша.
		debate, err := s.debatesRepository.GetByPostIDForUpdate(
			txCtx,
			comment.PostID,
		)
		if err != nil {
			return fmt.Errorf(
				"get debate: %w",
				err,
			)
		}

		if debate.Status != "OPEN" {
			return fmt.Errorf(
				"cannot rate argument after debate is finished: %w",
				core_errors.ErrAccessForbidden,
			)
		}

		commentPostID = comment.PostID

		existingRating, err := s.commentRatingsRepository.GetByCommentAndUser(
			txCtx,
			commentID,
			userID,
		)

		if err != nil {
			if !errors.Is(err, core_errors.ErrNotFound) {
				return fmt.Errorf(
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
				txCtx,
				rating,
			)
			if err != nil {
				return fmt.Errorf(
					"create comment rating: %w",
					err,
				)
			}

			result = createdRating
			isNew = true
			shouldPublish = true
			return nil
		}

		if existingRating.Rating == score {
			result = existingRating
			shouldPublish = false
			return nil
		}

		now := time.Now()

		existingRating.Rating = score
		existingRating.UpdatedAt = &now

		updatedRating, err := s.commentRatingsRepository.UpdateCommentRating(
			txCtx,
			existingRating,
		)
		if err != nil {
			return fmt.Errorf(
				"update comment rating: %w",
				err,
			)
		}

		result = updatedRating
		isNew = false
		shouldPublish = true
		return nil
	}); err != nil {
		return domain.CommentRating{}, err
	}

	if shouldPublish {
		if err := s.publishEvent(
			commentPostID,
			core_realtime.EventArgumentRatingUpdated,
			core_realtime.NewRatingData(result, commentPostID, isNew),
		); err != nil {
			return domain.CommentRating{}, fmt.Errorf("publish argument rating event: %w", err)
		}
	}

	return result, nil
}
