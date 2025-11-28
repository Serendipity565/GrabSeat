package crawler

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var (
	CASUrl          = "https://account.ccnu.edu.cn/cas/login"
	LibraryLoginUrl = "https://account.ccnu.edu.cn/cas/login?service=http://kjyy.ccnu.edu.cn/loginall.aspx?page="
	SearchUrl       = "http://kjyy.ccnu.edu.cn/ClientWeb/pro/ajax/device.aspx"
	GrabUrl         = "http://kjyy.ccnu.edu.cn/ClientWeb/pro/ajax/reserve.aspx"
	PersonUrl       = "http://kjyy.ccnu.edu.cn/ClientWeb/pro/ajax/center.aspx"
)

func FetchCASUrl(client *http.Client, username, password string) ([]byte, error) {
	resp, err := client.Get(CASUrl)
	if err != nil {
		return nil, errors.New("获取CAS登录页面失败: " + err.Error())
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("读取CAS登录页面失败: " + err.Error())
	}

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

	req, _ := http.NewRequest("POST", CASUrl, strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Origin", "https://account.ccnu.edu.cn")
	req.Header.Set("Referer", CASUrl)

	// 提交登录
	resp, err = client.Do(req)
	if err != nil {
		return nil, errors.New("CAS登录请求失败: " + err.Error())
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("读取CAS登录响应失败: " + err.Error())
	}
	return bodyBytes, nil
}

func FetchLibraryLoginUrl(client *http.Client, username, password string) ([]byte, error) {
	resp, err := client.Get(LibraryLoginUrl)
	if err != nil {
		return nil, errors.New("获取图书馆登录页面失败: " + err.Error())
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	lt, ok := doc.Find(`input[name="lt"]`).Attr("value")
	if !ok {
		return nil, errors.New("login token lt not found")
	}
	exec, _ := doc.Find(`input[name="execution"]`).Attr("value")

	// 提交表单执行登录
	form := url.Values{}
	form.Set("username", username)
	form.Set("password", password)
	form.Set("lt", lt)
	form.Set("execution", exec)
	form.Set("_eventId", "submit")
	form.Set("submit", "登录")

	req, _ := http.NewRequest("POST", LibraryLoginUrl, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Origin", "https://account.ccnu.edu.cn")
	req.Header.Set("Referer", LibraryLoginUrl)

	resp, err = client.Do(req)
	if err != nil {
		return nil, errors.New("图书馆登录请求失败" + err.Error())
	}
	defer resp.Body.Close()
	// 解析返回页面判断是否登录成功
	bodyByte, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("读取图书馆登录响应失败: " + err.Error())
	}
	return bodyByte, nil
}

func FetchSearchUrl(client *http.Client, Area string, year, month, day int, startTime, endTime string) ([]byte, error) {
	params := url.Values{}
	params.Set("byType", "devcls")
	params.Set("classkind", "8")
	params.Set("display", "fp")
	params.Set("md", "d")
	params.Set("room_id", Area)
	params.Set("purpose", "")
	params.Set("selectOpenAty", "")
	params.Set("cld_name", "default")
	params.Set("date", fmt.Sprintf("%d-%02d-%02d", year, month, day))
	params.Set("fr_start", startTime)
	params.Set("fr_end", endTime)
	params.Set("act", "get_rsv_sta")
	params.Set("_", "16698463729090")
	requestURL := SearchUrl + "?" + params.Encode()

	req, _ := http.NewRequest("GET", requestURL, nil)
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Proxy-Connection", "keep-alive")
	req.Header.Set("Referer", "http://kjyy.ccnu.edu.cn/clientweb/xcus/ic2/Default.aspx")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New("SearchUrl请求失败: " + err.Error())
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("读取SearchUrl响应失败: " + err.Error())
	}
	return bodyBytes, nil
}

func FetchGrabUrl(client *http.Client, seatID string, year, month, day int, startTime, endTime string) ([]byte, error) {
	params := url.Values{}
	params.Set("dialogid", "")
	params.Set("dev_id", seatID)
	params.Set("lab_id", "")
	params.Set("kind_id", "")
	params.Set("room_id", "")
	params.Set("type", "dev")
	params.Set("prop", "")
	params.Set("test_id", "")
	params.Set("term", "")
	params.Set("Vnumber", "")
	params.Set("classkind", "")
	params.Set("test_name", "")
	params.Set("start", fmt.Sprintf("%d-%02d-%02d %s", year, month, day, startTime))
	params.Set("end", fmt.Sprintf("%d-%02d-%02d %s", year, month, day, endTime))
	params.Set("start_time", "1000")
	params.Set("end_time", "2200")
	params.Set("up_file", "")
	params.Set("memo", "")
	params.Set("act", "set_resv")
	params.Set("_", "170481145010")
	requestURL := GrabUrl + "?" + params.Encode()

	req, _ := http.NewRequest("POST", requestURL, nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Referer", "http://kjyy.ccnu.edu.cn/clientweb/xcus/ic2/Default.aspx")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New("GrabUrl请求失败: " + err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("GrabUrl请求状态码异常: " + resp.Status)
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("读取GrabUrl响应失败: " + err.Error())
	}
	return bodyBytes, nil
}

func FetchPersonUrl(client *http.Client) ([]byte, error) {
	params := url.Values{}
	params.Set("act", "get_History_resv")
	params.Set("strat", "90")
	params.Set("StatFlag", "New")
	params.Set("_", "1704815632495")
	requestURL := PersonUrl + "?" + params.Encode()

	req, _ := http.NewRequest("GET", requestURL, nil)
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Referer", "http://kjyy.ccnu.edu.cn/clientweb/xcus/ic2/Default.aspx")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New("PersonUrl请求失败: " + err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("PersonUrl请求状态码异常: " + resp.Status)
	}
	// 进一步检查返回体，未登录情况会返回 JSON 提示
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("读取PersonUrl响应失败: " + err.Error())
	}
	return bodyBytes, nil
}
