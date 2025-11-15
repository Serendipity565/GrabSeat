package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Serendipity565/GrabSeat/api/request"
	"github.com/Serendipity565/GrabSeat/api/response"
	"github.com/Serendipity565/GrabSeat/pkg/ginx"
	"github.com/Serendipity565/GrabSeat/pkg/ijwt"
	"github.com/Serendipity565/GrabSeat/service/garb"

	"github.com/gin-gonic/gin"
)

var Areas = []string{"101699191", "101699189", "101699187", "101699179"}

type GarbController struct {
}

func NewGarbHandler() *GarbController {
	return &GarbController{}
}

func (gc *GarbController) RegisterGarbRouter(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	c := r.Group("/garb")
	{
		c.GET("/test", authMiddleware, ginx.WrapClaims(gc.Test))
		c.POST("/findvacantseats", authMiddleware, ginx.WrapClaimsAndReq(gc.FindVacantSeats))
		c.POST("/seatttoname", authMiddleware, ginx.WrapClaimsAndReq(gc.SeatToName))
		c.POST("/isinlibrary", authMiddleware, ginx.WrapClaimsAndReq(gc.IsInLibrary))
		c.POST("/mgarb", authMiddleware, ginx.WrapClaimsAndReq(gc.MGarb))
	}
}

// Test 测试接口
// @Summary 测试接口
// @Description 测试接口
// @Tags garb
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer {{JWT}}"
// @Success 200 {string} string "test success"
// @Router /api/v1/garb/test [get]
func (gc *GarbController) Test(c *gin.Context, uc ijwt.UserClaims) (response.Response, error) {
	return response.Response{
		Code: 200,
		Msg:  "test success",
		Data: uc,
	}, nil
}

// FindVacantSeats 查找空座位接口，可指定条件或模糊查找
// @Summary 查找空座位接口
// @Description 查找空座位接口
// @Tags garb
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer {{JWT}}"
// @Param request body request.MFindVacantSeatsReq true "查找空座位请求参数"
// @Success 200 {object} response.Response{data=response.MFindVacantSeatsResp} "成功返回空座位列表"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/garb/findvacantseats [post]
func (gc *GarbController) FindVacantSeats(c *gin.Context, req request.MFindVacantSeatsReq, uc ijwt.UserClaims) (response.Response, error) {
	if req.StartTime >= req.EndTime {
		return response.Response{
			Code: http.StatusBadRequest,
			Msg:  "请求参数错误",
			Data: "开始时间必须小于结束时间",
		}, nil
	}
	grabber := garb.NewGrabber(Areas, req.IsTomorrow, req.StartTime, req.EndTime)
	grabber.StartFlushClient(uc.UserId, uc.Password, time.Second*10)
	mDevId := grabber.MFindVacantSeats(req.KeyWord)
	return response.Response{
		Code: http.StatusOK,
		Msg:  "success",
		Data: response.MFindVacantSeatsResp{
			Seats: mDevId,
		},
	}, nil
}

// SeatToName 座位号转名字接口
// @Summary 座位号转名字接口
// @Description 座位号转名字接口
// @Tags garb
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer {{JWT}}"
// @Param request body request.SeatToNameReq true "座位号转名字请求参数"
// @Success 200 {object} response.Response{data=response.SeatToNameResp} "成功返回座位号对应的名字"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/garb/seatttoname [post]
func (gc *GarbController) SeatToName(c *gin.Context, req request.SeatToNameReq, uc ijwt.UserClaims) (response.Response, error) {
	grabber := garb.NewGrabber(Areas, false, "08:00", "22:00") //这里的是时间设置不重要
	grabber.StartFlushClient(uc.UserId, uc.Password, time.Second*10)
	ts := grabber.SeatToName(req.SeatId)
	return response.Response{
		Code: http.StatusOK,
		Msg:  "success",
		Data: response.SeatToNameResp{
			Ts: ts,
		},
	}, nil
}

// IsInLibrary 检查目标用户当前是否在图书馆
// @Summary 检查目标用户当前是否在图书馆
// @Description 检查目标用户当前是否在图书馆
// @Tags garb
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer {{JWT}}"
// @Param request body request.IsInLibraryReq true "检查目标用户当前是否在图书馆请求参数"
// @Success 200 {object} response.Response{data=string} "成功返回在图书馆的时间段"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/garb/isinlibrary [post]
func (gc *GarbController) IsInLibrary(c *gin.Context, req request.IsInLibraryReq, uc ijwt.UserClaims) (response.Response, error) {
	grabber := garb.NewGrabber(Areas, false, "08:00", "22:00") //这里的是时间设置不重要
	grabber.StartFlushClient(uc.UserId, uc.Password, time.Second*10)
	ot := grabber.IsInLibrary(req.Username)
	if ot != nil {
		return response.Response{
			Code: http.StatusOK,
			Msg:  "success",
			Data: fmt.Sprintf("在图书馆的%s，%s - %s\n", ot.Title, ot.Start, ot.End),
		}, nil
	} else {
		return response.Response{
			Code: http.StatusOK,
			Msg:  "success",
			Data: "不在图书馆",
		}, nil
	}
}

// MGarb 抢座接口
// @Summary 抢座接口
// @Description 抢座接口
// @Tags garb
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer {{JWT}}"
// @Param request body request.MGarbReq true "抢座请求参数"
// @Success 200 {object} response.Response{data=string} "成功返回抢座结果"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/garb/mgarb [post]
func (gc *GarbController) MGarb(c *gin.Context, req request.MGarbReq, uc ijwt.UserClaims) (response.Response, error) {
	if req.StartTime >= req.EndTime {
		return response.Response{
			Code: http.StatusBadRequest,
			Msg:  "请求参数错误",
			Data: "开始时间必须小于结束时间",
		}, nil
	}
	grabber := garb.NewGrabber(Areas, req.IsTomorrow, req.StartTime, req.EndTime)
	grabber.StartFlushClient(uc.UserId, uc.Password, time.Second*10)
	devId := grabber.MFindVacantSeats(req.KeyWord)
	if len(devId) == 0 {
		return response.Response{
			Code: http.StatusOK,
			Msg:  "fail",
			Data: "没有空座位",
		}, nil
	}
	select {
	case <-time.After(10 * time.Second):
		return response.Response{
			Code: http.StatusOK,
			Msg:  "fail",
			Data: "抢座超时，未能成功抢到座位",
		}, nil
	default:
		for _, val := range devId {
			grabber.Grab(val.DevId)
			// 二次成功验证
			if grabber.GrabSuccess() {
				return response.Response{
					Code: http.StatusOK,
					Msg:  "success",
					Data: fmt.Sprintf("抢座成功，座位号：%s", val.Title),
				}, nil
			}
		}
	}
	return response.Response{
		Code: http.StatusOK,
		Msg:  "fail",
		Data: "抢座超时，未能成功抢到座位",
	}, nil
}
