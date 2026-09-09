package comments_transport_http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	core_http_request "github.com/qandoni/debatesApp/internal/core/transport/http/request"
	comments_dto "github.com/qandoni/debatesApp/internal/features/comments/transport/http/dto"
)

func (h *CommentsHTTPHandler) GetArguments(c *gin.Context) {
	postID, err := core_http_request.GetIntPathValue(c, "id")
	if err != nil {
		c.Error(err).SetMeta("failed to get 'id' path value")
		return
	}

	limit, offset, err := core_http_request.GetLimitOffsetQueryParams(c)
	if err != nil {
		c.Error(err).SetMeta("failed to get limit/offset query params")
		return
	}

	arguments, comments, err := h.commentsService.GetArgumentsWithReplies(
		c.Request.Context(),
		postID,
		limit,
		offset,
	)
	if err != nil {
		c.Error(err).SetMeta("failed to get arguments")
		return
	}

	response := comments_dto.BuildArgumentTree(
		arguments,
		comments,
	)

	c.JSON(http.StatusOK, response)
}
