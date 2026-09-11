package posts_http_transport

import (
	"net/http"

	"github.com/gin-gonic/gin"
	core_http_request "github.com/qandoni/debatesApp/internal/core/transport/http/request"
)

type GetPostResponse PostDTOResponse

func (h *PostsHTTPHandler) GetPost(c *gin.Context) {
	ctx := c.Request.Context()

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
