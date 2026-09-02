package posts_service

import (
	"context"
	"mime/multipart"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_postgres "github.com/qandoni/debatesApp/internal/core/repository/postgres"
)

type PostsService struct {
	postsRepository       PostsRepository
	imagesService         ImagesService
	debatesRepository     DebatesRepository
	debateSidesRepository DebateSidesRepository
	txManager             core_postgres.TransactionManager
}

type DebateSidesRepository interface {
	CreateDebateSide(
		ctx context.Context,
		side domain.DebateSide,
	) (domain.DebateSide, error)

	GetByDebateID(
		ctx context.Context,
		debateID int,
	) ([]domain.DebateSide, error)
}

type DebatesRepository interface {
	CreateDebate(
		ctx context.Context,
		debate domain.Debate,
	) (domain.Debate, error)

	GetByPostID(
		ctx context.Context,
		postID int,
	) (domain.Debate, error)

	GetByID(
		ctx context.Context,
		debateID int,
	) (domain.Debate, error)

	// FinishDebate(
	// 	ctx context.Context,
	// 	debateID int,
	// 	winnerSideID int,
	// ) error
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
	debatesRepository DebatesRepository,
	debateSidesRepository DebateSidesRepository,
	txManager core_postgres.TransactionManager,
) *PostsService {
	return &PostsService{
		postsRepository:       postsRepository,
		imagesService:         imagesService,
		debatesRepository:     debatesRepository,
		debateSidesRepository: debateSidesRepository,
		txManager:             txManager,
	}
}
