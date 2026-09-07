package domain

import "time"

func NewCommentRating(
	iD int,
	version int,
	commentID int,
	userID int,
	rating int,
	createdAt time.Time,
	updatedAt *time.Time,
) CommentRating {
	return CommentRating{
		iD,
		version,
		commentID,
		userID,
		rating,
		createdAt,
		updatedAt,
	}
}

type CommentRating struct {
	ID        int
	Version   int
	CommentID int
	UserID    int
	Rating    int
	CreatedAt time.Time
	UpdatedAt *time.Time
}
