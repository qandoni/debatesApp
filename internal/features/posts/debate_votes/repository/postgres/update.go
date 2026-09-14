package debate_votes_repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
	core_postgres_pool "github.com/qandoni/debatesApp/internal/core/repository/postgres/pool"
)

const updateDebateVoteQuery = `
UPDATE debatesApp.debate_votes
SET
	debate_side_id = $1,
	updated_at = $2,
	is_changed = true,
	version = version + 1
WHERE debate_id = $3
  AND user_id = $4
  AND is_changed = false
  AND debate_side_id != $1
RETURNING
	id,
	version,
	debate_id,
	user_id,
	debate_side_id,
	created_at,
	updated_at,
	is_changed
`

func (r *DebateVotesRepository) Update(
	ctx context.Context,
	debateID int,
	userID int,
	debateSideID int,
	updatedAt time.Time,
) (domain.DebateVote, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	db := r.dbFromContext(ctx)

	row := db.QueryRow(
		ctx,
		updateDebateVoteQuery,
		debateSideID,
		updatedAt,
		debateID,
		userID,
	)

	var model DebateVoteModel

	err := row.Scan(
		&model.ID,
		&model.Version,
		&model.DebateID,
		&model.UserID,
		&model.DebateSideID,
		&model.CreatedAt,
		&model.UpdatedAt,
		&model.IsChanged,
	)
	if err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return domain.DebateVote{}, fmt.Errorf(
				"vote for debate id='%d' and user id='%d' cannot be changed: %w",
				debateID,
				userID,
				core_errors.ErrConflict,
			)
		}

		return domain.DebateVote{}, fmt.Errorf("scan error: %w", err)
	}

	return domain.NewDebateVote(
		model.ID,
		model.Version,
		model.DebateID,
		model.UserID,
		model.DebateSideID,
		model.CreatedAt,
		model.UpdatedAt,
		model.IsChanged,
	), nil
}
