package users_http_transport

import (
	"io"
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

	const maxAvatarSize = 5 * 1024 * 1024

	if fileHeader.Size > maxAvatarSize {
		c.Error(core_errors.ErrInvalidArgument).SetMeta("file is too large")
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		c.Error(core_errors.ErrInternalError).SetMeta("some internal error during open file")
		return
	}
	defer file.Close()
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil {
		c.Error(core_errors.ErrInternalError).SetMeta("some internal error during read file")
		return
	}

	contentType := http.DetectContentType(buffer[:n])

	extension, ok := extensionByContentType(contentType)
	if !ok {
		c.Error(core_errors.ErrUnsupportedMediaType)
		return
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		c.Error(err)
		return
	}
	if err := h.imagesService.UploadAvatar(
		ctx,
		authInfo.UserID,
		file,
		fileHeader.Size,
		contentType,
		extension,
	); err != nil {
		c.Error(err).SetMeta("can't upload avatar")
		return
	}
	c.Status(http.StatusNoContent)
}

func extensionByContentType(contentType string) (string, bool) {
	switch contentType {
	case "image/jpeg":
		return ".jpg", true
	case "image/png":
		return ".png", true
	case "image/webp":
		return ".webp", true
	default:
		return "", false
	}
}
