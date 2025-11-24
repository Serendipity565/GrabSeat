package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/Serendipity565/GrabSeat/api/response"
	"github.com/Serendipity565/GrabSeat/pkg/logger"
	"go.uber.org/zap"
)

type GrabberService interface {
	GetClient(username, password string) (*http.Client, error)
	FindVacantSeats(client *http.Client, startTime, endTime, keyWord string, isTomorrow bool) ([]response.Seat, error)
	IsInLibrary(client *http.Client, name string) (*response.Occupant, error)
	SeatToName(client *http.Client, seatName string, isTomorrow bool) ([]response.Ts, error)
	Grab(client *http.Client, seatID, startTime, endTime string, isTomorrow bool) (bool, error)
	GrabSuccess(client *http.Client) (bool, error)
}

type clientEntry struct {
	client *http.Client
	expire time.Time
}

type grabberService struct {
	mu         sync.RWMutex
	cookiePool map[string]*clientEntry
	ttl        time.Duration
	log        logger.Logger
}

func NewGrabberService(log logger.Logger) GrabberService {
	return &grabberService{
		cookiePool: make(map[string]*clientEntry),
		ttl:        25 * time.Minute, // 比 CAS session TTL 略短一些，防止临界时间产生一些问题
		log:        log,
	}
}

// FindVacantSeats 寻找空闲座位
// keyword 模糊匹配关键字，为空则返回所有空闲座位
func (g *grabberService) FindVacantSeats(client *http.Client, startTime, endTime, keyWord string, isTomorrow bool) ([]response.Seat, error) {
	vacantSeats := make([]response.Seat, 0)
	for _, area := range Areas {
		dateTime := time.Now()
		if isTomorrow {
			dateTime = dateTime.Add(time.Hour * 24)
		}
		year, month, day := dateTime.Date()

		params := url.Values{}
		params.Set("byType", "devcls")
		params.Set("classkind", "8")
		params.Set("display", "fp")
		params.Set("md", "d")
		params.Set("room_id", area)
		params.Set("purpose", "")
		params.Set("selectOpenAty", "")
		params.Set("cld_name", "default")
		params.Set("date", fmt.Sprintf("%d-%02d-%02d", year, month, day))
		params.Set("fr_start", startTime)
		params.Set("fr_end", endTime)
		params.Set("act", "get_rsv_sta")
		params.Set("_", "16698463729090")
		requestURL := SearchUrl + "?" + params.Encode()
		req, err := http.NewRequest("GET", requestURL, nil)
		if err != nil {
			g.log.Error("create HTTP request error",
				zap.String("method", "GET"),
				zap.String("url", requestURL),
				zap.Error(err),
			)
			return nil, err
		}
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
			return nil, err
		}
		bodyBytes, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return nil, err
		}
		var bodyData response.SearchResp
		if err := json.Unmarshal(bodyBytes, &bodyData); err != nil {
			return nil, err
		}

		keyWord = strings.ToUpper(keyWord)
		for _, locationInfo := range bodyData.Data {
			if !strings.Contains(locationInfo.Title, keyWord) && keyWord != "" {
				continue
			}
			isConflict := false
			for _, t := range locationInfo.Ts {
				// t.Start, t.End, 的结构都是2024-12-10 08:20这样的
				// 需要忽略前面的日期部分，只比较时间部分
				start, end := t.Start[len(t.Start)-5:len(t.Start)], t.End[len(t.End)-5:len(t.End)]
				// 交叉或者包含
				if (startTime >= start && startTime < end) ||
					(endTime > start && endTime <= end) ||
					(startTime <= start && endTime >= end) {
					// 冲突，该座位不能预约
					isConflict = true
					break
				}
			}
			if !isConflict {
				// 不冲突
				vacantSeats = append(vacantSeats, locationInfo)
			}
		}
	}
	return vacantSeats, nil
}

// IsInLibrary 当前是否在图书馆
func (g *grabberService) IsInLibrary(client *http.Client, name string) (*response.Occupant, error) {
	for _, area := range Areas {
		dateTime := time.Now()
		year, month, day := dateTime.Date()

		params := url.Values{}
		params.Set("byType", "devcls")
		params.Set("classkind", "8")
		params.Set("display", "fp")
		params.Set("md", "d")
		params.Set("room_id", area)
		params.Set("purpose", "")
		params.Set("selectOpenAty", "")
		params.Set("cld_name", "default")
		params.Set("date", fmt.Sprintf("%d-%02d-%02d", year, month, day))
		params.Set("fr_start", "8:00")
		params.Set("fr_end", "22:00")
		params.Set("act", "get_rsv_sta")
		params.Set("_", "1763566369394")
		requestURL := SearchUrl + "?" + params.Encode()
		req, err := http.NewRequest("GET", requestURL, nil)
		if err != nil {
			g.log.Error("create HTTP request error",
				zap.String("method", "GET"),
				zap.String("url", requestURL),
				zap.Error(err),
			)
			return nil, err
		}
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
			return nil, err
		}
		bodyBytes, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return nil, err
		}
		var bodyData response.SearchResp
		if err := json.Unmarshal(bodyBytes, &bodyData); err != nil {
			return nil, err
		}

		for _, locationInfo := range bodyData.Data {
			for _, t := range locationInfo.Ts {
				if t.Owner == name && t.State == "doing" {
					return &response.Occupant{
						Title: locationInfo.Title,
						Name:  name,
						Start: t.Start[len(t.Start)-5:],
						End:   t.End[len(t.End)-5:],
					}, nil
				}
			}
		}
	}
	return nil, errors.New("未找到对应座位的预约信息")
}

