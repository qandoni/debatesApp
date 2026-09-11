package comment_ratings_http_transport

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/qandoni/debatesApp/internal/core/domain"
)

func NewCommentRatingsHTTPHandler(
	commentRatingsService CommentRatingsService,
	jwt gin.HandlerFunc,
) *CommentRatingsHTTPHandler {
	return &CommentRatingsHTTPHandler{
		commentRatingsService,
		jwt,
	}
}

type CommentRatingsHTTPHandler struct {
	commentRatingsService CommentRatingsService
	jwt                   gin.HandlerFunc
}

type CommentRatingsService interface {
	Rate(
		ctx context.Context,
		userID int,
		commentID int,
		score int,
	) (domain.CommentRating, error)
}

func (h *CommentRatingsHTTPHandler) Register(
	rg *gin.RouterGroup,
) {
	comments := rg.Group("/comments")
	comments.Use(h.jwt)
	comments.PUT("/:id/rating", h.RateComment)
}
