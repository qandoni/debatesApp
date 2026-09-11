package users_http_transport

import (
	"net/http"

	"github.com/gin-gonic/gin"
	core_auth "github.com/qandoni/debatesApp/internal/core/auth"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
)

func (h *UsersHTTPHandler) UploadAvatar(c *gin.Context) {
	ctx := c.Request.Context()

	authInfo, ok := core_auth.AuthInfoFromContext(ctx)
	if !ok {
		c.Error(core_errors.ErrNotFound).SetMeta("no user authInfo in context")
		return
	}

	fileHeader, err := c.FormFile("avatar")
	if err != nil {
		c.Error(core_errors.ErrNotFound).SetMeta("file header not found")
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.Error(core_errors.ErrInternalError).
			SetMeta("some internal error during open file")
		return
	}
	defer file.Close()

	if err := h.avatarService.UploadAvatar(
		ctx,
		authInfo.UserID,
		file,
		fileHeader.Size,
	); err != nil {
		c.Error(err).SetMeta("can't upload avatar")
		return
	}

	c.Status(http.StatusNoContent)
}
