package middleware

import (
	"GrabSeat/pkg/logger"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type LoggerMiddleware struct {
	log logger.Logger
}

func NewLoggerMiddleware(log logger.Logger) *LoggerMiddleware {
	return &LoggerMiddleware{
		log: log,
	}
}

// GinLogger 处理响应逻辑
func (lm *LoggerMiddleware) MiddlewareFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		path := ctx.Request.URL.Path
		ctx.Next() // 处理请求

		cost := time.Since(start)
		lm.log.Info("HTTP request",
			zap.String("method", ctx.Request.Method),
			zap.String("path", path),
			zap.Int("status", ctx.Writer.Status()),
			zap.String("client_ip", ctx.ClientIP()),
			zap.Duration("latency", cost),
		)
	}
}
