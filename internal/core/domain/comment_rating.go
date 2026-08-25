package domain

import "time"

type CommentRating struct {
	ID        int
	Version   int
	CommentID int
	UserID    int
	Rating    int
	CreatedAt time.Time
	UpdatedAt time.Time
}
