package posts_contracts

import "time"

type CreatePostInput struct {
	AuthorID int
	Content  string
	IsDebate bool
	Debate   *CreateDebateInput
}

type CreateDebateInput struct {
	EndAt *time.Time
	Sides []CreateDebateSideInput
}

type CreateDebateSideInput struct {
	Name        string
	Description *string
}
