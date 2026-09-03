package debate_votes_http_transport

import (
	"net/http"

	"github.com/gin-gonic/gin"
	core_auth "github.com/qandoni/debatesApp/internal/core/auth"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
	core_http_request "github.com/qandoni/debatesApp/internal/core/transport/http/request"
)

func (h *DebateVotesHTTPHandler) FinishDebate(c *gin.Context) {
	ctx := c.Request.Context()

	authInfo, ok := core_auth.AuthInfoFromContext(ctx)
	if !ok {
		c.Error(core_errors.ErrAccessForbidden).
			SetMeta("no auth info in request context")
		return
	}
	debateId, err := core_http_request.GetIntPathValue(c, "id")
	if err != nil {
		c.Error(core_errors.ErrInvalidArgument).SetMeta("invalid debate id")
		return
	}
	if err := h.debateVotesService.FinishDebate(ctx, authInfo.UserID, debateId); err != nil {
		c.Error(err).SetMeta("failed to finish debate")
		return
	}
	c.Status(http.StatusNoContent)

}
