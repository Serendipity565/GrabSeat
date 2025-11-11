package main

import (
	"GrabSeat/service"

	"github.com/gin-gonic/gin"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	initViper()
	app := InitApp()
	//app.t.StartDailyTask()
	app.r.Run(":8080")
}

type App struct {
	// Logger *log.Logger
	r *gin.Engine
	t *service.Ticker
}

func initViper() {
	cfile := pflag.String("config", "config/config-example.yaml", "配置文件路径")
	pflag.Parse()

	viper.SetConfigType("yaml")
	viper.SetConfigFile(*cfile)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}
