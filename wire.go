//go:build wireinject

package main

import (
	"GrabSeat/config"
	"GrabSeat/controller"
	"GrabSeat/middleware"
	"GrabSeat/pkg/ijwt"
	"GrabSeat/pkg/logger"
	"GrabSeat/service"

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
		controller.ProviderSet,
	)
	return &App{}
}