// SeatToName 座位号转姓名: 查看该座位的预约信息，可以看到预约人是谁
func (g *grabberService) SeatToName(client *http.Client, seatName string, isTomorrow bool) ([]response.Ts, error) {
	for _, area := range Areas {
		dateTime := time.Now()
		if isTomorrow {
			dateTime = dateTime.Add(time.Hour * 24)
		}
		year, month, day := dateTime.Date()

		params := url.Values{}
		params.Set("byType", "devcls")
		params.Set("classkind", "8")
		params.Set("display", "fp")
		params.Set("md", "d")
		params.Set("room_id", area)
		params.Set("purpose", "")
		params.Set("selectOpenAty", "")
		params.Set("cld_name", "default")
		params.Set("date", fmt.Sprintf("%d-%02d-%02d", year, month, day))
		params.Set("fr_start", "8:00")
		params.Set("fr_end", "22:00")
		params.Set("act", "get_rsv_sta")
		params.Set("_", "1763566369394")
		requestURL := SearchUrl + "?" + params.Encode()
		req, err := http.NewRequest("GET", requestURL, nil)
		if err != nil {
			g.log.Error("create HTTP request error",
				zap.String("method", "GET"),
				zap.String("url", requestURL),
				zap.Error(err),
			)
			return nil, err
		}
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
			return nil, err
		}
		bodyBytes, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return nil, err
		}
		var bodyData response.SearchResp
		if err := json.Unmarshal(bodyBytes, &bodyData); err != nil {
			return nil, err
		}
		for _, locationInfo := range bodyData.Data {
			if locationInfo.Title == seatName {
				return locationInfo.Ts, nil
			}
		}
	}
	return nil, errors.New("未找到对应座位的预约信息")
}

// Grab 预约座位
func (g *grabberService) Grab(client *http.Client, seatID, startTime, endTime string, isTomorrow bool) (bool, error) {
	dateTime := time.Now()
	if isTomorrow {
		dateTime = dateTime.Add(time.Hour * 24)
	}
	year, month, day := dateTime.Date()

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
	req, err := http.NewRequest("POST", requestURL, nil)
	if err != nil {
		g.log.Error("create HTTP request error",
			zap.String("method", "POST"),
			zap.String("url", requestURL),
			zap.Error(err),
		)
		return false, err

	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Referer", "http://kjyy.ccnu.edu.cn/clientweb/xcus/ic2/Default.aspx")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	var respMap map[string]interface{}
	_ = json.Unmarshal(body, &respMap)
	// success {"ret":1,"act":"set_resv","msg":"操作成功！","data":null,"ext":null}
	if msg, ok := respMap["msg"].(string); ok && strings.Contains(msg, "操作成功") {
		return true, nil
	} else {
		return false, fmt.Errorf("预约失败: %s", msg)
	}
}

// GrabSuccess 预约是否成功
func (g *grabberService) GrabSuccess(client *http.Client) (bool, error) {
	params := url.Values{}
	params.Set("act", "get_History_resv")
	params.Set("strat", "90")
	params.Set("StatFlag", "New")
	params.Set("_", "1704815632495")
	requestURL := PersonUrl + "?" + params.Encode()
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		g.log.Error("create HTTP request error",
			zap.String("method", "GET"),
			zap.String("url", requestURL),
			zap.Error(err),
		)
		return false, err
	}
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Referer", "http://kjyy.ccnu.edu.cn/clientweb/xcus/ic2/Default.aspx")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false, err
	}

	// 进一步检查返回体，未登录情况会返回 JSON 提示
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		log.Fatalf("Failed to unmarshal response: %v", err)
	}

	msg, ok := response["msg"].(string)
	if !ok {

	}

	// success {
	//    "ret": 1,
	//    "act": "get_History_resv",
	//    "msg": "<tbody date='2025-11-24 14:00' state='4482' over='false'><tr class='head'><td colspan='6'><h3></h3><span><span class='orange uni_trans'>预约成功</span></span><span class='pull-right'><span class='grey'>2025-11-24 14:00</span></span></td></tr><tr class='content'><td><div class='box'><a>N1245</a><div class='grey'>南湖分馆一楼</div</div></td><td>姜高峰</td><td style='max-width:300px'><span class='grey'>个人预约</span></td><td><div><div><span class='grey'>开始:</span> <span class='text-primary'>11-24 19:00</span></div><div><span class='grey'>结束:</span> <span class='text-primary'>11-24 22:00</span></div></div></td><td><div><span style='color:green' class='uni_trans'>预约成功</span>,<span style='color:orange' class='uni_trans'>未生效</span>,<span style='color:green' class='uni_trans'>审核通过</span></div><div style='font-size:12px;color:#777;'></div></td><td class='text-center' style='vertical-align: middle;'><a class='click' rsvId='175901668' onclick='delRsv(this);'>取消</a></td></tr></tbody>",
	//    "data": null,
	//    "ext": null
	//}
	// failure {"ret":1,"act":"get_History_resv","msg":"<tbody><tr><td colspan='6' class='text-center'>没有数据</td></tr></tbody>","data":null,"ext":null}
	if msg == "" || strings.Contains(msg, "没有数据") {
		return false, nil
	}
	if strings.Contains(msg, "<tbody") {
		return true, nil
	}
	return false, fmt.Errorf("unexpected response: %s", msg)
}

