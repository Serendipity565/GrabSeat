//go:build wireinject

package main

import (
	"github.com/Serendipity565/GrabSeat/config"
	"github.com/Serendipity565/GrabSeat/controller"
	"github.com/Serendipity565/GrabSeat/ioc"
	"github.com/Serendipity565/GrabSeat/middleware"
	"github.com/Serendipity565/GrabSeat/pkg/ijwt"
	"github.com/Serendipity565/GrabSeat/service"
	"github.com/google/wire"
)

func InitApp() *App {
	wire.Build(
		wire.Struct(new(App), "*"),
		config.ProviderSet,
		ioc.ProviderSet,
		ijwt.NewJWT,
		middleware.NewCorsMiddleware,
		middleware.NewAuthMiddleware,
		middleware.NewLoggerMiddleware,
		middleware.NewLimitMiddleware,
		service.ProviderSet,
		controller.ProviderSet,
	)
	return &App{}
}
