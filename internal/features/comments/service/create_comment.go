package comments_service

import (
	"context"
	"errors"
	"fmt"

	"github.com/qandoni/debatesApp/internal/core/domain"
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
