package posts_http_transport

import (
	"net/http"

	"github.com/gin-gonic/gin"
	core_http_request "github.com/qandoni/debatesApp/internal/core/transport/http/request"
)

type GetPostsResponse []PostDTOResponse

func (h *PostsHTTPHandler) GetPosts(c *gin.Context) {
	ctx := c.Request.Context()

	limit, offset, err := core_http_request.GetLimitOffsetQueryParams(c)
	if err != nil {
		c.Error(err).SetMeta("failed to get 'limit'/'offset' query params")
		return
	}

	postsDomains, err := h.postsService.GetPosts(ctx, limit, offset)
	if err != nil {
		c.Error(err).SetMeta("failed to get posts")
		return
	}
	response := GetPostsResponse(postsDTOFromDomains(postsDomains))
	c.JSON(http.StatusOK, response)
}
