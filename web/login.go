package web

import (
	"GrabSeat/api/request"
	"GrabSeat/api/response"
	"GrabSeat/pkg/ginx"
	"github.com/gin-gonic/gin"
)

type LoginHandler interface {
	Login(c *gin.Context, req request.LoginRequest) (response.Response, error)
}

func RegisterLoginRouter(r *gin.Engine, lh LoginHandler) {
	c := r.Group("/ccnu")
	{
		c.POST("/login", ginx.WrapReq(lh.Login))
	}
}
