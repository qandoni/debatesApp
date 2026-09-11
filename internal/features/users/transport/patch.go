package users_http_transport

import (
	"net/http"

	"github.com/gin-gonic/gin"
	core_auth "github.com/qandoni/debatesApp/internal/core/auth"
	"github.com/qandoni/debatesApp/internal/core/domain"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
	core_http_request "github.com/qandoni/debatesApp/internal/core/transport/http/request"
)

type EditProfileRequest = PatchUserRequest

type EditProfileResponse UserDTOResponse

func (h *UsersHTTPHandler) EditProfile(c *gin.Context) {
	ctx := c.Request.Context()

	authInfo, ok := core_auth.AuthInfoFromContext(ctx)
	if !ok {
		c.Error(core_errors.ErrNotFound).SetMeta("no user auth info in JW token")
		return
	}

	var request EditProfileRequest
	if err := core_http_request.DecodeAndValidateRequest(c, &request); err != nil {
		c.Error(err).SetMeta("failed to decode and validate HTTP request")
		return
	}

	userProfilePatch := profilePatchFromRequest(request)

	userDomain, err := h.usersService.EditProfile(ctx, authInfo.UserID, userProfilePatch)
	if err != nil {
		c.Error(err).SetMeta("failed to edit profile")
		return
	}
	response := EditProfileResponse(UserDTOFromDomain(userDomain))
	c.JSON(http.StatusOK, response)
}

func profilePatchFromRequest(request EditProfileRequest) domain.UserPatch {
	return domain.NewUserPatch(
		request.Username.ToDomain(),
		request.Email.ToDomain(),
		request.Password.ToDomain(),
		request.Bio.ToDomain(),
	)
}
