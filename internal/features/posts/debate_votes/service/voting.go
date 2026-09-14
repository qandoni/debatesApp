package debate_votes_service

import (
	"context"
	"fmt"
	"time"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_enum "github.com/qandoni/debatesApp/internal/core/enum"
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

func (s *DebateVotesService) ChangeVote(
	ctx context.Context,
	userID int,
	debateID int,
	debateSideID int,
) (domain.DebateVote, error) {

	debate, err := s.getOpenDebate(ctx, debateID)
	if err != nil {
		return domain.DebateVote{}, fmt.Errorf("get open debate: %w", err)
	}

	if err := s.validateDebateSide(ctx, debateID, debateSideID); err != nil {
		return domain.DebateVote{}, err
	}

	hasArgument, err := s.commentsRepository.HasUserArgumentInPost(ctx, userID, debate.PostID)
	if err != nil {
		return domain.DebateVote{}, fmt.Errorf("failed to check if user have argument in this post already: %w", err)
	}
	if hasArgument {
		return domain.DebateVote{}, fmt.Errorf(
			"user id='%d' cannot change vote after creating an argument in debate id='%d': %w",
			userID,
			debateID,
			core_errors.ErrAccessForbidden,
		)
	}

	updatedVote, err := s.debateVotesRepository.Update(
		ctx,
		debateID,
		userID,
		debateSideID,
		time.Now(),
	)
	if err != nil {
		return domain.DebateVote{}, fmt.Errorf(
			"update debate vote: %w",
			err,
		)
	}

	return updatedVote, nil
}

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
