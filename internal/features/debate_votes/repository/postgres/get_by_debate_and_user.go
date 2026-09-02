package debate_votes_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
	core_postgres_pool "github.com/qandoni/debatesApp/internal/core/repository/postgres/pool"
)

func (r *DebateVotesRepository) GetByDebateAndUser(
	ctx context.Context,
	debateID int,
	userID int,
) (domain.DebateVote, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	query := `
		SELECT
			id,
			version,
			debate_id,
			user_id,
			debate_side_id,
			created_at,
			updated_at,
			is_changed
		FROM debatesApp.debate_votes
		WHERE debate_id = $1
		  AND user_id = $2
	`

	db := r.dbFromContext(ctx)

	row := db.QueryRow(ctx, query, debateID, userID)

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
				"vote for debate id='%d' and user id='%d': %w",
				debateID,
				userID,
				core_errors.ErrNotFound,
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
