package images_service

import (
	"context"

	"github.com/qandoni/debatesApp/internal/core/domain"
	"github.com/qandoni/debatesApp/internal/features/storage"
)

func NewImagesService(
	storage storage.ImageStorage,
	postsRepository PostsRepository,
	postImagesRepository PostImagesRepository,
) *ImagesService {
	return &ImagesService{
		storage,
		postsRepository,
		postImagesRepository,
	}
}

type ImagesService struct {
	storage              storage.ImageStorage
	postsRepository      PostsRepository
	postImagesRepository PostImagesRepository
}

func NewAvatarService(
	storage storage.ImageStorage,
	usersRepository UsersRepository,
) *AvatarService {
	return &AvatarService{
		storage,
		usersRepository,
	}
}

type AvatarService struct {
	storage         storage.ImageStorage
	usersRepository UsersRepository
}

type PostsRepository interface {
	GetPost(
		ctx context.Context,
		postID int,
	) (domain.Post, error)
}

type PostImagesRepository interface {
	CreatePostImage(
		ctx context.Context,
		image domain.PostImage,
	) (domain.PostImage, error)
	GetByPostID(
		ctx context.Context,
		postID int,
	) ([]domain.PostImage, error)
	DeleteByPostID(
		ctx context.Context,
		postID int,
	) error
}

type UsersRepository interface {
	UpdateAvatarURL(
		ctx context.Context,
		userID int,
		avatarURL string,
	) error
}
