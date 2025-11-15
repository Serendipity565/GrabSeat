package service

import (
	"errors"
	"net/http"
	"testing"

	mockservice "github.com/Serendipity565/GrabSeat/service/mocks"
	"github.com/golang/mock/gomock"
)

var ErrLoginFailed = errors.New("登录失败，用户名或密码错误")

// 辅助函数：调用 LoginService，便于测试
func performLogin(ls LoginService, username, password string) error {
	_, err := ls.Login2CAS(username, password)
	return err
}

func TestLoginController_Login2CAS_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLS := mockservice.NewMockLoginService(ctrl)
	gomock.InOrder(
		mockLS.EXPECT().Login2CAS("alice", "123456").Return(&http.Client{}, nil).Times(1),
	)

	if err := performLogin(mockLS, "alice", "123456"); err != nil {
		t.Fatalf("期望没有错误，实际: %v", err)
	}
}

func TestLoginController_Login2CAS_Failure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLS := mockservice.NewMockLoginService(ctrl)
	gomock.InOrder(
		mockLS.EXPECT().Login2CAS("alice", gomock.Not("123456")).Return(nil, ErrLoginFailed).Times(1),
	)

	if err := performLogin(mockLS, "alice", "wrongPassword"); !errors.Is(err, ErrLoginFailed) {
		t.Fatalf("期望返回错误，实际为 nil")
	}
}
