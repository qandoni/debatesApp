package debate_votes_service

import (
	"context"
	"fmt"

	core_enum "github.com/qandoni/debatesApp/internal/core/enum"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
)

func (s *DebateVotesService) FinishDebate(
	ctx context.Context,
	userID int,
	debateID int,
) error {
	debate, err := s.debatesRepository.GetByID(ctx, debateID)
	if err != nil {
		return fmt.Errorf("get debate: %w", err)
	}
	if debate.Status != core_enum.DebateStatusOpen {
		return fmt.Errorf("debate is already finished")
	}

	authorID, err := s.debatesRepository.GetAuthorID(ctx, debateID)
	if err != nil {
		return fmt.Errorf("get author id: %w", err)
	}
	if userID != authorID {
		return fmt.Errorf("user is not the author of the debate: %w", core_errors.ErrAccessForbidden)
	}

	winnerSideID, err := s.CalculateWinner(ctx, debateID)
	if err != nil {
		return fmt.Errorf("calculate winner: %w", err)
	}
	if err := s.debatesRepository.FinishDebate(ctx, debateID, winnerSideID); err != nil {
		return fmt.Errorf("finish debate in repository: %w", err)
	}
	return nil
}
