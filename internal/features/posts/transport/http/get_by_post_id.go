package posts_http_transport

import (
	"net/http"

	"github.com/gin-gonic/gin"
	core_http_request "github.com/qandoni/debatesApp/internal/core/transport/http/request"
)

func (h *PostImagesHTTPHandler) GetByPostID(c *gin.Context) {
	ctx := c.Request.Context()

	postID, err := core_http_request.GetIntPathValue(c, "id")
	if err != nil {
		c.Error(err).SetMeta("failed to get 'id' path value")
		return
	}

	images, err := h.imagesService.GetByPostID(ctx, postID)
	if err != nil {
		c.Error(err).SetMeta("failed to get by post id")
		return
	}
	c.JSON(http.StatusOK, images)
}
