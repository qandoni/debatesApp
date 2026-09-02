package debate_votes_repository

import "time"

type DebateVoteModel struct {
	ID           int
	Version      int
	DebateID     int
	UserID       int
	DebateSideID int
	CreatedAt    time.Time
	UpdatedAt    *time.Time
	IsChanged    bool
}
