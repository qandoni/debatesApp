package statistics_http_transport

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/qandoni/debatesApp/internal/core/domain"
)

func NewStatisticsHTTPHandler(
	debateStatisticsService StatisticsService,
) *StatisticsHTTPHandler {
	return &StatisticsHTTPHandler{
		debateStatisticsService,
	}
}

type StatisticsHTTPHandler struct {
	debateStatisticsService StatisticsService
}

type StatisticsService interface {
	GetStatistics(
		ctx context.Context,
		debateID int,
	) (domain.DebateStatistics, error)
}

func (h *StatisticsHTTPHandler) Register(rg *gin.RouterGroup) {
	rg.GET("/:id/statistics", h.GetStatistics)
}
