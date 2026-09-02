package domain

import "time"

func NewDebateVote(
	iD int,
	version int,
	debateID int,
	userID int,
	debateSideID int,
	createdAt time.Time,
	updatedAt *time.Time,
	isChanged bool,
) DebateVote {
	return DebateVote{
		iD,
		version,
		debateID,
		userID,
		debateSideID,
		createdAt,
		updatedAt,
		isChanged,
	}
}

type DebateVote struct {
	ID           int
	Version      int
	DebateID     int
	UserID       int
	DebateSideID int
	CreatedAt    time.Time
	UpdatedAt    *time.Time
	IsChanged    bool
}
