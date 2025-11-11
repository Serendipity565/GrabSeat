//go:build wireinject

package main

import (
	"GrabSeat/config"
	"GrabSeat/middleware"
	"GrabSeat/pkg/ijwt"
	"GrabSeat/pkg/logger"
	"GrabSeat/service"
	"GrabSeat/web"

	"github.com/google/wire"
)

func InitApp() *App {
	wire.Build(
		wire.Struct(new(App), "*"),
		config.ProviderSet,
		logger.ProviderSet,
		ijwt.NewJWT,
		middleware.NewCorsMiddleware,
		middleware.NewAuthMiddleware,
		middleware.NewLoggerMiddleware,
		service.ProviderSet,
		web.ProviderSet,
	)
	return &App{}
}
