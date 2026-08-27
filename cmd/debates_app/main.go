package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	core_logger "github.com/qandoni/debatesApp/internal/core/logger"
	core_password "github.com/qandoni/debatesApp/internal/core/password"
	core_password_hash "github.com/qandoni/debatesApp/internal/core/password/hash"
	core_pgx_pool "github.com/qandoni/debatesApp/internal/core/repository/postgres/pool/pgx"
	core_http_middleware "github.com/qandoni/debatesApp/internal/core/transport/http/middleware"
	core_http_server "github.com/qandoni/debatesApp/internal/core/transport/http/server"
	auth_jwt "github.com/qandoni/debatesApp/internal/features/auth/jwt"
	auth_service "github.com/qandoni/debatesApp/internal/features/auth/service"
	auth_http_transport "github.com/qandoni/debatesApp/internal/features/auth/transport"
	users_repository "github.com/qandoni/debatesApp/internal/features/users/repository"
	users_service "github.com/qandoni/debatesApp/internal/features/users/service"
	users_http_transport "github.com/qandoni/debatesApp/internal/features/users/transport"
	"go.uber.org/zap"
)

func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM,
	)
	defer cancel()

	logger, err := core_logger.NewLogger(core_logger.NewConfigMust())
	if err != nil {
		fmt.Println("failed to init app logger:", err)
		os.Exit(1)
	}
	defer logger.Close()

	logger.Debug("initializing postgres connection pool")
	pool, err := core_pgx_pool.NewPool(
		ctx,
		core_pgx_pool.NewConfigMust(),
	)
	if err != nil {
		logger.Fatal("failed to init postgres connection pool", zap.Error(err))
	}
	defer pool.Close()

	logger.Debug("initializing transaction manager")
	txManager := core_pgx_pool.NewTransactionManager(pool)

	logger.Debug("initializing feature", zap.String("feature", "users"))
	usersRepository := users_repository.NewUsersRepository(pool, pool.OpTimeout())
	usersService := users_service.NewUsersService(usersRepository)
	usersHTTPTransport := users_http_transport.NewUsersHTTPHandler(usersService)

	logger.Debug("initializing feature", zap.String("feature", "auth"))
	passwordHasher := core_password.NewBcryptHasher()
	sha256Hasher := core_password_hash.NewSHA256Hasher()
	jwtManager := auth_jwt.NewJWTManager("my-secret-key")
	authService := auth_service.NewAuthService(usersRepository, passwordHasher, sha256Hasher, jwtManager, txManager)
	authTransportHTTP := auth_http_transport.NewAuthHTTPHandler(authService)

	logger.Debug("initializing HTTP server")
	server := core_http_server.NewHTTPServer(
		core_http_server.NewConfigMust(),
		logger,
	)

	server.Engine().Use(
		core_http_middleware.RequestID(),
		core_http_middleware.Logger(logger),
		core_http_middleware.Trace(),
		gin.Recovery(),
		core_http_middleware.ErrorHandler(),
	)
	core_http_server.RegisterRoutes(
		server.Engine(),
		authTransportHTTP,
		usersHTTPTransport,
		jwtManager,
	)
	if err := server.Run(ctx); err != nil {
		logger.Error(
			"http server stopped",
			zap.Error(err),
		)
	}
}
