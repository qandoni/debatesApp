package core_http_server //TODO либо тут хранить руты к методам, либо сделать как то организованнее и понятнее

import (
	"github.com/gin-gonic/gin"
	core_http_middleware "github.com/qandoni/debatesApp/internal/core/transport/http/middleware"
	auth_http_transport "github.com/qandoni/debatesApp/internal/features/auth/transport"
	users_http_transport "github.com/qandoni/debatesApp/internal/features/users/transport"
)

func RegisterRoutes(
	engine *gin.Engine,
	authHandler *auth_http_transport.AuthHTTPHandler,
	usersHandler *users_http_transport.UsersHTTPHandler,

	parser core_http_middleware.TokenParser,
) {
	jwt := core_http_middleware.JWT(parser)
	api := engine.Group("/api/v1")
	auth := api.Group("/auth")
	authHandler.Register(auth)

	users := api.Group("/users")
	users.Use(jwt)
	usersHandler.Register(users)
}
