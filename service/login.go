package service

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

//go:generate mockgen -destination=/mocks/mock_login_service.go -package=mocks github.com/Serendipity565/service LoginService
type LoginService interface {
	Login2CAS(username, password string) (*http.Client, error)
}

type loginService struct {
}

func NewLoginService() LoginService {
	return &loginService{}
}

func (ls *loginService) Login2CAS(username, password string) (*http.Client, error) {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar:     jar,
		Timeout: 10 * time.Second,
	}

	casLogin := "https://account.ccnu.edu.cn/cas/login"

	// 第一次 GET 获取登录页面，提取lt和execution参数
	resp, err := client.Get(casLogin)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	lt, _ := doc.Find(`input[name="lt"]`).Attr("value")
	exec, _ := doc.Find(`input[name="execution"]`).Attr("value")

	// 构造登录表单
	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)
	data.Set("lt", lt)
	data.Set("execution", exec)
	data.Set("_eventId", "submit")
	data.Set("submit", "登录")

	req, _ := http.NewRequest("POST", casLogin, strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Origin", "https://account.ccnu.edu.cn")
	req.Header.Set("Referer", casLogin)

	// 提交登录
	resp, _ = client.Do(req)
	body, _ = io.ReadAll(resp.Body)
	resp.Body.Close()

	doc, _ = goquery.NewDocumentFromReader(strings.NewReader(string(body)))

	if errMsg := strings.TrimSpace(doc.Find("div#msg.errors").Text()); strings.Contains(errMsg, "您输入的用户名或密码有误") {
		return nil, fmt.Errorf("登录失败，用户名或密码错误")
	} else if msg := strings.TrimSpace(doc.Find("#msg.success").Text()); strings.Contains(msg, "登录成功") {
		return client, nil
	} else {
		return nil, fmt.Errorf("登录失败，未知错误")
	}
}
