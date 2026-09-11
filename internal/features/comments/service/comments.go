package comments_service

import (
	"context"
	"errors"
	"fmt"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_enum "github.com/qandoni/debatesApp/internal/core/enum"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
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

	return createdComment, nil

}

func (s *CommentsService) GetByPostID(
	ctx context.Context,
	postID int,
	limit *int,
	offset *int,
) ([]domain.Comment, error) {
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
	return comments, nil
}

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
