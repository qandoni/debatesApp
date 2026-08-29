package users_http_transport

import (
	"context"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/qandoni/debatesApp/internal/core/domain"
)

type UsersHTTPHandler struct {
	usersService  UsersService
	imagesService ImagesService
}

type UsersService interface {
	GetMyProfile(ctx context.Context, userID int) (domain.User, error)
	EditProfile(ctx context.Context, userID int, profilePatch domain.UserPatch) (domain.User, error)
}
type ImagesService interface {
	UploadAvatar(
		ctx context.Context,
		userID int,
		reader io.Reader,
		size int64,
		contentType string,
		extension string,
	) error
}

func NewUsersHTTPHandler(
	usersService UsersService,
	imagesService ImagesService,
) *UsersHTTPHandler {
	return &UsersHTTPHandler{
		usersService:  usersService,
		imagesService: imagesService,
	}
}

func (h *UsersHTTPHandler) Register(rg *gin.RouterGroup) {
	users := rg.Group("")
	users.GET("/me", h.GetMyProfile)
	users.PATCH("/me", h.EditProfile)
	users.POST("/me/avatar", h.UploadAvatar)
}
