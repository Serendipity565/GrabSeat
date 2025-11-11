package web

import (
	"GrabSeat/controller"
	"GrabSeat/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

// @title CCNU 图书馆预约抢座 API
// @version 1.0
// @description CCNU 图书馆预约抢座 API
// @host localhost:8080

var ProviderSet = wire.NewSet(
	NewGinEngine,
	controller.NewLoginController,
	wire.Bind(new(LoginHandler), new(*controller.LoginController)),
	controller.NewGarbHandler,
	wire.Bind(new(GarbHandler), new(*controller.GarbController)),
)

func NewGinEngine(
	gh GarbHandler,
	lh LoginHandler,
	corsMiddleware *middleware.CorsMiddleware,
	authMiddleware *middleware.AuthMiddleware,
	logMiddleware *middleware.LoggerMiddleware,
) *gin.Engine {
	gin.ForceConsoleColor()
	r := gin.Default()
	// 跨域
	r.Use(corsMiddleware.MiddlewareFunc())
	r.Use(logMiddleware.MiddlewareFunc())

	RegisterLoginRouter(r, lh)
	RegisterGarbRouter(r, gh, authMiddleware.MiddlewareFunc())
	return r
}
