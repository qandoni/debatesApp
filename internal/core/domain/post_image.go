package domain

import "time"

func NewPostImage(
	iD int,
	postID int,
	imageURL string,
	displayOrder int,
	createdAt time.Time,
) PostImage {
	return PostImage{
		iD,
		postID,
		imageURL,
		displayOrder,
		createdAt,
	}
}

type PostImage struct {
	ID           int
	PostID       int
	ImageURL     string
	DisplayOrder int
	CreatedAt    time.Time
}
