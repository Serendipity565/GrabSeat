package controller

import (
	"github.com/Serendipity565/GrabSeat/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

//	@title			CCNU 图书馆预约抢座 API
//	@version		1.0
//	@description	CCNU 图书馆预约抢座 API
//	@host			localhost:8080

var ProviderSet = wire.NewSet(
	NewLoginController,
	NewGarbHandler,
	NewGinEngine,
)

func NewGinEngine(
	lc *LoginController,
	gc *GarbController,

	corsMiddleware *middleware.CorsMiddleware,
	authMiddleware *middleware.AuthMiddleware,
	logMiddleware *middleware.LoggerMiddleware,
) *gin.Engine {
	gin.ForceConsoleColor()
	r := gin.Default()
	api := r.Group("/api/v1")
	// 跨域
	r.Use(corsMiddleware.MiddlewareFunc())
	r.Use(logMiddleware.MiddlewareFunc())

	lc.RegisterLoginRouter(api)
	gc.RegisterGarbRouter(api, authMiddleware.MiddlewareFunc())
	return r
}
