package domain

import "time"

type Post struct {
	ID        int
	Version   int
	AuthorID  int
	Content   string
	IsDebate  bool
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
