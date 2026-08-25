package domain

import "time"

type DebateStatus string

const (
	DebateStatusActive   DebateStatus = "active"
	DebateStatusFinished DebateStatus = "finished"
	// может мне еще добавить статус cancelled?
)

type Debate struct {
	ID           int64
	PostID       int64
	Status       DebateStatus
	EndAt        time.Time
	CreatedAt    time.Time
	FinishedAt   *time.Time
	WinnerSideID *int64
}
