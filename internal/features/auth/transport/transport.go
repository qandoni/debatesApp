package auth_http_transport

import (
	"context"

	"github.com/gin-gonic/gin"
	auth_contracts "github.com/qandoni/debatesApp/internal/features/auth/contracts"
)

type AuthHTTPHandler struct {
	authService AuthService
}

type AuthService interface {
	Login(
		ctx context.Context,
		input auth_contracts.LoginInput,
	) (auth_contracts.LoginOutput, error)
	Registrate(
		ctx context.Context,
		input auth_contracts.RegistrationInput,
	) error
}

func NewAuthHTTPHandler(
	authService AuthService,
) *AuthHTTPHandler {
	return &AuthHTTPHandler{
		authService: authService,
	}
}

func (h *AuthHTTPHandler) Register(rg *gin.RouterGroup) {
	rg.POST("/login", h.Login)
	rg.POST("/registrate", h.Registrate)
}
