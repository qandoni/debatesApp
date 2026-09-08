package debate_votes_service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
)

// TODO Пользователь может изменить свой голос в дебате только до того, как опубликует свой первый аргумент/комментарий. После добавления фичи аргументов нужно доделать этот метод
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

	vote, err := s.debateVotesRepository.GetByDebateAndUser(
		ctx,
		debateID,
		userID,
	)

	if err != nil {
		if errors.Is(err, core_errors.ErrNotFound) {
			return domain.DebateVote{}, fmt.Errorf(
				"user id='%d' has not voted in debate id='%d': %w",
				userID,
				debateID,
				core_errors.ErrAccessForbidden,
			)
		}

		return domain.DebateVote{}, fmt.Errorf(
			"get debate vote: %w",
			err,
		)
	}

	if vote.IsChanged {
		return domain.DebateVote{}, fmt.Errorf(
			"user id='%d' has already changed vote in debate id='%d': %w",
			userID,
			debateID,
			core_errors.ErrAccessForbidden,
		)
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

	if vote.DebateSideID == debateSideID {
		return domain.DebateVote{}, fmt.Errorf(
			"user id='%d' already votes for debate side id='%d': %w",
			userID,
			debateSideID,
			core_errors.ErrAccessForbidden,
		)
	}

	now := time.Now()

	vote.DebateSideID = debateSideID
	vote.UpdatedAt = &now
	vote.IsChanged = true

	updatedVote, err := s.debateVotesRepository.Update(ctx, vote)
	if err != nil {
		return domain.DebateVote{}, fmt.Errorf(
			"update debate vote: %w",
			err,
		)
	}

	return updatedVote, nil
}
