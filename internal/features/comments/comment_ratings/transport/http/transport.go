package comment_ratings_http_transport

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/qandoni/debatesApp/internal/core/domain"
)

func NewCommentRatingsHTTPHandler(
	commentRatingsService CommentRatingsService,
) *CommentRatingsHTTPHandler {
	return &CommentRatingsHTTPHandler{
		commentRatingsService,
	}
}

type CommentRatingsHTTPHandler struct {
	commentRatingsService CommentRatingsService
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
	rg.PUT("/:id/rating", h.RateComment)
}
