package debate_votes_service

import (
	"context"
	"fmt"

	core_enum "github.com/qandoni/debatesApp/internal/core/enum"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
	core_realtime "github.com/qandoni/debatesApp/internal/core/realtime"
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

	if err := s.debatesRepository.FinishDebate(ctx, debateID); err != nil {
		return fmt.Errorf("finish debate in repository: %w", err)
	}

	finishedDebate, err := s.debatesRepository.GetByID(ctx, debateID)
	if err != nil {
		if publishErr := s.publishEvent(
			debate.PostID,
			core_realtime.EventDebateFinished,
			core_realtime.NewDebateFinishedData(debate, userID),
		); publishErr != nil {
			return fmt.Errorf("publish debate finished event: %w", publishErr)
		}

		return nil
	}

	if err := s.publishEvent(
		finishedDebate.PostID,
		core_realtime.EventDebateFinished,
		core_realtime.NewDebateFinishedData(finishedDebate, userID),
	); err != nil {
		return fmt.Errorf("publish debate finished event: %w", err)
	}

	return nil
}
