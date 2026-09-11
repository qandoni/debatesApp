package images_service

import (
	"context"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_postgres "github.com/qandoni/debatesApp/internal/core/repository/postgres"
	"github.com/qandoni/debatesApp/internal/features/storage"
)

func NewImagesService(
	storage storage.ImageStorage,
	postsRepository PostsRepository,
	postImagesRepository PostImagesRepository,
	txManager core_postgres.TransactionManager,
) *ImagesService {
	return &ImagesService{
		storage,
		postsRepository,
		postImagesRepository,
		txManager,
	}
}

type ImagesService struct {
	storage              storage.ImageStorage
	postsRepository      PostsRepository
	postImagesRepository PostImagesRepository
	txManager            core_postgres.TransactionManager
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
	GetByPostIDs(
		ctx context.Context,
		postIDs []int,
	) (map[int][]domain.PostImage, error)
}

type UsersRepository interface {
	UpdateAvatarURL(
		ctx context.Context,
		userID int,
		avatarURL string,
	) error
	GetAvatarURL(
		ctx context.Context,
		userID int,
	) (*string, error)
}
