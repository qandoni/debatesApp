package debate_sides_repository

import (
	"context"
	"fmt"

	"github.com/qandoni/debatesApp/internal/core/domain"
)

func (r *DebateSidesRepository) GetByDebateID(
	ctx context.Context,
	debateID int,
) ([]domain.DebateSide, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	query := `
	SELECT id, debate_id, name, description, display_order
	FROM debatesApp.debate_sides
	WHERE debate_id=$1
	ORDER BY display_order ASC;
	`

	db := r.dbFromContext(ctx)
	rows, err := db.Query(ctx, query, debateID)
	if err != nil {
		return []domain.DebateSide{}, fmt.Errorf("select debate sides: %w", err)
	}
	defer rows.Close()

	var debateSideModels []DebateSideModel

	for rows.Next() {
		var debateSideModel DebateSideModel
		err := rows.Scan(
			&debateSideModel.ID,
			&debateSideModel.DebateID,
			&debateSideModel.Name,
			&debateSideModel.Description,
			&debateSideModel.DisplayOrder,
		)
		if err != nil {
			return nil, fmt.Errorf("scan debate sides: %w", err)
		}
		debateSideModels = append(debateSideModels, debateSideModel)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("next rows: %w", err)
	}
	debateSidesDomains := debateSidesModelsToDomains(debateSideModels)
	return debateSidesDomains, nil
}
