package posts_http_transport

import (
	"context"
	"mime/multipart"

	"github.com/gin-gonic/gin"
	"github.com/qandoni/debatesApp/internal/core/domain"
	posts_contracts "github.com/qandoni/debatesApp/internal/features/posts/contracts"
)

type PostsHTTPHandler struct {
	postsService PostsService
}

type PostImagesHTTPHandler struct {
	imagesService ImagesService
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

type PostsService interface {
	CreatePost(
		ctx context.Context,
		input posts_contracts.CreatePostInput,
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

func NewPostImagesHTTPHandler(
	imagesService ImagesService,
) *PostImagesHTTPHandler {
	return &PostImagesHTTPHandler{
		imagesService,
	}
}
func NewPostsHTTPHandler(
	postsService PostsService,
) *PostsHTTPHandler {
	return &PostsHTTPHandler{
		postsService: postsService,
	}
}

func (h *PostsHTTPHandler) Register(rg *gin.RouterGroup) {
	posts := rg.Group("")
	posts.POST("", h.CreatePost)
	posts.GET("/:id", h.GetPost)
	posts.PATCH("/:id", h.PatchPost)
	posts.DELETE("/:id", h.DeletePost)
	posts.GET("", h.GetPosts)
}

func (h *PostImagesHTTPHandler) Register(rg *gin.RouterGroup) {
	posts := rg.Group("")
	posts.GET("/:id/images", h.GetByPostID)
	posts.POST("/:id/images", h.CreatePostImages)
	posts.DELETE("/:id/images", h.DeleteByPostID)
}
