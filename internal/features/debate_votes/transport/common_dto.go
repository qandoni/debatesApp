package debate_votes_http_transport

import (
	"time"

	"github.com/qandoni/debatesApp/internal/core/domain"
)

type DebateVoteResponse struct {
	ID           int        `json:"id"`
	Version      int        `json:"version"`
	DebateID     int        `json:"debate_id"`
	UserID       int        `json:"user_id"`
	DebateSideID int        `json:"debate_side_id"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`
	IsChanged    bool       `json:"is_changed"`
}

func debateVoteResponseFromDomain(
	vote domain.DebateVote,
) DebateVoteResponse {
	return DebateVoteResponse{
		ID:           vote.ID,
		Version:      vote.Version,
		DebateID:     vote.DebateID,
		UserID:       vote.UserID,
		DebateSideID: vote.DebateSideID,
		CreatedAt:    vote.CreatedAt,
		UpdatedAt:    vote.UpdatedAt,
		IsChanged:    vote.IsChanged,
	}
}
