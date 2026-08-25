package core_http_server

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	core_logger "github.com/qandoni/debatesApp/internal/core/logger"
	"go.uber.org/zap"
)

type HTTPServer struct {
	engine *gin.Engine
	config Config
	log    *core_logger.Logger
}

func NewHTTPServer(
	config Config,
	log *core_logger.Logger,
) *HTTPServer {
	engine := gin.New()

	return &HTTPServer{
		engine: engine,
		config: config,
		log:    log,
	}
}

func (s *HTTPServer) Engine() *gin.Engine {
	return s.engine
}

func (s *HTTPServer) Run(ctx context.Context) error {
	server := &http.Server{
		Addr:    s.config.Addr,
		Handler: s.engine,
	}

	ch := make(chan error, 1)

	go func() {
		s.log.Info("start HTTP server",
			zap.String("addr", s.config.Addr))
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			ch <- err
		}
		close(ch)
	}()

	select {
	case err := <-ch:
		return err
	case <-ctx.Done():
		s.log.Info("shutdown http server...")
		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			s.config.ShutdownTimeout,
		)
		defer cancel()

		return server.Shutdown(shutdownCtx)
	}
}
