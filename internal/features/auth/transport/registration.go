package auth_http_transport

import (
	"net/http"

	"github.com/gin-gonic/gin"
	core_http_request "github.com/qandoni/debatesApp/internal/core/transport/http/request"
	auth_contracts "github.com/qandoni/debatesApp/internal/features/auth/contracts"
)

type RegistrationRequest struct {
	UserName string `json:"user_name"`
	Email    string `json:"email" validate:"email"`
	Password string `json:"password"`
}

func (h *AuthHTTPHandler) Registrate(c *gin.Context) {
	ctx := c.Request.Context()

	var request RegistrationRequest
	if err := core_http_request.DecodeAndValidateRequest(c, &request); err != nil {
		c.Error(err).SetMeta("failed to decode and validate HTTP request")
		return
	}
	input := auth_contracts.RegistrationInput{
		UserName: request.UserName,
		Email:    request.Email,
		Password: request.Password,
	}
	if err := h.authService.Registrate(ctx, input); err != nil {
		c.Error(err).SetMeta("failed to registrate the user")
		return
	}
	c.Status(http.StatusCreated)
}
