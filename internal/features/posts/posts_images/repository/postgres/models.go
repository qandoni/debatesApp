package post_images_repository

import (
	"time"

	"github.com/qandoni/debatesApp/internal/core/domain"
)

func NewPostImagesModel(
	iD int,
	postID int,
	imageURL string,
	displayOrder int,
	createdAt time.Time,
) PostImagesModel {
	return PostImagesModel{
		iD,
		postID,
		imageURL,
		displayOrder,
		createdAt,
	}
}

type PostImagesModel struct {
	ID           int
	PostID       int
	ImageURL     string
	DisplayOrder int
	CreatedAt    time.Time
}

func postImagesDomainsFromModels(postImages []PostImagesModel) []domain.PostImage {
	postImagesDomains := make([]domain.PostImage, len(postImages))
	for i, postImage := range postImages {
		postImagesDomains[i] = domain.NewPostImage(
			postImage.ID,
			postImage.PostID,
			postImage.ImageURL,
			postImage.DisplayOrder,
			postImage.CreatedAt,
		)
	}
	return postImagesDomains
}
