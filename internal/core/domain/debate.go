package domain

import (
	"time"

	core_enum "github.com/qandoni/debatesApp/internal/core/enum"
)

func NewDebate(
	iD int,
	postID int,
	status core_enum.DebatesStatus,
	endAt *time.Time,
	createdAt time.Time,
	finishedAt *time.Time,
	winnerSideID *int,
) Debate {
	return Debate{
		iD,
		postID,
		status,
		endAt,
		createdAt,
		finishedAt,
		winnerSideID,
		nil,
	}
}

type Debate struct {
	ID           int
	PostID       int
	Status       core_enum.DebatesStatus
	EndAt        *time.Time
	CreatedAt    time.Time
	FinishedAt   *time.Time
	WinnerSideID *int
	Sides        []DebateSide
}

func NewDebateUnitialized(
	postID int,
	endAt *time.Time,
) Debate {
	return NewDebate(
		UninitializedID,
		postID,
		core_enum.DebateStatusOpen,
		endAt,
		time.Now(),
		nil,
		nil,
	)
}
