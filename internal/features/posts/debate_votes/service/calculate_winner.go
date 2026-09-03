package debate_votes_service

import (
	"context"
	"fmt"
)

func (s *DebateVotesService) CalculateWinner(
	ctx context.Context,
	debateID int,
) (*int, error) {
	results, err := s.debateVotesRepository.GetResults(ctx, debateID)
	if err != nil {
		return nil, fmt.Errorf("get vote results: %w", err)
	}

	if len(results) == 0 {
		return nil, nil
	}
	maxvotes := results[0].VotesCount
	if len(results) > 1 && results[1].VotesCount == maxvotes {
		return nil, nil
	}
	winnerSideID := results[0].DebateSideID
	return &winnerSideID, nil
}
