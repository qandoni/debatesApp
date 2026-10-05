package debate_votes_service

import (
	"context"
	"fmt"
	"time"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_enum "github.com/qandoni/debatesApp/internal/core/enum"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
	core_realtime "github.com/qandoni/debatesApp/internal/core/realtime"
)

func (s *DebateVotesService) Vote(
	ctx context.Context,
	userID int,
	debateID int,
	debateSideID int,
) (domain.DebateVote, error) {
	var (
		createdVote  domain.DebateVote
		debatePostID int
	)

	if err := s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		debate, err := s.getOpenDebateForUpdate(txCtx, debateID)
		if err != nil {
			return err
		}

		if err := s.validateDebateSide(txCtx, debateID, debateSideID); err != nil {
			return err
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

		createdVote, err = s.debateVotesRepository.Create(txCtx, vote)
		if err != nil {
			return fmt.Errorf(
				"create debate vote: %w",
				err,
			)
		}
		debatePostID = debate.PostID
		return nil
	}); err != nil {
		return domain.DebateVote{}, err
	}

	if err := s.publishEvent(
		debatePostID,
		core_realtime.EventDebateVoteCreated,
		core_realtime.NewVoteData(createdVote, debatePostID, false),
	); err != nil {
		return domain.DebateVote{}, fmt.Errorf("publish debate vote created event: %w", err)
	}

	return createdVote, nil
}

func (s *DebateVotesService) ChangeVote(
	ctx context.Context,
	userID int,
	debateID int,
	debateSideID int,
) (domain.DebateVote, error) {
	var (
		updatedVote  domain.DebateVote
		debatePostID int
	)

	if err := s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		// Блокировка строки дебата сериализует проверку статуса и
		// проверку «нет ли уже аргумента» с параллельными FinishDebate
		// и CreateArgument.
		debate, err := s.getOpenDebateForUpdate(txCtx, debateID)
		if err != nil {
			return fmt.Errorf("get open debate: %w", err)
		}

		if err := s.validateDebateSide(txCtx, debateID, debateSideID); err != nil {
			return err
		}

		hasArgument, err := s.commentsRepository.HasUserArgumentInPost(txCtx, userID, debate.PostID)
		if err != nil {
			return fmt.Errorf("failed to check if user have argument in this post already: %w", err)
		}
		if hasArgument {
			return fmt.Errorf(
				"user id='%d' cannot change vote after creating an argument in debate id='%d': %w",
				userID,
				debateID,
				core_errors.ErrAccessForbidden,
			)
		}

		updatedVote, err = s.debateVotesRepository.Update(
			txCtx,
			debateID,
			userID,
			debateSideID,
			time.Now(),
		)
		if err != nil {
			return fmt.Errorf(
				"update debate vote: %w",
				err,
			)
		}
		debatePostID = debate.PostID
		return nil
	}); err != nil {
		return domain.DebateVote{}, err
	}

	if err := s.publishEvent(
		debatePostID,
		core_realtime.EventDebateVoteChanged,
		core_realtime.NewVoteData(updatedVote, debatePostID, true),
	); err != nil {
		return domain.DebateVote{}, fmt.Errorf("publish debate vote changed event: %w", err)
	}

	return updatedVote, nil
}

// getOpenDebateForUpdate читает дебат с блокировкой строки (FOR UPDATE).
// Вызывается только внутри WithinTransaction: блокировка сериализует
// проверку статуса с параллельным FinishDebate и операциями того же дебата.
func (s *DebateVotesService) getOpenDebateForUpdate(
	ctx context.Context,
	debateID int,
) (domain.Debate, error) {
	debate, err := s.debatesRepository.GetByIDForUpdate(ctx, debateID)
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
