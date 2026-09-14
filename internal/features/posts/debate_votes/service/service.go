package debate_votes_service

import (
	"context"
	"time"

	"github.com/qandoni/debatesApp/internal/core/domain"
)

func NewDebateVotesService(
	debateVotesRepository DebateVotesRepository,
	debatesRepository DebatesRepository,
	debateSidesRepository DebateSidesRepository,
	commentsRepository CommentsRepository,
) *DebateVotesService {
	return &DebateVotesService{
		debateVotesRepository,
		debatesRepository,
		debateSidesRepository,
		commentsRepository,
	}
}

type DebateVotesService struct {
	debateVotesRepository DebateVotesRepository
	debatesRepository     DebatesRepository
	debateSidesRepository DebateSidesRepository
	commentsRepository    CommentsRepository
}

type CommentsRepository interface {
	HasUserArgumentInPost(
		ctx context.Context,
		userID int,
		postID int,
	) (bool, error)
}

type DebatesRepository interface {
	GetByID(ctx context.Context, debateID int) (domain.Debate, error)
	FinishDebate(
		ctx context.Context,
		debateID int,
	) error
	GetAuthorID(
		ctx context.Context,
		debateID int,
	) (int, error)
}

type DebateSidesRepository interface {
	GetByDebateID(ctx context.Context, debateID int) ([]domain.DebateSide, error)
}

type DebateVotesRepository interface {
	Create(ctx context.Context, vote domain.DebateVote) (domain.DebateVote, error)

	GetByDebateAndUser(
		ctx context.Context,
		debateID int,
		userID int,
	) (domain.DebateVote, error)

	Update(
		ctx context.Context,
		debateID int,
		userID int,
		debateSideID int,
		updatedAt time.Time,
	) (domain.DebateVote, error)
}
