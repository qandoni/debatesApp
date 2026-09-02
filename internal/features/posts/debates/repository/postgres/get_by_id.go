package debates_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
	core_postgres_pool "github.com/qandoni/debatesApp/internal/core/repository/postgres/pool"
)

func (r *DebatesRepository) GetByID(
	ctx context.Context,
	debateID int,
) (domain.Debate, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	query := `
	SELECT id, post_id, status, end_at, created_at, finished_at, winner_side_id
	FROM debatesApp.debates
	WHERE id = $1
	`

	db := r.dbFromContext(ctx)
	row := db.QueryRow(ctx, query, debateID)
	var debateModel DebateModel

	err := row.Scan(
		&debateModel.ID,
		&debateModel.PostID,
		&debateModel.Status,
		&debateModel.EndAt,
		&debateModel.CreatedAt,
		&debateModel.FinishedAt,
		&debateModel.WinnerSideID,
	)
	if err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return domain.Debate{}, fmt.Errorf("debate with id='%d': %w", debateID, core_errors.ErrNotFound)
		}
		return domain.Debate{}, fmt.Errorf("scan error: %w", err)
	}

	debateDomain := domain.NewDebate(
		debateModel.ID,
		debateModel.PostID,
		debateModel.Status,
		debateModel.EndAt,
		debateModel.CreatedAt,
		debateModel.FinishedAt,
		debateModel.WinnerSideID,
	)
	return debateDomain, nil
}
