package users_http_transport

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/qandoni/debatesApp/internal/core/domain"
)

type UsersHTTPHandler struct {
	usersService UsersService
}

type UsersService interface {
	GetMyProfile(ctx context.Context, userID int) (domain.User, error)
}

func NewUsersHTTPHandler(
	usersService UsersService,
) *UsersHTTPHandler {
	return &UsersHTTPHandler{
		usersService: usersService,
	}
}

func (h *UsersHTTPHandler) Register(rg *gin.RouterGroup) {
	users := rg.Group("")
	users.GET("/me", h.GetMyProfile)
}
