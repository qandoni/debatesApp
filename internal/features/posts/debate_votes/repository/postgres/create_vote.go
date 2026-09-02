package debate_votes_repository

import (
	"context"
	"fmt"

	"github.com/qandoni/debatesApp/internal/core/domain"
)

func (r *DebateVotesRepository) Create(
	ctx context.Context,
	vote domain.DebateVote,
) (domain.DebateVote, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	query := `
		INSERT INTO debatesApp.debate_votes (
			debate_id,
			user_id,
			debate_side_id,
			created_at,
			updated_at,
			is_changed
		)
		VALUES ($1, $2, $3, $4, $5, $6)
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
