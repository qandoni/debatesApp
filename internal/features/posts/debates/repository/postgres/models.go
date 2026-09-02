package debates_repository

import (
	"time"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_enum "github.com/qandoni/debatesApp/internal/core/enum"
)

type DebateModel struct {
	ID           int
	PostID       int
	Status       core_enum.DebatesStatus
	EndAt        *time.Time
	CreatedAt    time.Time
	FinishedAt   *time.Time
	WinnerSideID *int
}

func debateModelToDomain(debateModel DebateModel) domain.Debate {
	return domain.NewDebate(
		debateModel.ID,
		debateModel.PostID,
		debateModel.Status,
		debateModel.EndAt,
		debateModel.CreatedAt,
		debateModel.FinishedAt,
		debateModel.WinnerSideID,
	)
}
