package main

import (
	"github.com/Serendipity565/GrabSeat/service"

	"github.com/gin-gonic/gin"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// @title			CCNU 图书馆预约抢座 API
// @version		1.0
// @description	CCNU 图书馆预约抢座 API
// @host			localhost:8080
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
	cfile := pflag.String("config", "config/config.yaml", "配置文件路径")
	pflag.Parse()

	viper.SetConfigType("yaml")
	viper.SetConfigFile(*cfile)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}
