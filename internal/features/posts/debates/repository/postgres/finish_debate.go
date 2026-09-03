package debates_repository

import (
	"context"
	"fmt"

	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
)

func (r *DebatesRepository) FinishDebate(
	ctx context.Context,
	debateID int,
	winnerSideID *int,
) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	query := `
		UPDATE debatesApp.debates
		SET
			status = 'FINISHED',
			finished_at = NOW(),
			winner_side_id = $1
		WHERE id = $2
		  AND status = 'OPEN'
	`
	db := r.dbFromContext(ctx)
	cmdTag, err := db.Exec(ctx, query, winnerSideID, debateID)
	if err != nil {
		return fmt.Errorf("finish debate: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return core_errors.ErrConflict
	}
	return nil
}
