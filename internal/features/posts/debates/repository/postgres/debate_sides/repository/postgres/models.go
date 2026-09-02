package debate_sides_repository

import "github.com/qandoni/debatesApp/internal/core/domain"

type DebateSideModel struct {
	ID           int
	DebateID     int
	Name         string
	Description  *string
	DisplayOrder int
}

func debateSideModelToDomain(debateSideModel DebateSideModel) domain.DebateSide {
	return domain.NewDebateSide(
		debateSideModel.ID,
		debateSideModel.DebateID,
		debateSideModel.Name,
		debateSideModel.Description,
		debateSideModel.DisplayOrder,
	)
}

func debateSidesModelsToDomains(debateSideModels []DebateSideModel) []domain.DebateSide {
	debateSidesDomains := make([]domain.DebateSide, len(debateSideModels))
	for i, debateSide := range debateSideModels {
		debateSidesDomains[i] = domain.NewDebateSide(
			debateSide.ID,
			debateSide.DebateID,
			debateSide.Name,
			debateSide.Description,
			debateSide.DisplayOrder,
		)
	}
	return debateSidesDomains
}
