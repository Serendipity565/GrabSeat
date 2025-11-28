package middleware

import "github.com/gin-gonic/gin"

type LimitMiddleware struct {
	// TODO redis限流
}

func NewLimitMiddleware() *LimitMiddleware {
	return &LimitMiddleware{}
}

func (m *LimitMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO 实现限流逻辑
	}
}
