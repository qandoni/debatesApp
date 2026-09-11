package users_http_transport

import (
	"context"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/qandoni/debatesApp/internal/core/domain"
)

type UsersHTTPHandler struct {
	usersService  UsersService
	avatarService AvatarService
	jwt           gin.HandlerFunc
}

type UsersService interface {
	GetMyProfile(ctx context.Context, userID int) (domain.User, error)
	EditProfile(ctx context.Context, userID int, profilePatch domain.UserPatch) (domain.User, error)
}
type AvatarService interface {
	UploadAvatar(
		ctx context.Context,
		userID int,
		reader io.Reader,
		size int64,
	) error
}

func NewUsersHTTPHandler(
	usersService UsersService,
	avatarService AvatarService,
	jwt gin.HandlerFunc,
) *UsersHTTPHandler {
	return &UsersHTTPHandler{
		usersService:  usersService,
		avatarService: avatarService,
		jwt:           jwt,
	}
}

func (h *UsersHTTPHandler) Register(rg *gin.RouterGroup) {
	users := rg.Group("/users")
	users.Use(h.jwt)
	users.GET("/me", h.GetMyProfile)
	users.PATCH("/me", h.EditProfile)
	users.POST("/me/avatar", h.UploadAvatar)
}
