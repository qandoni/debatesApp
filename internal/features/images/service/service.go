package images_service

import (
	"context"

	"github.com/qandoni/debatesApp/internal/core/domain"
	"github.com/qandoni/debatesApp/internal/features/storage"
)

func NewImagesService(
	storage storage.ImageStorage,
	usersRepository UsersRepository,
) *ImagesService {
	return &ImagesService{
		storage,
		usersRepository,
	}
}

type ImagesService struct {
	storage         storage.ImageStorage
	usersRepository UsersRepository
	//TODO добавить PostImages repository
}

type UsersRepository interface {
	GetMyProfile(
		ctx context.Context,
		userID int,
	) (domain.User, error)

	UpdateAvatarURL(
		ctx context.Context,
		userID int,
		avatarURL string,
	) error
}
