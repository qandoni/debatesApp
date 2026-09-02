package posts_http_transport

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	core_auth "github.com/qandoni/debatesApp/internal/core/auth"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
	core_http_request "github.com/qandoni/debatesApp/internal/core/transport/http/request"
	posts_contracts "github.com/qandoni/debatesApp/internal/features/posts/contracts"
)

type CreatePostRequest struct {
	Content  string               `json:"content"`
	IsDebate bool                 `json:"is_debate"`
	Debate   *CreateDebateRequest `json:"debate,omitempty"`
}

type CreateDebateSideRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
}

type CreateDebateRequest struct {
	EndAt *time.Time                `json:"end_at,omitempty"`
	Sides []CreateDebateSideRequest `json:"sides" validate:"required,min=2,max=5,dive"`
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
	var debateInput *posts_contracts.CreateDebateInput

	if request.Debate != nil {
		sides := make([]posts_contracts.CreateDebateSideInput, 0, len(request.Debate.Sides))

		for _, side := range request.Debate.Sides {
			sides = append(sides, posts_contracts.CreateDebateSideInput{
				Name:        side.Name,
				Description: side.Description,
			})
		}

		debateInput = &posts_contracts.CreateDebateInput{
			EndAt: request.Debate.EndAt,
			Sides: sides,
		}
	}

	input := posts_contracts.CreatePostInput{
		AuthorID: authInfo.UserID,
		Content:  request.Content,
		IsDebate: request.IsDebate,
		Debate:   debateInput,
	}
	post, err := h.postsService.CreatePost(ctx, input)
	if err != nil {
		c.Error(err).SetMeta("failed to create post")
		return
	}
	response := CreatePostResponse(postDTOFromDomain(post))
	c.JSON(http.StatusCreated, response)

}
