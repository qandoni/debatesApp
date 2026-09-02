package debate_votes_service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
)

func (s *DebateVotesService) Vote(
	ctx context.Context,
	userID int,
	debateID int,
	debateSideID int,
) (domain.DebateVote, error) {

	if _, err := s.getOpenDebate(ctx, debateID); err != nil {
		return domain.DebateVote{}, err
	}

	if err := s.validateDebateSide(ctx, debateID, debateSideID); err != nil {
		return domain.DebateVote{}, err
	}

	_, err := s.debateVotesRepository.GetByDebateAndUser(
		ctx,
		debateID,
		userID,
	)

	if err == nil {
		return domain.DebateVote{}, fmt.Errorf(
			"user id='%d' has already voted in debate id='%d'",
			userID,
			debateID,
		)
	}

	if !errors.Is(err, core_errors.ErrNotFound) {
		return domain.DebateVote{}, fmt.Errorf(
			"get existing vote: %w",
			err,
		)
	}

	vote := domain.NewDebateVote(
		0,
		1,
		debateID,
		userID,
		debateSideID,
		time.Now(),
		nil,
		false,
	)

	createdVote, err := s.debateVotesRepository.Create(ctx, vote)
	if err != nil {
		return domain.DebateVote{}, fmt.Errorf(
			"create debate vote: %w",
			err,
		)
	}

	return createdVote, nil
}
