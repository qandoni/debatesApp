package images_service

import (
	"context"
	"io"
	"time"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_postgres "github.com/qandoni/debatesApp/internal/core/repository/postgres"
	"github.com/qandoni/debatesApp/internal/features/storage"
)

func NewImagesService(
	minio MinioManager,
	postsRepository PostsRepository,
	postImagesRepository PostImagesRepository,
	txManager core_postgres.TransactionManager,
) *ImagesService {
	return &ImagesService{
		minio,
		postsRepository,
		postImagesRepository,
		txManager,
	}
}

type ImagesService struct {
	minio                MinioManager
	postsRepository      PostsRepository
	postImagesRepository PostImagesRepository
	txManager            core_postgres.TransactionManager
}

func NewAvatarService(
	minio MinioManager,
	usersRepository UsersRepository,
) *AvatarService {
	return &AvatarService{
		minio,
		usersRepository,
	}
}

type AvatarService struct {
	minio           MinioManager
	usersRepository UsersRepository
}

type MinioManager interface {
	Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error
	Begin() *storage.Tx
	GetURL(ctx context.Context, objectName string) (string, error)
	ObjectExists(ctx context.Context, objectName string) (bool, error)
	DeleteAll(ctx context.Context, objectNames []string) []string
	NewObjectName(prefix, extension string) string
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
	MarkActiveByIDs(
		ctx context.Context,
		imageIDs []int,
	) error
	DeletePendingByIDs(
		ctx context.Context,
		imageIDs []int,
	) error
	GetPending(
		ctx context.Context,
	) ([]domain.PostImage, error)
	GetPendingOlderThan(
		ctx context.Context,
		staleAfter time.Duration,
	) ([]domain.PostImage, error)
	GetImagesOfDeletedPosts(
		ctx context.Context,
	) ([]domain.PostImage, error)
}

type UsersRepository interface {
	UpdateAvatarURL(
		ctx context.Context,
		userID int,
		avatarURL string,
	) (*string, error)
}
