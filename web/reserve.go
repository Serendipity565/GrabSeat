package web

import (
	"GrabSeat/api/request"
	"GrabSeat/api/response"
	"GrabSeat/pkg/ginx"
	"GrabSeat/pkg/ijwt"
	"github.com/gin-gonic/gin"
)

type ReserveHandler interface {
	Reserve(c *gin.Context, req request.ReserveReq, uc ijwt.UserClaims) (response.Response, error)
}

func RegisterReserveRouter(r *gin.Engine, rc ReserveHandler, authMiddleware gin.HandlerFunc) {
	c := r.Group("/reserve")
	{
		c.POST("/reserve", authMiddleware, ginx.WrapClaimsAndReq(rc.Reserve))
	}
}
