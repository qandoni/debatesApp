package statistics_http_transport

import (
	"net/http"

	"github.com/gin-gonic/gin"
	core_http_request "github.com/qandoni/debatesApp/internal/core/transport/http/request"
)

func (h *StatisticsHTTPHandler) GetStatistics(c *gin.Context) {
	debateID, err := core_http_request.GetIntPathValue(c, "id")
	if err != nil {
		c.Error(err).SetMeta("failed to get debate id")
		return
	}

	statistics, err := h.debateStatisticsService.GetStatistics(
		c.Request.Context(),
		debateID,
	)
	if err != nil {
		c.Error(err).SetMeta("failed to get debate statistics")
		return
	}

	response := NewDebateStatisticsDTOFromDomain(statistics)

	c.JSON(http.StatusOK, response)
}
