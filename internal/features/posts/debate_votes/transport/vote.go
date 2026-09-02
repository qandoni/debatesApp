package debate_votes_http_transport

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	core_auth "github.com/qandoni/debatesApp/internal/core/auth"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
	core_http_request "github.com/qandoni/debatesApp/internal/core/transport/http/request"
)

func (h *DebateVotesHTTPHandler) Vote(c *gin.Context) {
	ctx := c.Request.Context()

	authInfo, ok := core_auth.AuthInfoFromContext(ctx)
	if !ok {
		c.Error(core_errors.ErrAccessForbidden).
			SetMeta("no auth info in request context")
		return
	}

	debateID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Error(core_errors.ErrInvalidArgument).
			SetMeta("invalid debate id")
		return
	}

	var request VoteRequest

	if err := core_http_request.DecodeAndValidateRequest(c, &request); err != nil {
		c.Error(err).SetMeta(
			"failed to decode and validate HTTP request",
		)
		return
	}

	vote, err := h.debateVotesService.Vote(
		ctx,
		authInfo.UserID,
		debateID,
		request.DebateSideID,
	)
	if err != nil {
		c.Error(err).SetMeta("failed to vote in debate")
		return
	}

	response := debateVoteResponseFromDomain(vote)

	c.JSON(http.StatusCreated, response)
}
