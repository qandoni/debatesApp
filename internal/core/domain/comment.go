package domain

import "time"

type Comment struct {
	ID              int
	Version         int
	PostID          int
	ParentCommentID *int
	AuthorID        int
	DebateSideID    *int
	Content         string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
