package domain

import "time"

type PostImage struct {
	ID           int
	PostID       int
	ImageURL     string
	DisplayOrder int
	CreatedAt    time.Time
}
