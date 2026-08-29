package auth_http_transport

import (
	"net/http"

	"github.com/gin-gonic/gin"
	core_http_request "github.com/qandoni/debatesApp/internal/core/transport/http/request"
	auth_contracts "github.com/qandoni/debatesApp/internal/features/auth/contracts"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
}

func (h *AuthHTTPHandler) Login(c *gin.Context) {
	ctx := c.Request.Context()

	var request LoginRequest
	if err := core_http_request.DecodeAndValidateRequest(c, &request); err != nil {
		c.Error(err).SetMeta("failed to decode and validate HTTP request")
		return
	}

	input := auth_contracts.LoginInput{
		Email:    request.Email,
		Password: request.Password,
	}
	output, err := h.authService.Login(
		ctx,
		input,
	)
	if err != nil {
		c.Error(err).SetMeta("failed to login")
		return
	}
	response := LoginResponse{
		AccessToken: output.AccessToken,
	}
	c.JSON(http.StatusOK, response)
}
