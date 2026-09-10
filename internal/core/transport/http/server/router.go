package core_http_server

import (
	"github.com/gin-gonic/gin"
	core_http_middleware "github.com/qandoni/debatesApp/internal/core/transport/http/middleware"
	comments_transport_http "github.com/qandoni/debatesApp/internal/features/comments/transport/http"
)

type RouterRegistrar interface {
	Register(router *gin.RouterGroup)
}

func RegisterRoutes(
	engine *gin.Engine,
	authHandler RouterRegistrar,
	usersHandler RouterRegistrar,
	postsHandler RouterRegistrar,
	postImagesHandler RouterRegistrar,
	debateVotesHandler RouterRegistrar,
	commentsHandler *comments_transport_http.CommentsHTTPHandler,
	commentRatingsHandler RouterRegistrar,
	statisticsHandler RouterRegistrar,
	parser core_http_middleware.TokenParser,
) {
	jwt := core_http_middleware.JWT(parser)
	api := engine.Group("/api/v1")
	auth := api.Group("/auth")
	authHandler.Register(auth)

	users := api.Group("/users")
	users.Use(jwt)
	usersHandler.Register(users)

	posts := api.Group("/posts")
	posts.Use(jwt)
	postsHandler.Register(posts)
	postImagesHandler.Register(posts)
	commentsHandler.Register(posts)
	debates := api.Group("/debates")
	debates.Use(jwt)
	debateVotesHandler.Register(debates)

	comments := api.Group("/comments")
	comments.Use(jwt)
	commentRatingsHandler.Register(comments)

	statisticsHandler.Register(debates)

}
