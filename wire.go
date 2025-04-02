//go:build wireinject

package main

import (
	"GrabSeat/config"
	"GrabSeat/middleware"
	"GrabSeat/pkg/ijwt"
	"GrabSeat/web"
	"github.com/google/wire"
)

func InitApp() *App {
	wire.Build(
		wire.Struct(new(App), "*"),
		config.ProviderSet,
		//log.ProviderSet,
		ijwt.NewJWT,
		middleware.NewCorsMiddleware,
		middleware.NewAuthMiddleware,
		//service.ProviderSet,
		web.ProviderSet,
	)
	return &App{}
}
