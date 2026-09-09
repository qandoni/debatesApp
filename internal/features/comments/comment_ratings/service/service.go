package comment_ratings_service

import (
	"context"

	"github.com/qandoni/debatesApp/internal/core/domain"
)

func NewCommentRatingsService(
	commentRatingsRepository CommentRatingsRepository,
	commentsRepository CommentsRepository,
	debatesRepository DebatesRepository,
) *CommentRatingsService {
	return &CommentRatingsService{
		commentRatingsRepository,
		commentsRepository,
		debatesRepository,
	}
}

type CommentRatingsService struct {
	commentRatingsRepository CommentRatingsRepository
	commentsRepository       CommentsRepository
	debatesRepository        DebatesRepository
}

type CommentRatingsRepository interface {
	CreateCommentRating(
		ctx context.Context,
		rating domain.CommentRating,
	) (domain.CommentRating, error)

	GetByCommentAndUser(
		ctx context.Context,
		commentID int,
		userID int,
	) (domain.CommentRating, error)

	UpdateCommentRating(
		ctx context.Context,
		rating domain.CommentRating,
	) (domain.CommentRating, error)
}

type CommentsRepository interface {
	GetByID(
		ctx context.Context,
		commentID int,
	) (domain.Comment, error)
}

type DebatesRepository interface {
	GetByPostID(
		ctx context.Context,
		postID int,
	) (domain.Debate, error)
}
