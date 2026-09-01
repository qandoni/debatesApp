package posts_service

import (
	"context"
	"mime/multipart"

	"github.com/qandoni/debatesApp/internal/core/domain"
)

type PostsService struct {
	postsRepository PostsRepository
	imagesService   ImagesService
	//TODO добавить транзакции для взаимодействия с будущей фичой дебатов
}
type PostsRepository interface {
	CreatePost(
		ctx context.Context,
		post domain.Post,
	) (domain.Post, error)
	DeletePost(
		ctx context.Context,
		userID int,
		postID int,
	) error
	GetPost(
		ctx context.Context,
		postID int,
	) (domain.Post, error)
	PatchPost(
		ctx context.Context,
		userID int,
		postID int,
		postPatch domain.PostPatch,
	) (domain.Post, error)
	GetPosts(
		ctx context.Context,
		limit *int,
		offset *int,
	) ([]domain.Post, error)
}

type ImagesService interface {
	GetByPostID(
		ctx context.Context,
		postID int,
	) ([]domain.PostImage, error)
	CreatePostImages(
		ctx context.Context,
		userID int,
		postID int,
		files []*multipart.FileHeader,
	) ([]domain.PostImage, error)
	DeleteByPostID(
		ctx context.Context,
		userID int,
		postID int,
	) error
}

func NewPostsService(
	postsRepository PostsRepository,
	imagesService ImagesService,
) *PostsService {
	return &PostsService{
		postsRepository: postsRepository,
		imagesService:   imagesService,
	}
}
