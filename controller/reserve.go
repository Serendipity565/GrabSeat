package controller

import (
	"GrabSeat/api/request"
	"GrabSeat/api/response"
	"GrabSeat/pkg/ijwt"
	"GrabSeat/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type ReserveController struct {
}

func NewReserveHandler() *ReserveController {
	return &ReserveController{}
}

// Reserve 预约座位接口
// @Summary 预约座位接口
// @Description 预约5天内座位接口
// @Tags reserve
// @Accept  json
// @Produce  json
func (r *ReserveController) Reserve(c *gin.Context, req request.ReserveReq, uc ijwt.UserClaims) (response.Response, error) {
	// 预约逻辑
	bData, err := service.BeforeDate(req.Data)
	if err != nil {
		return response.Response{
			Code: http.StatusBadRequest,
			Msg:  "data错误",
			Data: err.Error(),
		}, nil
	}
	diff := bData.Truncate(24 * time.Hour).Sub(time.Now().Truncate(24 * time.Hour))
	days := int(diff.Hours() / 24)
	if days > 5 {
		return response.Response{
			Code: http.StatusBadRequest,
			Msg:  "预约时间超过5天",
			Data: nil,
		}, nil
	}

	return response.Response{
		Code: http.StatusOK,
		Msg:  "已加入预约队列",
		Data: nil,
	}, nil

}
