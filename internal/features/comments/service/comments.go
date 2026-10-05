package comments_service

import (
	"context"
	"errors"
	"fmt"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_enum "github.com/qandoni/debatesApp/internal/core/enum"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
	core_realtime "github.com/qandoni/debatesApp/internal/core/realtime"
)

func (s *CommentsService) CreateComment(
	ctx context.Context,
	userID int,
	postID int,
	parentCommentID *int,
	content string,
) (domain.Comment, error) {
	_, err := s.postsRepository.GetPost(ctx, postID)
	if err != nil {
		return domain.Comment{}, fmt.Errorf(
			"get post: %w",
			err,
		)
	}

	var debateSideID *int

	if parentCommentID == nil {
		_, err := s.debatesRepository.GetByPostID(ctx, postID)
		if err == nil {
			return domain.Comment{}, fmt.Errorf("this post is a debate and need a parent comment id to work: %w", core_errors.ErrConflict)
		}

		if !errors.Is(err, core_errors.ErrNotFound) {
			return domain.Comment{}, fmt.Errorf(
				"check post debate: %w",
				err,
			)
		}
	}

	if parentCommentID != nil {
		parentComment, err := s.commentsRepository.GetByID(
			ctx,
			*parentCommentID,
		)
		if err != nil {
			return domain.Comment{}, fmt.Errorf(
				"get parent comment: %w",
				err,
			)
		}

		if parentComment.PostID != postID {
			return domain.Comment{}, fmt.Errorf("parent comment post id != postID: %w", core_errors.ErrConflict)
		}

		debateSideID = parentComment.DebateSideID
	}

	comment := domain.Comment{
		PostID:          postID,
		ParentCommentID: parentCommentID,
		AuthorID:        userID,
		DebateSideID:    debateSideID,
		Content:         content,
	}

	createdComment, err := s.commentsRepository.CreateComment(
		ctx,
		comment,
	)
	if err != nil {
		return domain.Comment{}, fmt.Errorf(
			"create comment: %w",
			err,
		)
	}

	if err := s.publishEvent(
		createdComment.PostID,
		core_realtime.EventCommentCreated,
		core_realtime.NewCommentCreatedData(createdComment),
	); err != nil {
		return domain.Comment{}, fmt.Errorf("publish comment created event: %w", err)
	}

	return createdComment, nil

}

func (s *CommentsService) GetByPostID(
	ctx context.Context,
	postID int,
	limit *int,
	offset *int,
) ([]domain.Comment, error) {
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

func (s *CommentsService) UpdateComment(
	ctx context.Context,
	userID int,
	commentID int,
	content string,
) (domain.Comment, error) {
	var updatedComment domain.Comment

	if err := s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		comment, err := s.commentsRepository.GetByID(
			txCtx, commentID,
		)
		if err != nil {
			return fmt.Errorf("get by comment id: %w", err)
		}

		if comment.AuthorID != userID {
			return core_errors.ErrAccessForbidden
		}

		if comment.DebateSideID != nil {
			// Блокировка строки дебата сериализует проверку статуса
			// с параллельным FinishDebate: статус не меняется между
			// проверкой и записью.
			debate, err := s.debatesRepository.GetByPostIDForUpdate(txCtx, comment.PostID)
			if err != nil {
				return fmt.Errorf("get debate: %w", err)
			}
			if debate.Status != core_enum.DebateStatusOpen {
				return core_errors.ErrConflict
			}
		}

		comment.Content = content

		updatedComment, err = s.commentsRepository.UpdateComment(txCtx, comment)
		if err != nil {
			return fmt.Errorf("update comment: %w", err)
		}
		return nil
	}); err != nil {
		return domain.Comment{}, err
	}

	if err := s.publishEvent(
		updatedComment.PostID,
		core_realtime.EventCommentUpdated,
		core_realtime.NewCommentUpdatedData(updatedComment),
	); err != nil {
		return domain.Comment{}, fmt.Errorf("publish comment updated event: %w", err)
	}

	return updatedComment, nil

}

func (s *CommentsService) SetAuthorLike(
	ctx context.Context,
	userID int,
	commentID int,
	liked bool,
) (domain.Comment, error) {
	var (
		updatedComment domain.Comment
		authorID       int
	)

	if err := s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		comment, err := s.commentsRepository.GetByID(txCtx, commentID)
		if err != nil {
			return fmt.Errorf("get comment: %w", err)
		}

		if comment.DebateSideID == nil {
			return fmt.Errorf(
				"author like is available only for arguments: %w",
				core_errors.ErrInvalidArgument,
			)
		}

		if comment.ParentCommentID != nil {
			return fmt.Errorf(
				"author like is available only for root arguments: %w",
				core_errors.ErrInvalidArgument,
			)
		}

		debate, err := s.debatesRepository.GetByPostIDForUpdate(txCtx, comment.PostID)
		if err != nil {
			return fmt.Errorf("get debate: %w", err)
		}

		if debate.Status != "OPEN" {
			return fmt.Errorf(
				"cannot change author like after debate finished: %w",
				core_errors.ErrConflict,
			)
		}

		authorID, err = s.debatesRepository.GetAuthorID(txCtx, debate.ID)
		if err != nil {
			return fmt.Errorf("get debate author: %w", err)
		}

		if authorID != userID {
			return fmt.Errorf(
				"only debate author can set author like: %w",
				core_errors.ErrAccessForbidden,
			)
		}

		updatedComment, err = s.commentsRepository.SetAuthorLike(
			txCtx,
			commentID,
			liked,
		)
		if err != nil {
			return fmt.Errorf("set author like: %w", err)
		}
		return nil
	}); err != nil {
		return domain.Comment{}, err
	}

	if err := s.publishEvent(
		updatedComment.PostID,
		core_realtime.EventArgumentAuthorLiked,
		core_realtime.NewAuthorLikeData(updatedComment, authorID),
	); err != nil {
		return domain.Comment{}, fmt.Errorf("publish argument author like event: %w", err)
	}

	return updatedComment, nil
}
