package controller

import (
	"GrabSeat/api/request"
	"GrabSeat/api/response"
	"GrabSeat/pkg/ijwt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type LoginController struct {
	jwtHandler *ijwt.JWT
}

func NewLoginController(jwtHandler *ijwt.JWT) *LoginController {
	return &LoginController{jwtHandler: jwtHandler}
}

// Login 登录接口
// @Summary 用户登录
// @Description 用户登录，返回 JWT 令牌
// @Tags auth
// @Accept  json
// @Produce  json
// @Param request body request.LoginRequest true "登录请求参数"
// @Success 200 {object} response.Response{data=map[string]string} "成功返回 JWT 令牌"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /ccnu/login [post]
func (lc *LoginController) Login(c *gin.Context, req request.LoginRequest) (response.Response, error) {
	// 验证用户名和密码（这里假设验证通过）
	// 生成JWT令牌
	token, err := lc.jwtHandler.SetJWTToken(req.Username, req.Password)
	if err != nil {
		return response.Response{
			Code: http.StatusInternalServerError,
			Msg:  "生成令牌失败",
			Data: nil,
		}, err
	}

	return response.Response{
		Code: http.StatusOK,
		Msg:  "登录成功",
		Data: map[string]string{
			"token": "Bearer " + token,
		},
	}, nil
}
