package debate_votes_repository

import (
	"context"
	"fmt"

	debate_votes_contracts "github.com/qandoni/debatesApp/internal/features/posts/debate_votes/contracts"
)

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
