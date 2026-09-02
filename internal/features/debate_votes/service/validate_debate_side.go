package debate_votes_service

import (
	"context"
	"fmt"

	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
)

func (s *DebateVotesService) validateDebateSide(
	ctx context.Context,
	debateID int,
	debateSideID int,
) error {
	sides, err := s.debateSidesRepository.GetByDebateID(ctx, debateID)
	if err != nil {
		return fmt.Errorf("get debate sides: %w", err)
	}

	for _, side := range sides {
		if side.ID == debateSideID {
			return nil
		}
	}

	return fmt.Errorf(
		"debate side with id='%d' does not belong to debate id='%d': %w",
		debateSideID,
		debateID,
		core_errors.ErrNotFound,
	)
}
