package comments_transport_http

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/qandoni/debatesApp/internal/core/domain"
)

func NewCommentsHTTPHandler(
	commentsService CommentsService,
) *CommentsHTTPHandler {
	return &CommentsHTTPHandler{
		commentsService,
	}
}

type CommentsHTTPHandler struct {
	commentsService CommentsService
}

type CommentsService interface {
	CreateComment(
		ctx context.Context,
		userID int,
		postID int,
		parentCommentID *int,
		content string,
	) (domain.Comment, error)

	CreateArgument(
		ctx context.Context,
		userID int,
		postID int,
		debateSideID int,
		content string,
	) (domain.Comment, error)

	GetByPostID(
		ctx context.Context,
		postID int,
		limit *int,
		offset *int,
	) ([]domain.Comment, error)

	UpdateComment(
		ctx context.Context,
		userID int,
		commentID int,
		content string,
	) (domain.Comment, error)
	GetArgumentsWithReplies(
		ctx context.Context,
		postID int,
		limit *int,
		offset *int,
	) ([]domain.CommentWithRating, []domain.Comment, error)
	SetAuthorLike(
		ctx context.Context,
		userID int,
		commentID int,
		liked bool,
	) (domain.Comment, error)
}

func (h *CommentsHTTPHandler) Register(rg *gin.RouterGroup) {
	posts := rg.Group("/posts")
	posts.POST("/:id/comments", h.CreateComment)
	posts.GET("/:id/comments", h.GetComments)
	posts.POST("/:id/arguments", h.CreateArgument)
	posts.GET("/:id/arguments", h.GetArguments)

	comments := rg.Group("/comments")
	comments.PATCH("/:id", h.UpdateComment)
	comments.PATCH("/:id/author-like", h.SetAuthorLike)
}
