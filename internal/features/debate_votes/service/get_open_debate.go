package debate_votes_service

import (
	"context"
	"fmt"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_enum "github.com/qandoni/debatesApp/internal/core/enum"
)

func (s *DebateVotesService) getOpenDebate(
	ctx context.Context,
	debateID int,
) (domain.Debate, error) {
	debate, err := s.debatesRepository.GetByID(ctx, debateID)
	if err != nil {
		return domain.Debate{}, fmt.Errorf("get debate: %w", err)
	}

	if debate.Status != core_enum.DebateStatusOpen {
		return domain.Debate{}, fmt.Errorf(
			"debate with id='%d' is not open",
			debateID,
		)
	}

	return debate, nil
}
