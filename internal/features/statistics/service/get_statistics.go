package statistics_service

import (
	"context"
	"fmt"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
)

func (s *StatisticsService) GetStatistics(
	ctx context.Context,
	debateID int,
) (domain.DebateStatistics, error) {

	debate, err := s.debatesRepository.GetByID(
		ctx,
		debateID,
	)
	if err != nil {
		return domain.DebateStatistics{}, fmt.Errorf(
			"get debate: %w",
			err,
		)
	}

	if debate.Status != "FINISHED" {
		return domain.DebateStatistics{}, fmt.Errorf(
			"statistics are available only after debate is finished: %w",
			core_errors.ErrAccessForbidden,
		)
	}

	statistics, err := s.statisticsRepository.GetStatistics(
		ctx,
		debateID,
	)
	if err != nil {
		return domain.DebateStatistics{}, fmt.Errorf(
			"get debate statistics: %w",
			err,
		)
	}

	return statistics, nil
}
