package debate_votes_http_transport

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/qandoni/debatesApp/internal/core/domain"
)

func NewDebateVotesHTTPTransport(
	debateVotesService DebateVotesService,
) *DebateVotesHTTPHandler {
	return &DebateVotesHTTPHandler{
		debateVotesService,
	}
}

type DebateVotesHTTPHandler struct {
	debateVotesService DebateVotesService
}

type DebateVotesService interface {
	Vote(ctx context.Context, userID int, debateID int, debateSideID int) (domain.DebateVote, error)
	ChangeVote(ctx context.Context, userID int, debateID int, debateSideID int) (domain.DebateVote, error)
}

type VoteRequest struct {
	DebateSideID int `json:"debate_side_id" validate:"required,min=1"`
}

func (h *DebateVotesHTTPHandler) Register(rg *gin.RouterGroup) {
	rg.POST("/:id/vote", h.Vote)
	rg.PATCH("/:id/vote", h.ChangeVote)
}
