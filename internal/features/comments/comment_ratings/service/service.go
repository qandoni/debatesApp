package comment_ratings_service

import (
	"context"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_realtime "github.com/qandoni/debatesApp/internal/core/realtime"
	core_postgres "github.com/qandoni/debatesApp/internal/core/repository/postgres"
)

func NewCommentRatingsService(
	commentRatingsRepository CommentRatingsRepository,
	commentsRepository CommentsRepository,
	debatesRepository DebatesRepository,
	txManager core_postgres.TransactionManager,
	realtimeHub core_realtime.Publisher,
) *CommentRatingsService {
	return &CommentRatingsService{
		commentRatingsRepository,
		commentsRepository,
		debatesRepository,
		txManager,
		realtimeHub,
	}
}

type CommentRatingsService struct {
	commentRatingsRepository CommentRatingsRepository
	commentsRepository       CommentsRepository
	debatesRepository        DebatesRepository
	txManager                core_postgres.TransactionManager
	realtimeHub              core_realtime.Publisher
}

func (s *CommentRatingsService) publishEvent(postID int, eventType string, data any) error {
	return s.realtimeHub.Publish(
		core_realtime.PostTopic(postID),
		core_realtime.NewEvent(eventType, data),
	)
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

	GetByPostIDForUpdate(
		ctx context.Context,
		postID int,
	) (domain.Debate, error)
}
