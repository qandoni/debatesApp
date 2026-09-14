package debate_votes_http_transport

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/qandoni/debatesApp/internal/core/domain"
)

func NewDebateVotesHTTPTransport(
	debateVotesService DebateVotesService,
	jwt gin.HandlerFunc,
) *DebateVotesHTTPHandler {
	return &DebateVotesHTTPHandler{
		debateVotesService,
		jwt,
	}
}

type DebateVotesHTTPHandler struct {
	debateVotesService DebateVotesService
	jwt                gin.HandlerFunc
}

type DebateVotesService interface {
	Vote(ctx context.Context, userID int, debateID int, debateSideID int) (domain.DebateVote, error)
	ChangeVote(ctx context.Context, userID int, debateID int, debateSideID int) (domain.DebateVote, error)
	FinishDebate(
		ctx context.Context,
		userID int,
		debateID int,
	) error
}

type VoteRequest struct {
	DebateSideID int `json:"debate_side_id" validate:"required,min=1"`
}

func (h *DebateVotesHTTPHandler) Register(rg *gin.RouterGroup) {
	debates := rg.Group("/debates")
	debates.Use(h.jwt)
	debates.POST("/:id/vote", h.Vote)
	debates.PATCH("/:id/vote", h.ChangeVote)
	debates.POST("/:id/finish", h.FinishDebate)
}
