package comments_service

import (
	"context"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_postgres "github.com/qandoni/debatesApp/internal/core/repository/postgres"
)

func NewCommentsService(
	commentsRepository CommentsRepository,
	postsRepository PostsRepository,
	debatesRepository DebatesRepository,
	debateSidesRepository DebateSidesRepository,
	debateVotesRepository DebateVotesRepository,
	txManager core_postgres.TransactionManager,
) *CommentsService {
	return &CommentsService{
		commentsRepository,
		postsRepository,
		debatesRepository,
		debateSidesRepository,
		debateVotesRepository,
		txManager,
	}
}

type CommentsService struct {
	commentsRepository    CommentsRepository
	postsRepository       PostsRepository
	debatesRepository     DebatesRepository
	debateSidesRepository DebateSidesRepository
	debateVotesRepository DebateVotesRepository
	txManager             core_postgres.TransactionManager
}

type DebateVotesRepository interface {
	Create(ctx context.Context, vote domain.DebateVote) (domain.DebateVote, error)

	GetByDebateAndUser(
		ctx context.Context,
		debateID int,
		userID int,
	) (domain.DebateVote, error)

	Update(ctx context.Context, vote domain.DebateVote) (domain.DebateVote, error)
}

type DebateSidesRepository interface {
	CreateDebateSide(
		ctx context.Context,
		side domain.DebateSide,
	) (domain.DebateSide, error)

	GetByDebateID(
		ctx context.Context,
		debateID int,
	) ([]domain.DebateSide, error)
}

type DebatesRepository interface {
	CreateDebate(
		ctx context.Context,
		debate domain.Debate,
	) (domain.Debate, error)

	GetByPostID(
		ctx context.Context,
		postID int,
	) (domain.Debate, error)

	GetByID(
		ctx context.Context,
		debateID int,
	) (domain.Debate, error)
	FinishDebate(
		ctx context.Context,
		debateID int,
	) error
	GetAuthorID(
		ctx context.Context,
		debateID int,
	) (int, error)
}

type PostsRepository interface {
	CreatePost(
		ctx context.Context,
		post domain.Post,
	) (domain.Post, error)
	DeletePost(
		ctx context.Context,
		userID int,
		postID int,
	) error
	GetPost(
		ctx context.Context,
		postID int,
	) (domain.Post, error)
	PatchPost(
		ctx context.Context,
		userID int,
		postID int,
		postPatch domain.PostPatch,
	) (domain.Post, error)
	GetPosts(
		ctx context.Context,
		limit *int,
		offset *int,
	) ([]domain.Post, error)
}

type CommentsRepository interface {
	SetAuthorLike(
		ctx context.Context,
		commentID int,
		liked bool,
	) (domain.Comment, error)
	GetByPostIDWithoutPagination(
		ctx context.Context,
		postID int,
	) ([]domain.Comment, error)
	GetArgumentsWithReplies(
		ctx context.Context,
		postID int,
		limit *int,
		offset *int,
	) ([]domain.CommentWithRating, error)
	UpdateComment(
		ctx context.Context,
		comment domain.Comment,
	) (domain.Comment, error)
	HasUserArgumentInPost(
		ctx context.Context,
		userID int,
		postID int,
	) (bool, error)
	GetByPostID(
		ctx context.Context,
		postID int,
		limit *int,
		offset *int,
	) ([]domain.Comment, error)
	GetByID(
		ctx context.Context,
		commentID int,
	) (domain.Comment, error)
	CreateComment(
		ctx context.Context,
		comment domain.Comment,
	) (domain.Comment, error)
}
