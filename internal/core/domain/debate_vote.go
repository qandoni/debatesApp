package domain

import "time"

type DebateVote struct {
	ID           int
	Version      int
	DebateID     int
	UserID       int
	DebateSideID int
	CreatedAt    time.Time
	ChangedAt    *time.Time
	IsChanged    bool
}
