package statistics_http_transport

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/qandoni/debatesApp/internal/core/domain"
)

func NewStatisticsHTTPHandler(
	debateStatisticsService StatisticsService,
	jwt gin.HandlerFunc,
) *StatisticsHTTPHandler {
	return &StatisticsHTTPHandler{
		debateStatisticsService,
		jwt,
	}
}

type StatisticsHTTPHandler struct {
	debateStatisticsService StatisticsService
	jwt                     gin.HandlerFunc
}

type StatisticsService interface {
	GetStatistics(
		ctx context.Context,
		debateID int,
	) (domain.DebateStatistics, error)
}

func (h *StatisticsHTTPHandler) Register(rg *gin.RouterGroup) {
	debates := rg.Group("/debates")
	debates.Use(h.jwt)
	debates.GET("/:id/statistics", h.GetStatistics)
}
