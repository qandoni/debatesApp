package statistics_service

import (
	"context"

	"github.com/qandoni/debatesApp/internal/core/domain"
)

func NewStatisticsService(
	statisticsRepository StatisticsRepository,
	debatesRepository DebatesRepository,
) *StatisticsService {
	return &StatisticsService{
		statisticsRepository,
		debatesRepository,
	}
}

type StatisticsService struct {
	statisticsRepository StatisticsRepository
	debatesRepository    DebatesRepository
}

type StatisticsRepository interface {
	GetStatistics(
		ctx context.Context,
		debateID int,
	) (domain.DebateStatistics, error)
}

type DebatesRepository interface {
	CreateDebate(
		ctx context.Context,
		debate domain.Debate,
	) (domain.Debate, error)
	GetAuthorID(
		ctx context.Context,
		debateID int,
	) (int, error)
	GetByID(
		ctx context.Context,
		debateID int,
	) (domain.Debate, error)
	GetByPostID(
		ctx context.Context,
		postID int,
	) (domain.Debate, error)
	FinishDebate(
		ctx context.Context,
		debateID int,
	) error
}
