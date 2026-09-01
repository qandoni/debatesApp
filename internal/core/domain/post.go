package domain

import "time"

func NewPost(
	iD int,
	version int,
	authorID int,
	content string,
	isDebate bool,
	createdAt time.Time,
	updatedAt *time.Time,
	deletedAt *time.Time,
) Post {
	return Post{
		iD,
		version,
		authorID,
		content,
		isDebate,
		nil,
		createdAt,
		updatedAt,
		deletedAt,
	}
}

type Post struct {
	ID        int
	Version   int
	AuthorID  int
	Content   string
	IsDebate  bool
	Images    []PostImage
	CreatedAt time.Time
	UpdatedAt *time.Time
	DeletedAt *time.Time
}

func NewPostUninitialized(
	authorID int,
	content string,
	isDebate bool,
) Post {
	return NewPost(
		UninitializedID,
		UninitializedVersion,
		authorID,
		content,
		isDebate,
		time.Now(),
		nil,
		nil,
	)
}

func NewPostPatch(
	content Nullable[string],
) PostPatch {
	return PostPatch{
		content,
	}
}

type PostPatch struct {
	Content Nullable[string]
}
