package debate_votes_service

import (
	"context"

	"github.com/qandoni/debatesApp/internal/core/domain"
)

func NewDebateVotesService(
	debateVotesRepository DebateVotesRepository,
	debatesRepository DebatesRepository,
	debateSidesRepository DebateSidesRepository,
) *DebateVotesService {
	return &DebateVotesService{
		debateVotesRepository,
		debatesRepository,
		debateSidesRepository,
	}
}

type DebateVotesService struct {
	debateVotesRepository DebateVotesRepository
	debatesRepository     DebatesRepository
	debateSidesRepository DebateSidesRepository
}

type DebatesRepository interface {
	GetByID(ctx context.Context, debateID int) (domain.Debate, error)
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
}
