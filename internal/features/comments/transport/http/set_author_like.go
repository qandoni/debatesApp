package comments_transport_http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	core_auth "github.com/qandoni/debatesApp/internal/core/auth"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
	core_http_request "github.com/qandoni/debatesApp/internal/core/transport/http/request"
	comments_dto "github.com/qandoni/debatesApp/internal/features/comments/transport/http/dto"
)

type SetAuthorLikeRequest struct {
	Liked bool `json:"liked"`
}

func (h *CommentsHTTPHandler) SetAuthorLike(c *gin.Context) {
	ctx := c.Request.Context()
	commentID, err := core_http_request.GetIntPathValue(c, "id")
	if err != nil {
		c.Error(err).SetMeta("failed to get comment id")
		return
	}

	var request SetAuthorLikeRequest

	if err := core_http_request.DecodeAndValidateRequest(
		c,
		&request,
	); err != nil {
		c.Error(err).SetMeta("failed to decode author like request")
		return
	}

	authInfo, ok := core_auth.AuthInfoFromContext(ctx)
	if !ok {
		c.Error(core_errors.ErrAccessForbidden).SetMeta("no auth info in request context")
		return
	}

	comment, err := h.commentsService.SetAuthorLike(
		c.Request.Context(),
		authInfo.UserID,
		commentID,
		request.Liked,
	)
	if err != nil {
		c.Error(err).SetMeta("failed to set author like")
		return
	}

	response := comments_dto.NewCommentDTOFromDomain(comment)

	c.JSON(http.StatusOK, response)
}
