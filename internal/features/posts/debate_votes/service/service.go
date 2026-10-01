package debate_votes_service

import (
	"context"
	"time"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_realtime "github.com/qandoni/debatesApp/internal/core/realtime"
)

func NewDebateVotesService(
	debateVotesRepository DebateVotesRepository,
	debatesRepository DebatesRepository,
	debateSidesRepository DebateSidesRepository,
	commentsRepository CommentsRepository,
	realtimeHub core_realtime.Publisher,
) *DebateVotesService {
	return &DebateVotesService{
		debateVotesRepository,
		debatesRepository,
		debateSidesRepository,
		commentsRepository,
		realtimeHub,
	}
}

type DebateVotesService struct {
	debateVotesRepository DebateVotesRepository
	debatesRepository     DebatesRepository
	debateSidesRepository DebateSidesRepository
	commentsRepository    CommentsRepository
	realtimeHub           core_realtime.Publisher
}

// publishEvent отправляет событие всем подписчикам топика поста.
// Вызывается только после успешной записи в репозитории.
func (s *DebateVotesService) publishEvent(postID int, eventType string, data any) {
	s.realtimeHub.Publish(
		core_realtime.PostTopic(postID),
		core_realtime.NewEvent(eventType, data),
	)
}

type CommentsRepository interface {
	HasUserArgumentInPost(
		ctx context.Context,
		userID int,
		postID int,
	) (bool, error)
}

type DebatesRepository interface {
	GetByID(ctx context.Context, debateID int) (domain.Debate, error)
	FinishDebate(
		ctx context.Context,
		debateID int,
	) error
	GetAuthorID(
		ctx context.Context,
		debateID int,
	) (int, error)
}

type DebateSidesRepository interface {
	GetByDebateID(ctx context.Context, debateID int) ([]domain.DebateSide, error)
}

type DebateVotesRepository interface {
	Create(ctx context.Context, vote domain.DebateVote) (domain.DebateVote, error)

	GetByDebateAndUser(
		ctx context.Context,
		debateID int,
		userID int,
	) (domain.DebateVote, error)

	Update(
		ctx context.Context,
		debateID int,
		userID int,
		debateSideID int,
		updatedAt time.Time,
	) (domain.DebateVote, error)
}
