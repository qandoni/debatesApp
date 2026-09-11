package debate_votes_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
	core_postgres_pool "github.com/qandoni/debatesApp/internal/core/repository/postgres/pool"
	debate_votes_contracts "github.com/qandoni/debatesApp/internal/features/posts/debate_votes/contracts"
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

func (r *DebateVotesRepository) GetResults(
	ctx context.Context,
	debateID int,
) ([]debate_votes_contracts.DebateVoteResult, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	query := `
SELECT 
	debate_side_id,
	COUNT(*) AS votes_count
FROM debatesApp.debate_votes
WHERE debate_id = $1
GROUP BY debate_side_id
ORDER BY votes_count DESC;
`

	db := r.dbFromContext(ctx)
	rows, err := db.Query(ctx, query, debateID)
	if err != nil {
		return []debate_votes_contracts.DebateVoteResult{}, fmt.Errorf("select debate vote results: %w", err)
	}
	defer rows.Close()

	results := make([]debate_votes_contracts.DebateVoteResult, 0)
	for rows.Next() {
		var result debate_votes_contracts.DebateVoteResult
		err := rows.Scan(
			&result.DebateSideID,
			&result.VotesCount,
		)
		if err != nil {
			return []debate_votes_contracts.DebateVoteResult{}, fmt.Errorf("scan debate vote result: %w", err)
		}
		results = append(results, result)
	}
	if err := rows.Err(); err != nil {
		return []debate_votes_contracts.DebateVoteResult{}, fmt.Errorf("iterate debate vote results: %w", err)
	}
	return results, nil
}
