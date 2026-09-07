package domain

import (
	"time"
)

func NewComment(
	iD int,
	version int,
	postID int,
	parentCommentID *int,
	authorID int,
	debateSideID *int,
	content string,
	createdAt time.Time,
	updatedAt *time.Time,
) Comment {
	return Comment{
		iD,
		version,
		postID,
		parentCommentID,
		authorID,
		debateSideID,
		content,
		createdAt,
		updatedAt,
	}
}

func NewCommentUninitialized(
	postID int,
	authorID int,
	parentCommentID *int,
	debateSideID *int,
	content string,
) Comment {
	return NewComment(
		UninitializedID,
		UninitializedVersion,
		postID,
		parentCommentID,
		authorID,
		debateSideID,
		content,
		time.Now(),
		nil,
	)
}

type Comment struct {
	ID              int
	Version         int
	PostID          int
	ParentCommentID *int
	AuthorID        int
	DebateSideID    *int
	Content         string
	CreatedAt       time.Time
	UpdatedAt       *time.Time
}
