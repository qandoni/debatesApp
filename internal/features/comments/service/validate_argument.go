package comments_service

import (
	"context"
	"fmt"

	core_enum "github.com/qandoni/debatesApp/internal/core/enum"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
)

func (s *CommentsService) validateArgumentCreation(
	ctx context.Context,
	userID int,
	postID int,
	debateSideID int,
) error {
	if _, err := s.postsRepository.GetPost(ctx, postID); err != nil {
		return fmt.Errorf("get post: %w", err)
	}

	debate, err := s.debatesRepository.GetByPostID(ctx, postID)
	if err != nil {
		return fmt.Errorf("get debate: %w", err)
	}
	if debate.Status != core_enum.DebateStatusOpen {
		return fmt.Errorf("debates are closed: %w", core_errors.ErrConflict)
	}

	sides, err := s.debateSidesRepository.GetByDebateID(ctx, debate.ID)
	if err != nil {
		return fmt.Errorf("get debate sides: %w", err)
	}

	var sideExists bool

	for _, side := range sides {
		if side.ID == debateSideID {
			sideExists = true
			break
		}
	}
	if !sideExists {
		return core_errors.ErrNotFound
	}

	vote, err := s.debateVotesRepository.GetByDebateAndUser(ctx, debate.ID, userID)
	if err != nil {
		return fmt.Errorf("get user vote: %w", err)
	}
	if vote.DebateSideID != debateSideID {
		return fmt.Errorf("user voted for other side: '%d': %w", vote.DebateSideID, core_errors.ErrConflict)
	}

	return nil
}
