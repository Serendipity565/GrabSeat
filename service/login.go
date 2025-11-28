package service

import (
	"errors"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/Serendipity565/GrabSeat/errs"
	"github.com/Serendipity565/GrabSeat/service/crawler"
)

//go:generate mockgen -destination=./mocks/mock_login_service.go -package=mocks github.com/Serendipity565/GrabSeat/service LoginService
type LoginService interface {
	Login2CAS(username, password string) (*http.Client, error)
}

type loginService struct {
}

func NewLoginService() LoginService {
	return &loginService{}
}

func (l *loginService) Login2CAS(username, password string) (*http.Client, error) {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar:     jar,
		Timeout: 10 * time.Second,
	}

	bodyBytes, err := crawler.FetchCASUrl(client, username, password)
	if err != nil {
		return nil, errs.CrawlerServerError(err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, errs.InternalServerError(err)
	}

	if errMsg := strings.TrimSpace(doc.Find("div#msg.errors").Text()); strings.Contains(errMsg, "您输入的用户名或密码有误") {
		err = errors.New("用户名或密码错误")
		return nil, errs.UserIdOrPasswordError(err)
	} else if msg := strings.TrimSpace(doc.Find("#msg.success").Text()); strings.Contains(msg, "登录成功") {
		return client, nil
	} else {
		err = errors.New(msg)
		return nil, errs.UserIdOrPasswordError(err)
	}
}
