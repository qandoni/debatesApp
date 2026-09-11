package posts_http_transport

import (
	"net/http"

	"github.com/gin-gonic/gin"
	core_auth "github.com/qandoni/debatesApp/internal/core/auth"
	"github.com/qandoni/debatesApp/internal/core/domain"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
	core_http_request "github.com/qandoni/debatesApp/internal/core/transport/http/request"
	core_http_types "github.com/qandoni/debatesApp/internal/core/transport/http/types"
)

type PatchPostRequest struct {
	Content core_http_types.Nullable[string] `json:"content"`
}

type PatchPostResponse PostDTOResponse

func (h *PostsHTTPHandler) PatchPost(c *gin.Context) {
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

	var request PatchPostRequest
	if err := core_http_request.DecodeAndValidateRequest(c, &request); err != nil {
		c.Error(err).SetMeta("failed to decode and validate HTTP request")
		return
	}
	postPatch := postPatchFromRequest(request)

	postDomain, err := h.postsService.PatchPost(ctx, authInfo.UserID, postID, postPatch)
	if err != nil {
		c.Error(err).SetMeta("failed to patch user")
		return
	}
	response := PatchPostResponse(postDTOFromDomain(postDomain))
	c.JSON(http.StatusOK, response)

}

func postPatchFromRequest(request PatchPostRequest) domain.PostPatch {
	return domain.NewPostPatch(
		request.Content.ToDomain(),
	)
}
