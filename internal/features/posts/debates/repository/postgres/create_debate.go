package debates_repository

import (
	"context"
	"fmt"

	"github.com/qandoni/debatesApp/internal/core/domain"
)

func (r *DebatesRepository) CreateDebate(
	ctx context.Context,
	debate domain.Debate,
) (domain.Debate, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	query := `	
	INSERT INTO debatesApp.debates(post_id, status, end_at, created_at, finished_at, winner_side_id)
	VALUES($1, $2, $3, $4, $5, $6)
	RETURNING id, post_id, status, end_at, created_at, finished_at, winner_side_id
	`

	db := r.dbFromContext(ctx)
	row := db.QueryRow(
		ctx,
		query,
		debate.PostID,
		debate.Status,
		debate.EndAt,
		debate.CreatedAt,
		debate.FinishedAt,
		debate.WinnerSideID,
	)

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
