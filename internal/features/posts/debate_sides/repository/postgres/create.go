package debate_sides_repository

import (
	"context"
	"fmt"

	"github.com/qandoni/debatesApp/internal/core/domain"
)

const createDebateSideQuery = `
INSERT INTO debatesApp.debate_sides(debate_id, name, description, display_order)
VALUES($1, $2, $3, $4)
RETURNING id, debate_id, name, description, display_order
`

func (r *DebateSidesRepository) CreateDebateSide(
	ctx context.Context,
	side domain.DebateSide,
) (domain.DebateSide, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	db := r.dbFromContext(ctx)
	row := db.QueryRow(ctx, createDebateSideQuery, side.DebateID, side.Name, side.Description, side.DisplayOrder)
	var debateSideModel DebateSideModel

	err := row.Scan(
		&debateSideModel.ID,
		&debateSideModel.DebateID,
		&debateSideModel.Name,
		&debateSideModel.Description,
		&debateSideModel.DisplayOrder,
	)
	if err != nil {
		return domain.DebateSide{}, fmt.Errorf("scan error: %w", err)
	}
	debateSideDomain := domain.NewDebateSide(
		debateSideModel.ID,
		debateSideModel.DebateID,
		debateSideModel.Name,
		debateSideModel.Description,
		debateSideModel.DisplayOrder,
	)
	return debateSideDomain, nil
}
