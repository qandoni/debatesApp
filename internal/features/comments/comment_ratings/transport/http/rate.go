package comment_ratings_http_transport

import (
	"net/http"

	"github.com/gin-gonic/gin"
	core_auth "github.com/qandoni/debatesApp/internal/core/auth"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
	core_http_request "github.com/qandoni/debatesApp/internal/core/transport/http/request"
)

type RateCommentRequest struct {
	Score int `json:"score" validate:"required,min=1,max=5"`
}

func (h *CommentRatingsHTTPHandler) RateComment(c *gin.Context) {
	ctx := c.Request.Context()
	commentID, err := core_http_request.GetIntPathValue(c, "id")
	if err != nil {
		c.Error(err).SetMeta("failed to get 'id' path value")
		return
	}

	var request RateCommentRequest

	if err := core_http_request.DecodeAndValidateRequest(
		c,
		&request,
	); err != nil {
		c.Error(err).SetMeta("failed to decode rate comment request")
		return
	}

	authInfo, ok := core_auth.AuthInfoFromContext(ctx)
	if !ok {
		c.Error(core_errors.ErrAccessForbidden).SetMeta("no auth info in request context")
		return
	}

	rating, err := h.commentRatingsService.Rate(
		c.Request.Context(),
		authInfo.UserID,
		commentID,
		request.Score,
	)
	if err != nil {
		c.Error(err).SetMeta("failed to rate comment")
		return
	}

	c.JSON(http.StatusOK, rating)
}
