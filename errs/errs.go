package errs

import (
	"GrabSeat/pkg/errorx"
	"net/http"
)

const (
	UserIdOrPasswordErrorCode = iota + 40001
	UnAuthoredErrorCode
)

const (
	InternalServerErrorCode = iota + 50001
	ErrorTypeErrorCode
)

var (
	UserIdOrPasswordError = func(err error) error {
		return errorx.New(http.StatusUnauthorized, UserIdOrPasswordErrorCode, "账号或者密码错误!", err)
	}
	UnAuthoredError = func(err error) error {
		return errorx.New(http.StatusUnauthorized, UnAuthoredErrorCode, "Authorization错误", err)
	}
)

var (
	InternalServerError = func(err error) error {
		return errorx.New(http.StatusInternalServerError, InternalServerErrorCode, "服务器内部错误", err)
	}
)
