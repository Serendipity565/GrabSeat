package controller

import (
	"github.com/Serendipity565/GrabSeat/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	basicAuthMiddleware *middleware.BasicAuthMiddleware,
	logMiddleware *middleware.LoggerMiddleware,
	limitMiddleware *middleware.LimitMiddleware,
	prometheusMiddleware *middleware.PrometheusMiddleware,
) *gin.Engine {
	gin.ForceConsoleColor()
	r := gin.Default()

	r.Use(corsMiddleware.MiddlewareFunc())
	r.Use(logMiddleware.MiddlewareFunc())
	r.Use(prometheusMiddleware.MiddlewareFunc())
	r.Use(limitMiddleware.Middleware())

	// Prometheus metrics endpoint with basic auth
	reg := prometheusMiddleware.GetRegistry()
	r.GET("/metrics", basicAuthMiddleware.MiddlewareFunc(), gin.WrapH(promhttp.HandlerFor(
		reg,
		promhttp.HandlerOpts{},
	)))

	api := r.Group("/api/v1")

	hc.RegisterHealthCheckRouter(api)
	lc.RegisterLoginRouter(api)
	gc.RegisterGarbRouter(api, authMiddleware.MiddlewareFunc())

	return r
}
