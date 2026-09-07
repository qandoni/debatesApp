package comments_transport_http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	core_auth "github.com/qandoni/debatesApp/internal/core/auth"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
	core_http_request "github.com/qandoni/debatesApp/internal/core/transport/http/request"
	comments_dto "github.com/qandoni/debatesApp/internal/features/comments/transport/http/dto"
)

type CreateArgumentRequest struct {
	DebateSideID int    `json:"debate_side_id" validate:"required,gt=0"`
	Content      string `json:"content" validate:"required,min=1,max=5000"`
}

func (h *CommentsHTTPHandler) CreateArgument(c *gin.Context) {
	ctx := c.Request.Context()

	authInfo, ok := core_auth.AuthInfoFromContext(ctx)
	if !ok {
		c.Error(core_errors.ErrAccessForbidden).SetMeta("no auth info in request context")
		return
	}

	postID, err := core_http_request.GetIntPathValue(c, "id")
	if err != nil {
		c.Error(err).SetMeta("failed to get 'id' path value")
		return
	}

	var req CreateArgumentRequest

	if err := core_http_request.DecodeAndValidateRequest(c, &req); err != nil {
		c.Error(err).SetMeta("failed to decode and validate request")
		return
	}

	comment, err := h.commentsService.CreateArgument(
		c.Request.Context(),
		authInfo.UserID,
		postID,
		req.DebateSideID,
		req.Content,
	)
	if err != nil {
		c.Error(err).SetMeta("failed to create agrument")
		return
	}
	response := comments_dto.NewCommentDTOFromDomain(comment)

	c.JSON(http.StatusCreated, response)
}