// GetClient 获取或创建带有有效 cookie 的 http.Client
func (g *grabberService) GetClient(username, password string) (*http.Client, error) {
	// 先用读锁快速检查
	g.mu.RLock()
	entry, ok := g.cookiePool[username]
	if ok && entry != nil && time.Now().Before(entry.expire) {
		validate, _ := g.validateClient(entry.client)
		if validate {
			c := entry.client
			g.mu.RUnlock()
			return c, nil
		}
	}
	g.mu.RUnlock()

	// 升级到写锁 double-check
	g.mu.Lock()
	defer g.mu.Unlock()

	// double check
	entry, ok = g.cookiePool[username]
	//validate, _ = g.validateClient(entry.client)
	if ok && entry != nil && time.Now().Before(entry.expire) {
		validate, _ := g.validateClient(entry.client)
		if validate {
			return entry.client, nil
		}
	}

	// 需要创建/刷新
	newClient, err := g.GetLibraryClient(username, password)
	if err != nil {
		return nil, err
	}

	// 关闭旧 client 的 idle connections
	if ok && entry != nil && entry.client != nil {
		if tr, ok := entry.client.Transport.(*http.Transport); ok {
			tr.CloseIdleConnections()
		}
	}

	g.cookiePool[username] = &clientEntry{
		client: newClient,
		expire: time.Now().Add(g.ttl),
	}

	return newClient, nil
}

// CloseAll 用于优雅关闭：关闭所有 client 的空闲连接（不关闭正在使用的连接）
// TODO : 在程序退出时调用
func (g *grabberService) CloseAll() {
	g.mu.Lock()
	defer g.mu.Unlock()
	for _, e := range g.cookiePool {
		if e != nil && e.client != nil {
			if tr, ok := e.client.Transport.(*http.Transport); ok {
				tr.CloseIdleConnections()
			}
		}
	}
}

// GetLibraryClient 登入图书馆
func (g *grabberService) GetLibraryClient(username, password string) (*http.Client, error) {
	service := "http://kjyy.ccnu.edu.cn/loginall.aspx?page="

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Jar:       jar,
		Timeout:   15 * time.Second,
		Transport: &http.Transport{},
	}

	// 获取登录页面，读取 lt 和 execution 等隐藏字段
	loginURL := fmt.Sprintf("https://account.ccnu.edu.cn/cas/login?service=%s", url.QueryEscape(service))
	resp, err := client.Get(loginURL)
	if err != nil {
		return nil, err
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

	req, err := http.NewRequest("POST", loginURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Origin", "https://account.ccnu.edu.cn")
	req.Header.Set("Referer", loginURL)

	resp, err = client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 解析返回页面判断是否登录成功
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	doc, err = goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err == nil {
		errMsg := strings.TrimSpace(doc.Find("div#msg.errors").Text())
		if strings.Contains(errMsg, "您输入的用户名或密码有误") {
			return nil, errors.New(errMsg)
		}
	}

	// 返回已带 cookie 的 client，调用方可以直接使用 client.Jar.Cookies(...)
	return client, nil
}

// validateClient 状态验证，判断 cookie 是否有效
func (g *grabberService) validateClient(client *http.Client) (bool, error) {
	params := url.Values{}
	params.Add("act", "get_History_resv")
	params.Add("strat", "90")
	params.Add("StatFlag", "New")
	params.Add("_", "1763559000805")
	requestURL := PersonUrl + "?" + params.Encode()
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		g.log.Error("create HTTP request error",
			zap.String("method", "GET"),
			zap.String("url", requestURL),
			zap.Error(err),
		)
		return false, err
	}
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Referer", "http://kjyy.ccnu.edu.cn/clientweb/xcus/ic2/Default.aspx")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false, err
	}

	// 进一步检查返回体，未登录情况会返回 JSON 提示
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	// 尝试解析为 JSON，检查 msg 字段
	var m map[string]interface{}
	_ = json.Unmarshal(body, &m)
	if msg, ok := m["msg"].(string); ok {
		if strings.Contains(msg, "未登录") || strings.Contains(msg, "登录超时") || strings.Contains(msg, "session =null") {
			return false, err
		}
	}

	return true, nil
}
