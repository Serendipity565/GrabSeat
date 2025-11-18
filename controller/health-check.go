package controller

import (
	"github.com/Serendipity565/GrabSeat/api/response"
	"github.com/Serendipity565/GrabSeat/service"
	"github.com/gin-gonic/gin"
)

type HealthCheckController struct {
}

func NewHealthCheckController() *HealthCheckController {
	return &HealthCheckController{}
}

// HealthCheck 健康检查
// @Summary		健康检查，返回当前服务占用的资源等信息
// @Description	健康检查，返回当前服务占用的资源等信息
// @Tags			health
// @Accept			json
// @Produce		json
// @Success		200	{object}	response.Response{data=response.HealthCheckResponse}	"成功返回健康检查结果"
// @Failure		400	{object}	response.Response										"请求参数错误"
// @Failure		500	{object}	response.Response										"服务器内部错误"
// @Router			/api/v1/health [get]
func HealthCheck(c *gin.Context) (response.Response, error) {
	resp := service.HealthCheckService()

	return response.Response{
		Code: 0,
		Msg:  "Success",
		Data: resp,
	}, nil
}
