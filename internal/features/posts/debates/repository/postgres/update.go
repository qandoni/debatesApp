package debates_repository

import (
	"context"
	"fmt"

	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
)

const finishDebateQuery = `
WITH vote_counts AS (
    SELECT
        debate_side_id,
        COUNT(*) AS votes_count
    FROM debatesApp.debate_votes
    WHERE debate_id = $1
    GROUP BY debate_side_id
    ORDER BY
        votes_count DESC,
        debate_side_id ASC
),
first_place AS (
    SELECT
        debate_side_id,
        votes_count
    FROM vote_counts
    LIMIT 1
),
second_place AS (
    SELECT
        debate_side_id,
        votes_count
    FROM vote_counts
    OFFSET 1
    LIMIT 1
)
UPDATE debatesApp.debates AS d
SET
    status = 'FINISHED',
    finished_at = NOW(),
    winner_side_id = (
        SELECT CASE
                   WHEN second_place.debate_side_id IS NULL
                        OR first_place.votes_count <> second_place.votes_count
                   THEN first_place.debate_side_id
                   ELSE NULL
               END
        FROM first_place
        LEFT JOIN second_place ON TRUE
    )
WHERE d.id = $1
  AND d.status = 'OPEN';
`

func (r *DebatesRepository) FinishDebate(
	ctx context.Context,
	debateID int,
) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	db := r.dbFromContext(ctx)
	cmdTag, err := db.Exec(ctx, finishDebateQuery, debateID)
	if err != nil {
		return fmt.Errorf("finish debate: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return core_errors.ErrConflict
	}
	return nil
}
