package posts_http_transport

import (
	"net/http"

	"github.com/gin-gonic/gin"
	core_auth "github.com/qandoni/debatesApp/internal/core/auth"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
	core_http_request "github.com/qandoni/debatesApp/internal/core/transport/http/request"
)

func (h *PostImagesHTTPHandler) CreatePostImages(c *gin.Context) {
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

	form, err := c.MultipartForm()
	if err != nil {
		c.Error(err).SetMeta("failed to get attached files")
		return
	}

	files := form.File["images"]
	if len(files) == 0 {
		c.Error(err).SetMeta("files not given")
		return
	}
	images, err := h.imagesService.CreatePostImages(
		ctx,
		authInfo.UserID,
		postID,
		files,
	)
	if err != nil {
		c.Error(err).SetMeta("failed to create post image")
		return
	}
	c.JSON(http.StatusCreated, images)

}
