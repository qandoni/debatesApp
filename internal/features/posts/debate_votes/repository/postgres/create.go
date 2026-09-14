package debate_votes_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
	core_postgres_pool "github.com/qandoni/debatesApp/internal/core/repository/postgres/pool"
)

const createDebateVoteQuery = `
INSERT INTO debatesApp.debate_votes (
	debate_id,
	user_id,
	debate_side_id,
	created_at,
	updated_at,
	is_changed
)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (debate_id, user_id) DO NOTHING
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

func (r *DebateVotesRepository) Create(
	ctx context.Context,
	vote domain.DebateVote,
) (domain.DebateVote, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	db := r.dbFromContext(ctx)

	row := db.QueryRow(
		ctx,
		createDebateVoteQuery,
		vote.DebateID,
		vote.UserID,
		vote.DebateSideID,
		vote.CreatedAt,
		vote.UpdatedAt,
		vote.IsChanged,
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
				"user already voted in debate id='%d': %w",
				vote.DebateID,
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
