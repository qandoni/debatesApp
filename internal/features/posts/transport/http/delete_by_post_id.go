package posts_http_transport

import (
	"net/http"

	"github.com/gin-gonic/gin"
	core_auth "github.com/qandoni/debatesApp/internal/core/auth"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
	core_http_request "github.com/qandoni/debatesApp/internal/core/transport/http/request"
)

func (h *PostImagesHTTPHandler) DeleteByPostID(c *gin.Context) {
	ctx := c.Request.Context()

	authInfo, ok := core_auth.AuthInfoFromContext(ctx)
	if !ok {
		c.Error(core_errors.ErrAccessForbidden).SetMeta("no auth info in request context")
		return
	}
	postID, err := core_http_request.GetIntPathValue(c, "id")
	if err != nil {
		c.Error(err).SetMeta("failed to get 'id' int path value")
		return
	}
	if err := h.imagesService.DeleteByPostID(ctx, authInfo.UserID, postID); err != nil {
		c.Error(err).SetMeta("failed to delete by post id")
		return
	}
	c.Status(http.StatusNoContent)

}
