package debate_votes_service

import (
	"context"

	"github.com/qandoni/debatesApp/internal/core/domain"
	debate_votes_contracts "github.com/qandoni/debatesApp/internal/features/posts/debate_votes/contracts"
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
		winnerSideID *int,
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

	Update(ctx context.Context, vote domain.DebateVote) (domain.DebateVote, error)
	GetResults(ctx context.Context, debateID int) ([]debate_votes_contracts.DebateVoteResult, error)
}
