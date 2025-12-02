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
	NewHealthCheckController,
	NewGinEngine,
)

func NewGinEngine(
	hc *HealthCheckController,
	lc *LoginController,
	gc *GarbController,

	corsMiddleware *middleware.CorsMiddleware,
	authMiddleware *middleware.AuthMiddleware,
	logMiddleware *middleware.LoggerMiddleware,
	limitMiddleware *middleware.LimitMiddleware,
) *gin.Engine {
	gin.ForceConsoleColor()
	r := gin.Default()

	r.Use(corsMiddleware.MiddlewareFunc())
	r.Use(logMiddleware.MiddlewareFunc())
	r.Use(limitMiddleware.Middleware())

	api := r.Group("/api/v1")

	hc.RegisterHealthCheckRouter(api)
	lc.RegisterLoginRouter(api)
	gc.RegisterGarbRouter(api, authMiddleware.MiddlewareFunc())

	return r
}
