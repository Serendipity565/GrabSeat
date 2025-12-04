package middleware

import (
	_ "embed"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Serendipity565/GrabSeat/api/response"
	"github.com/Serendipity565/GrabSeat/config"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

//go:embed scripts/limiter.lua
var limiterScriptSource string // lua 脚本内容将被加载到这个字符串变量中

type LimitMiddleware struct {
	capacity     int // 容量
	fillInterval int // 每秒补充令牌的次数
	quantum      int // 每次发放令牌数量
	client       redis.Cmdable
	script       *redis.Script
}

func NewLimitMiddleware(conf *config.LimiterConfig, client redis.Cmdable) *LimitMiddleware {
	return &LimitMiddleware{
		capacity:     conf.Capacity,
		fillInterval: conf.FillInterval,
		quantum:      conf.Quantum,
		client:       client,
		script:       redis.NewScript(limiterScriptSource),
	}
}

func (m *LimitMiddleware) Middleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		prefix := "ratelimit:" + strings.ReplaceAll(ctx.FullPath(), ":", "_")
		if prefix == "ratelimit:" {
			prefix = "ratelimit:" + strings.ReplaceAll(ctx.Request.URL.Path, ":", "_")
		}
		prefix = prefix + "_"

		availableKey := prefix + "tokens"
		latestKey := prefix + "ts"
		now := time.Now().UnixMilli()
		res, err := m.script.Run(
			ctx.Request.Context(),
			m.client,
			[]string{availableKey, latestKey},
			m.capacity,     // ARGV[1]
			m.quantum,      // ARGV[2]
			m.fillInterval, // ARGV[3] 每秒补充几次
			now,            // ARGV[4] 毫秒
			1,              // ARGV[5]
		).Int()
		if err != nil {
			ctx.Error(fmt.Errorf("限流器执行错误: %v", err))
			ctx.JSON(http.StatusInternalServerError, response.Response{
				Code: http.StatusInternalServerError,
				Msg:  "限流器内部错误",
				Data: nil,
			})
			return
		}
		if res == 0 {
			ctx.Error(errors.New("请求过于频繁，请稍后再试"))
			ctx.JSON(http.StatusTooManyRequests, response.Response{
				Code: http.StatusTooManyRequests,
				Msg:  "请求过于频繁，请稍后再试",
				Data: nil,
			})
			return
		}
		ctx.Next()
	}
}
