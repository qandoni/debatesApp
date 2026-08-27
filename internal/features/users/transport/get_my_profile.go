package users_http_transport

import (
	"net/http"

	"github.com/gin-gonic/gin"
	core_auth "github.com/qandoni/debatesApp/internal/core/auth"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
)

type GetMyProfileResponse UserDTOResponse

func (h *UsersHTTPHandler) GetMyProfile(c *gin.Context) {
	ctx := c.Request.Context()

	authInfo, ok := core_auth.AuthInfoFromContext(ctx)
	if !ok {
		c.Error(core_errors.ErrNotFound).SetMeta("no user auth info in JW token")
		return
	}
	userDomain, err := h.usersService.GetMyProfile(ctx, authInfo.UserID)
	if err != nil {
		c.Error(err).SetMeta("failed to get profile")
		return
	}
	response := GetMyProfileResponse(UserDTOFromDomain(userDomain))
	c.JSON(http.StatusOK, response)
}
