package web

import (
	"GrabSeat/api/request"
	"GrabSeat/api/response"
	"GrabSeat/pkg/ginx"
	"GrabSeat/pkg/ijwt"
	"github.com/gin-gonic/gin"
)

type GarbHandler interface {
	Test(c *gin.Context, uc ijwt.UserClaims) (response.Response, error)
	FindVacantSeats(c *gin.Context, req request.MFindVacantSeatsReq, uc ijwt.UserClaims) (response.Response, error)
	SeatToName(c *gin.Context, req request.SeatToNameReq, uc ijwt.UserClaims) (response.Response, error)
	IsInLibrary(c *gin.Context, req request.IsInLibraryReq, uc ijwt.UserClaims) (response.Response, error)
	MGarb(c *gin.Context, req request.MGarbReq, uc ijwt.UserClaims) (response.Response, error)
}

func RegisterGarbRouter(r *gin.Engine, gc GarbHandler, authMiddleware gin.HandlerFunc) {
	c := r.Group("/garb")
	{
		c.GET("/test", authMiddleware, ginx.WrapClaims(gc.Test))
		c.POST("/findvacantseats", authMiddleware, ginx.WrapClaimsAndReq(gc.FindVacantSeats))
		c.POST("/seatttoname", authMiddleware, ginx.WrapClaimsAndReq(gc.SeatToName))
		c.POST("/isinlibrary", authMiddleware, ginx.WrapClaimsAndReq(gc.IsInLibrary))
		c.POST("/mgarb", authMiddleware, ginx.WrapClaimsAndReq(gc.MGarb))
	}
}
