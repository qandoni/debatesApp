package posts_http_transport

import (
	"net/http"

	"github.com/gin-gonic/gin"
	core_auth "github.com/qandoni/debatesApp/internal/core/auth"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
	core_http_request "github.com/qandoni/debatesApp/internal/core/transport/http/request"
	posts_contracts "github.com/qandoni/debatesApp/internal/features/posts/contracts"
)

type CreatePostRequest struct {
	Content  string `json:"content"`
	IsDebate bool   `json:"is_debate" required:"true"`
	//TODO добавить возможность прикреплять изображения к посту
}

type CreatePostResponse PostDTOResponse

func (h *PostsHTTPHandler) CreatePost(c *gin.Context) {
	ctx := c.Request.Context()

	authInfo, ok := core_auth.AuthInfoFromContext(ctx)
	if !ok {
		c.Error(core_errors.ErrAccessForbidden).SetMeta("no auth info in request context")
		return
	}
	var request CreatePostRequest

	if err := core_http_request.DecodeAndValidateRequest(c, &request); err != nil {
		c.Error(err).SetMeta("failed to decode and validate HTTP request")
		return
	}
	input := posts_contracts.CreatePostInput{
		AuthorID: authInfo.UserID,
		Content:  request.Content,
		IsDebate: request.IsDebate,
	}
	post, err := h.postsService.CreatePost(ctx, input)
	if err != nil {
		c.Error(err).SetMeta("failed to create post")
		return
	}
	response := CreatePostResponse(postDTOFromDomain(post))
	c.JSON(http.StatusCreated, response)

}
