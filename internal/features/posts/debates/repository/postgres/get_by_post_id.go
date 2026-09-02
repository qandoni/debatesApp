package debates_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
	core_postgres_pool "github.com/qandoni/debatesApp/internal/core/repository/postgres/pool"
)

func (r *DebatesRepository) GetByPostID(
	ctx context.Context,
	postID int,
) (domain.Debate, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	query := `
	SELECT id, post_id, status, end_at, created_at, finished_at, winner_side_id
	FROM debatesApp.debates
	WHERE post_id = $1;
	`

	db := r.dbFromContext(ctx)
	row := db.QueryRow(ctx, query, postID)
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
			return domain.Debate{}, fmt.Errorf("post with id='%d': %w", postID, core_errors.ErrNotFound)
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
