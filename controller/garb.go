package controller

import (
	"fmt"
	"net/http"

	"github.com/Serendipity565/GrabSeat/api/request"
	"github.com/Serendipity565/GrabSeat/api/response"
	"github.com/Serendipity565/GrabSeat/pkg/ginx"
	"github.com/Serendipity565/GrabSeat/pkg/ijwt"
	"github.com/Serendipity565/GrabSeat/service"
	"github.com/gin-gonic/gin"
)

type GarbController struct {
	gs service.GrabberService
}

func NewGarbHandler(gs service.GrabberService) *GarbController {
	return &GarbController{
		gs: gs,
	}
}

func (gc *GarbController) RegisterGarbRouter(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	c := r.Group("/garb")
	{
		c.POST("/findvacantseats", authMiddleware, ginx.WrapClaimsAndReq(gc.FindVacantSeats))
		c.POST("/seatttoname", authMiddleware, ginx.WrapClaimsAndReq(gc.SeatToName))
		c.POST("/isinlibrary", authMiddleware, ginx.WrapClaimsAndReq(gc.IsInLibrary))
		c.POST("/garb", authMiddleware, ginx.WrapClaimsAndReq(gc.Garb))
	}
}

// FindVacantSeats 查找空座位接口，可指定条件或模糊查找
//
//	@Summary		查找空座位接口
//	@Description	查找空座位接口
//	@Tags			garb
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string									true	"Bearer {{JWT}}"
//	@Param			request			body		request.FindVacantSeatsReq				true	"查找空座位请求参数"
//	@Success		200				{object}	response.Response{data=[]response.Seat}	"成功返回空座位列表"
//	@Failure		400				{object}	response.Response						"请求参数错误"
//	@Failure		500				{object}	response.Response						"服务器内部错误"
//	@Router			/api/v1/garb/findvacantseats [post]
func (gc *GarbController) FindVacantSeats(c *gin.Context, req request.FindVacantSeatsReq, uc ijwt.UserClaims) (response.Response, error) {
	if req.StartTime >= req.EndTime {
		return response.Response{
			Code: http.StatusBadRequest,
			Msg:  "请求参数错误",
			Data: "开始时间必须小于结束时间",
		}, nil
	}
	client, err := gc.gs.GetClient(uc.UserId, uc.Password)
	if err != nil {
		return response.Response{}, err
	}
	seats, err := gc.gs.FindVacantSeats(client, req.StartTime, req.EndTime, req.KeyWord, *req.IsTomorrow)
	if err != nil {
		return response.Response{}, err
	}
	return response.Response{
		Code: 0,
		Msg:  "Success",
		Data: seats,
	}, nil
}

// SeatToName 座位号转名字接口
//
//	@Summary		座位号转名字接口
//	@Description	座位号转名字接口
//	@Tags			garb
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string									true	"Bearer {{JWT}}"
//	@Param			request			body		request.SeatToNameReq					true	"座位号转名字请求参数"
//	@Success		200				{object}	response.Response{data=[]response.Ts}	"成功返回座位号对应的名字"
//	@Failure		400				{object}	response.Response						"请求参数错误"
//	@Failure		500				{object}	response.Response						"服务器内部错误"
//	@Router			/api/v1/garb/seatttoname [post]
func (gc *GarbController) SeatToName(c *gin.Context, req request.SeatToNameReq, uc ijwt.UserClaims) (response.Response, error) {
	client, err := gc.gs.GetClient(uc.UserId, uc.Password)
	if err != nil {
		return response.Response{}, err
	}
	ts, err := gc.gs.SeatToName(client, req.SeatName, *req.IsTomorrow)
	return response.Response{
		Code: 0,
		Msg:  "Success",
		Data: ts,
	}, nil
}

// IsInLibrary 检查目标用户当前是否在图书馆
//
//	@Summary		检查目标用户当前是否在图书馆
//	@Description	检查目标用户当前是否在图书馆
//	@Tags			garb
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string							true	"Bearer {{JWT}}"
//	@Param			request			body		request.IsInLibraryReq			true	"检查目标用户当前是否在图书馆请求参数"
//	@Success		200				{object}	response.Response{data=string}	"成功返回在图书馆的时间段"
//	@Failure		400				{object}	response.Response				"请求参数错误"
//	@Failure		500				{object}	response.Response				"服务器内部错误"
//	@Router			/api/v1/garb/isinlibrary [post]
func (gc *GarbController) IsInLibrary(c *gin.Context, req request.IsInLibraryReq, uc ijwt.UserClaims) (response.Response, error) {
	client, err := gc.gs.GetClient(uc.UserId, uc.Password)
	if err != nil {
		return response.Response{}, err
	}
	ot, _ := gc.gs.IsInLibrary(client, req.StudentName)
	if ot != nil {
		return response.Response{
			Code: 0,
			Msg:  "Success",
			Data: fmt.Sprintf("在图书馆的%s，%s - %s\n", ot.Title, ot.Start, ot.End),
		}, nil
	} else {
		return response.Response{
			Code: 0,
			Msg:  "Success",
			Data: "不在图书馆",
		}, nil
	}
}

// Garb 抢座接口
//
//	@Summary		抢座接口
//	@Description	抢座接口
//	@Tags			garb
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string							true	"Bearer {{JWT}}"
//	@Param			request			body		request.GarbReq					true	"抢座请求参数"
//	@Success		200				{object}	response.Response{data=string}	"成功返回抢座结果"
//	@Failure		400				{object}	response.Response				"请求参数错误"
//	@Failure		500				{object}	response.Response				"服务器内部错误"
//	@Router			/api/v1/garb/garb [post]
func (gc *GarbController) Garb(c *gin.Context, req request.GarbReq, uc ijwt.UserClaims) (response.Response, error) {
	if req.StartTime >= req.EndTime {
		return response.Response{
			Code: http.StatusBadRequest,
			Msg:  "请求参数错误",
			Data: "开始时间必须小于结束时间",
		}, nil
	}
	client, err := gc.gs.GetClient(uc.UserId, uc.Password)
	if err != nil {
		return response.Response{}, err
	}
	success, err := gc.gs.Grab(client, req.SeatID, req.StartTime, req.EndTime, *req.IsTomorrow)
	if err != nil {
		return response.Response{}, err
	}
	if !success {
		return response.Response{
			Code: http.StatusOK,
			Msg:  "fail",
			Data: "抢座超时，未能成功抢到座位",
		}, nil
	}
	return response.Response{
		Code: http.StatusOK,
		Msg:  "success",
		Data: "抢座成功",
	}, nil
}
