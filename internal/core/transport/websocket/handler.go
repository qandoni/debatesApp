package core_transport_websocket

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	core_logger "github.com/qandoni/debatesApp/internal/core/logger"
	core_realtime "github.com/qandoni/debatesApp/internal/core/realtime"
	"go.uber.org/zap"
)

type Handler struct {
	hub      *core_realtime.Hub
	parser   TokenParser
	config   Config
	upgrader websocket.Upgrader
}

func NewHandler(hub *core_realtime.Hub, parser TokenParser, config Config) *Handler {
	return &Handler{
		hub:    hub,
		parser: parser,
		config: config,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  int(config.ReadLimit),
			WriteBufferSize: int(config.ReadLimit),
			CheckOrigin:     originChecker(config.AllowedOrigins),
		},
	}
}

func originChecker(allowedOrigins []string) func(request *http.Request) bool {
	return func(request *http.Request) bool {
		origin := request.Header.Get("Origin")
		if origin == "" {
			return true
		}
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				return true
			}
		}
		return false
	}
}

func (h *Handler) Register(router *gin.RouterGroup) {
	router.GET("/ws", h.Handle)
}

func (h *Handler) Handle(c *gin.Context) {
	log := core_logger.FromContext(c.Request.Context()) // ставится middleware.Logger

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		// gorilla уже отправила HTTP-ответ (обычно 400) — c.Error/AbortWithStatus здесь
		// нельзя: WriteHeader уже вызван. Только лог + Abort.
		log.Warn("failed to upgrade to websocket", zap.Error(err))
		c.Abort()
		return
	}

	client := NewClient(conn, h.hub, h.parser, h.config, log)
	client.Run() // блокируется на всё время жизни соединения
}
