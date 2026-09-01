package posts_http_transport

import (
	"net/http"

	"github.com/gin-gonic/gin"
	core_http_request "github.com/qandoni/debatesApp/internal/core/transport/http/request"
)

type GetPostResponse PostDTOResponse

func (h *PostsHTTPHandler) GetPost(c *gin.Context) {
	ctx := c.Request.Context()

	// authInfo, ok := core_auth.AuthInfoFromContext(ctx)
	// if !ok {
	// 	c.Error(core_errors.ErrAccessForbidden).SetMeta("no auth info in request context")
	// 	return
	// }
	postID, err := core_http_request.GetIntPathValue(c, "id")
	if err != nil {
		c.Error(err).SetMeta("failed to get 'id' int path value")
		return
	}
	postDomain, err := h.postsService.GetPost(ctx, postID)
	if err != nil {
		c.Error(err).SetMeta("failed to get post")
		return
	}
	response := GetPostResponse(postDTOFromDomain(postDomain))
	c.JSON(http.StatusOK, response)
}
