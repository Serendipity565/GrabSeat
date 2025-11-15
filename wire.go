//go:build wireinject

package main

import (
	"github.com/Serendipity565/GrabSeat/config"
	"github.com/Serendipity565/GrabSeat/controller"
	"github.com/Serendipity565/GrabSeat/middleware"
	"github.com/Serendipity565/GrabSeat/pkg/ijwt"
	"github.com/Serendipity565/GrabSeat/pkg/logger"
	"github.com/Serendipity565/GrabSeat/service"

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
