package debate_votes_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
	core_postgres_pool "github.com/qandoni/debatesApp/internal/core/repository/postgres/pool"
)

func (r *DebateVotesRepository) Update(
	ctx context.Context,
	vote domain.DebateVote,
) (domain.DebateVote, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	query := `
		UPDATE debatesApp.debate_votes
		SET
			debate_side_id = $1,
			updated_at = $2,
			is_changed = $3,
			version = version + 1
		WHERE id = $4
		  AND version = $5
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

	db := r.dbFromContext(ctx)

	row := db.QueryRow(
		ctx,
		query,
		vote.DebateSideID,
		vote.UpdatedAt,
		vote.IsChanged,
		vote.ID,
		vote.Version,
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
				"vote with id='%d' concurrently accessed: %w",
				vote.ID,
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
