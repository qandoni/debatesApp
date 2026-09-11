package core_http_server

import (
	"github.com/gin-gonic/gin"
)

type RouterRegistrar interface {
	Register(router *gin.RouterGroup)
}

func RegisterRoutes(
	engine *gin.Engine,
	registrars ...RouterRegistrar,
) {
	api := engine.Group("/api/v1")
	for _, registrar := range registrars {
		registrar.Register(api)
	}
}
