package login

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func Login2CAS(username, password string) error {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: jar,
	}
	casLogin := "https://account.ccnu.edu.cn/cas/login"
	resp, _ := client.Get(casLogin)
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	lt, _ := doc.Find(`input[name="lt"]`).Attr("value")
	exec, _ := doc.Find(`input[name="execution"]`).Attr("value")

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

	resp, _ = client.Do(req)
	body, _ = io.ReadAll(resp.Body)
	resp.Body.Close()

	doc, _ = goquery.NewDocumentFromReader(strings.NewReader(string(body)))

	if errMsg := strings.TrimSpace(doc.Find("div#msg.errors").Text()); strings.Contains(errMsg, "您输入的用户名或密码有误") {
		return fmt.Errorf("登录失败，用户名或密码错误")
	} else if msg := strings.TrimSpace(doc.Find("#msg.success").Text()); strings.Contains(msg, "登录成功") {
		return nil
	} else {
		return fmt.Errorf("登录失败，未知错误")
	}
}

func extractHidden(html, name string) string {
	re := regexp.MustCompile(fmt.Sprintf(`name="%s" value="([^"]+)"`, name))
	match := re.FindStringSubmatch(html)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}
