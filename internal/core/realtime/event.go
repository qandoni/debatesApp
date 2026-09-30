package core_realtime

import "time"

func NewEvent(
	eventType string,
	data any,
) Event {
	return Event{
		eventType,
		data,
	}
}

type Event struct {
	Type string
	Data any
}

type Envelope struct {
	Type       string    `json:"type"`
	Topic      string    `json:"topic"`
	Data       any       `json:"data,omitempty"`
	OccurredAt time.Time `json:"occurred_at"`
}

const (
	EventCommentCreated        = "comment.created"
	EventArgumentCreated       = "argument.created"
	EventCommentUpdated        = "comment.updated"
	EventArgumentAuthorLiked   = "argument.author_like.updated"
	EventArgumentRatingUpdated = "argument.rating.updated"
	EventDebateVoteCreated     = "debate.vote.created"
	EventDebateVoteChanged     = "debate.vote.changed"
	EventDebateFinished        = "debate.finished"
)
