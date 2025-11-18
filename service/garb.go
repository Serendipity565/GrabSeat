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
	"github.com/anaskhan96/soup"
	"github.com/go-resty/resty/v2"
)

type GrabberService interface {
	GetClient(username, password string) *http.Client
	FindOneVacantSeat() string
	FindVacantSeats() []Seat
	MFindVacantSeats(keyword string) []Seat
	IsInLibrary(name string) *Occupant
	SeatToName(seat string) []Ts
	Grab(devId string)
	GrabSuccess() bool
	StartFlushClient(username, password string, dur time.Duration)
}

type clientEntry struct {
	client *http.Client
	expire time.Time
}

type grabberService struct {
	mu         sync.RWMutex
	cookiePool map[string]*clientEntry
	ttl        time.Duration
}

func NewGrabberService() GrabberService {
	return &grabberService{
		cookiePool: make(map[string]*clientEntry),
		ttl:        25 * time.Minute, // 比 CAS session TTL 略短一些，防止临界时间产生一些问题
	}
}

func (g *grabberService) GetClient(username, password string) *http.Client {
	//TODO implement me
	panic("implement me")
}

func (g *grabberService) FindOneVacantSeat() string {
	//TODO implement me
	panic("implement me")
}

func (g *grabberService) FindVacantSeats() []Seat {
	//TODO implement me
	panic("implement me")
}

func (g *grabberService) MFindVacantSeats(keyword string) []Seat {
	//TODO implement me
	panic("implement me")
}

func (g *grabberService) IsInLibrary(name string) *Occupant {
	//TODO implement me
	panic("implement me")
}

func (g *grabberService) SeatToName(seat string) []Ts {
	//TODO implement me
	panic("implement me")
}

func (g *grabberService) Grab(devId string) {
	//TODO implement me
	panic("implement me")
}

func (g *grabberService) GrabSuccess() bool {
	//TODO implement me
	panic("implement me")
}

func (g *grabberService) StartFlushClient(username, password string, dur time.Duration) {
	//TODO implement me
	panic("implement me")
}

func (g *grabberService) getClient(username, password string) (*http.Client, error) {
	// 先用读锁快速检查
	g.mu.RLock()
	entry, ok := g.cookiePool[username]
	if ok && entry != nil && time.Now().Before(entry.expire) && validateClient(entry.client) {
		c := entry.client
		g.mu.RUnlock()
		return c, nil
	}
	g.mu.RUnlock()

	// 升级到写锁 double-check
	g.mu.Lock()
	defer g.mu.Unlock()

	// double check
	entry, ok = g.cookiePool[username]
	if ok && entry != nil && time.Now().Before(entry.expire) && validateClient(entry.client) {
		return entry.client, nil
	}

	// 需要创建/刷新
	newClient, err := getLibraryClient(username, password)
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

// closeAll 用于优雅关闭：关闭所有 client 的空闲连接（不关闭正在使用的连接）
func (g *grabberService) closeAll() {
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

// getLibraryClient 登入图书馆
func getLibraryClient(username, password string) (*http.Client, error) {
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

// 状态验证
func validateClient(client *http.Client) bool {
	// 发送一个轻量级请求验证session是否有效
	testURL := "http://kjyy.ccnu.edu.cn/ClientWeb/pro/ajax/center.aspx?act=get_History_resv&strat=90&StatFlag=New&_=1763436057006"
	resp, err := client.Get(testURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false
	}

	// 进一步检查返回体，未登录情况会返回 JSON 提示
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false
	}

	// 尝试解析为 JSON，检查 msg 字段
	var m map[string]interface{}
	_ = json.Unmarshal(body, &m)
	if msg, ok := m["msg"].(string); ok {
		if strings.Contains(msg, "未登录") || strings.Contains(msg, "登录超时") || strings.Contains(msg, "session =null") {
			return false
		}
	}

	return true
}

type Grabber struct {
	authClient *http.Client
	areas      []string // 目标区域
	isTomorrow bool     // 是否是明天, false：默认抢今天
	start      string   // 目标起始时间
	end        string   // 目标结束时间
	searchUrl  string
	grabUrl    string
}

// StartFlushClient 启动一个 goroutine 定时刷新 authClient
func (g *Grabber) StartFlushClient(username, password string, dur time.Duration) {
	authClient := g.GetClient(username, password)
	g.setAuthClient(authClient)
	go func() {
		for {
			authClient := g.GetClient(username, password)
			g.setAuthClient(authClient)
			time.Sleep(dur)
		}
	}()
}

func (g *Grabber) setAuthClient(authClient *http.Client) {
	g.authClient = authClient
}

// NewGrabber 创建一个新的 Grabber 实例
func NewGrabber(areas []string, isTomorrow bool, start string, end string) *Grabber {
	return &Grabber{
		areas:      areas,
		isTomorrow: isTomorrow,
		start:      start,
		end:        end,
		searchUrl:  "http://kjyy.ccnu.edu.cn/ClientWeb/pro/ajax/device.aspx",
		grabUrl:    "http://kjyy.ccnu.edu.cn/ClientWeb/pro/ajax/reserve.aspx",
	}
}

type Seat struct {
	Title string `json:"title"`
	Ts    []Ts   `json:"ts"`
	DevId string `json:"devId"`
}

type Ts struct { // 预约信息
	Start string `json:"start"`
	End   string `json:"end"`
	Owner string `json:"owner"`
	State string `json:"state"`
}

type searchResp struct {
	Data []Seat
}

// FindOneVacantSeat 寻找一个空闲座位
func (g *Grabber) FindOneVacantSeat() string {
	for _, area := range g.areas {
		dateTime := time.Now()
		if g.isTomorrow {
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
		params.Set("fr_start", g.start)
		params.Set("fr_end", g.end)
		params.Set("act", "get_rsv_sta")
		params.Set("_", "16698463729090")
		parsedSearchUrl, _ := url.Parse(g.searchUrl)
		cookies := g.authClient.Jar.Cookies(parsedSearchUrl)

		client, bodyData := resty.New(), &searchResp{}
		_, _ = client.SetCookies(cookies).R().SetQueryParamsFromValues(params).SetResult(&bodyData).Get(g.searchUrl)

		for _, locationInfo := range bodyData.Data {
			isConflict := false
			for _, t := range locationInfo.Ts {
				// t.Start, t.End, 的结构都是2024-12-10 08:20这样的
				// 需要忽略前面的日期部分，只比较时间部分
				start, end := t.Start[len(t.Start)-5:len(t.Start)], t.End[len(t.End)-5:len(t.End)]
				if (g.start >= start && g.start < end) || (g.end > start && g.end <= end) || (g.start <= start && g.end >= end) {
					// 冲突，该座位不能预约
					isConflict = true
					break
				}
			}
			if !isConflict {
				// 不冲突
				return locationInfo.DevId
			}
		}
	}
	return ""
}

// FindVacantSeats 寻找所有空闲座位
func (g *Grabber) FindVacantSeats() []Seat {
	vacantSeats := make([]Seat, 0)
	for _, area := range g.areas {
		dateTime := time.Now()
		if g.isTomorrow {
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
		params.Set("fr_start", g.start)
		params.Set("fr_end", g.end)
		params.Set("act", "get_rsv_sta")
		params.Set("_", "16698463729090")
		parsedSearchUrl, _ := url.Parse(g.searchUrl)
		cookies := g.authClient.Jar.Cookies(parsedSearchUrl)

		client, bodyData := resty.New(), &searchResp{}
		_, _ = client.SetCookies(cookies).R().SetQueryParamsFromValues(params).SetResult(&bodyData).Get(g.searchUrl)

		for _, locationInfo := range bodyData.Data {
			isConflict := false
			for _, t := range locationInfo.Ts {
				// t.Start, t.End, 的结构都是2024-12-10 08:20这样的
				// 需要忽略前面的日期部分，只比较时间部分
				start, end := t.Start[len(t.Start)-5:len(t.Start)], t.End[len(t.End)-5:len(t.End)]
				if (g.start >= start && g.start < end) || (g.end > start && g.end <= end) || (g.start <= start && g.end >= end) {
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
	return vacantSeats
}

// MFindVacantSeats 根据模糊关键字查找设备，只查询空的位置
// 只是在原有的基础上增加了一个关键字的判断
func (g *Grabber) MFindVacantSeats(keyword string) []Seat {
	vacantSeats := make([]Seat, 0)
	for _, area := range g.areas {
		dateTime := time.Now()
		if g.isTomorrow {
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
		params.Set("fr_start", g.start)
		params.Set("fr_end", g.end)
		params.Set("act", "get_rsv_sta")
		params.Set("_", "16698463729090")
		parsedSearchUrl, _ := url.Parse(g.searchUrl)
		cookies := g.authClient.Jar.Cookies(parsedSearchUrl)

		client, bodyData := resty.New(), &searchResp{}
		_, _ = client.SetCookies(cookies).R().SetQueryParamsFromValues(params).SetResult(&bodyData).Get(g.searchUrl)

		keyword = strings.ToUpper(keyword)
		for _, locationInfo := range bodyData.Data {
			if !strings.Contains(locationInfo.Title, keyword) {
				continue
			}
			isConflict := false
			for _, t := range locationInfo.Ts {
				// t.Start, t.End, 的结构都是2024-12-10 08:20这样的
				// 需要忽略前面的日期部分，只比较时间部分
				start, end := t.Start[len(t.Start)-5:len(t.Start)], t.End[len(t.End)-5:len(t.End)]
				// 交叉或者包含
				if (g.start >= start && g.start < end) || (g.end > start && g.end <= end) || (g.start <= start && g.end >= end) {
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
	return vacantSeats
}

type Occupant struct {
	Title string
	Name  string
	Start string
	End   string
}

// IsInLibrary 当前是否在图书馆
func (g *Grabber) IsInLibrary(name string) *Occupant {
	for _, area := range g.areas {
		dateTime := time.Now()
		if g.isTomorrow {
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
		params.Set("fr_start", g.start)
		params.Set("fr_end", g.end)
		params.Set("act", "get_rsv_sta")
		params.Set("_", "16698463729090")
		parsedSearchUrl, _ := url.Parse(g.searchUrl)
		cookies := g.authClient.Jar.Cookies(parsedSearchUrl)

		client, bodyData := resty.New(), &searchResp{}
		_, _ = client.SetCookies(cookies).R().SetQueryParamsFromValues(params).SetResult(&bodyData).Get(g.searchUrl)

		for _, locationInfo := range bodyData.Data {
			for _, t := range locationInfo.Ts {
				if t.Owner == name && t.State == "doing" {
					return &Occupant{
						Title: locationInfo.Title,
						Name:  name,
						Start: t.Start[len(t.Start)-5:],
						End:   t.End[len(t.End)-5:],
					}
				}
			}
		}
	}
	return nil
}

// SeatToName 座位号转姓名: 查看该座位的预约信息，可以看到当前时间预约人是谁
func (g *Grabber) SeatToName(seat string) []Ts {
	for _, area := range g.areas {
		dateTime := time.Now()
		if g.isTomorrow {
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
		params.Set("fr_start", g.start)
		params.Set("fr_end", g.end)
		params.Set("act", "get_rsv_sta")
		params.Set("_", "16698463729090")
		parsedSearchUrl, _ := url.Parse(g.searchUrl)
		cookies := g.authClient.Jar.Cookies(parsedSearchUrl)

		client, bodyData := resty.New(), &searchResp{}
		_, _ = client.SetCookies(cookies).R().SetQueryParamsFromValues(params).SetResult(&bodyData).Get(g.searchUrl)

		for _, locationInfo := range bodyData.Data {
			if locationInfo.Title == seat {
				return locationInfo.Ts
			}
		}
	}
	return nil
}

// Grab 预约座位
func (g *Grabber) Grab(devId string) {
	dateTime := time.Now()
	if g.isTomorrow {
		dateTime = dateTime.Add(time.Hour * 24)
	}
	year, month, day := dateTime.Date()

	params := url.Values{}
	params.Set("dialogid", "")
	params.Set("dev_id", devId)
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
	params.Set("start", fmt.Sprintf("%d-%02d-%02d %s", year, month, day, g.start))
	params.Set("end", fmt.Sprintf("%d-%02d-%02d %s", year, month, day, g.end))
	params.Set("start_time", "1000")
	params.Set("end_time", "2200")
	params.Set("up_file", "")
	params.Set("memo", "")
	params.Set("act", "set_resv")
	params.Set("_", "17048114508")

	parsedUrl, _ := url.Parse(g.grabUrl)
	cookies := g.authClient.Jar.Cookies(parsedUrl)

	client := resty.New()
	_, _ = client.SetCookies(cookies).R().
		SetQueryParamsFromValues(params).
		SetHeader("Accept", "application/json, text/javascript, */*; q=0.01").
		SetHeader("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6").
		SetHeader("Connection", "keep-alive").
		SetHeader("Referer", "http://kjyy.ccnu.edu.cn/clientweb/xcus/ic2/Default.aspx").
		SetHeader("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0").
		SetHeader("X-Requested-With", "XMLHttpRequest").
		Get(parsedUrl.String())
}

// GetClient 登入图书馆
func (g *Grabber) GetClient(username, password string) *http.Client {
	resp, err := soup.Get("https://account.ccnu.edu.cn/cas/login?service=http://kjyy.ccnu.edu.cn/loginall.aspx?page=")
	if err != nil {
		log.Fatalf("Failed to get login page: %v", err)
	}
	doc := soup.HTMLParse(resp)
	jsessionID := doc.Find("body", "id", "cas").FindAll("script")[2].Attrs()["src"]
	ltValue := doc.Find("div", "class", "logo").FindAll("input")[2].Attrs()["value"]

	jar, _ := cookiejar.New(&cookiejar.Options{})
	client := &http.Client{
		Jar:     jar,
		Timeout: 5 * time.Second,
	}

	jsessionID = jsessionID[26:]
	loginURL := fmt.Sprintf("https://account.ccnu.edu.cn/cas/login;jsessionid=%v?service=http://kjyy.ccnu.edu.cn/loginall.aspx?page=", jsessionID)
	formData := fmt.Sprintf("username=%v&password=%v&lt=%v&execution=e1s1&_eventId=submit&submit=登录", username, password, ltValue)
	body := strings.NewReader(formData)

	req, _ := http.NewRequest("POST", loginURL, body)
	req.Header = http.Header{
		"Accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8"},
		"Accept-Encoding":           {"gzip, deflate, br"},
		"Accept-Language":           {"zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2"},
		"Cache-Control":             {"max-age=0"},
		"Connection":                {"keep-alive"},
		"Content-Length":            {"162"},
		"Content-Type":              {"application/x-www-form-urlencoded"},
		"Cookie":                    {"JSESSIONID=" + jsessionID},
		"Host":                      {"account.ccnu.edu.cn"},
		"Origin":                    {"https://account.ccnu.edu.cn"},
		"Referer":                   {"https://account.ccnu.edu.cn/cas/login?service=http://kjyy.ccnu.edu.cn/loginall.aspx?page="},
		"Sec-Fetch-Dest":            {"document"},
		"Sec-Fetch-Mode":            {"navigate"},
		"Sec-Fetch-Site":            {"same-origin"},
		"Sec-Fetch-User":            {"?1"},
		"Upgrade-Insecure-Requests": {"1"},
		"User-Agent":                {"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:107.0) Gecko/20100101 Firefox/107.0"},
		"sec-ch-ua":                 {""},
		"sec-ch-ua-mobile":          {"?0"},
		"sec-ch-ua-platform":        {"Windows"},
	}

	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to execute login request: %v", err)
	}
	defer res.Body.Close()

	return client
}

// GrabSuccess 二次验证
func (g *Grabber) GrabSuccess() bool {
	params := url.Values{}
	params.Set("act", "get_History_resv")
	params.Set("strat", "90")
	params.Set("StatFlag", "New")
	params.Set("_", "1704815632495")

	parsedUrl, _ := url.Parse("http://kjyy.ccnu.edu.cn/ClientWeb/pro/ajax/center.aspx")
	cookies := g.authClient.Jar.Cookies(parsedUrl)

	client := resty.New()
	resp, err := client.SetCookies(cookies).R().
		SetQueryParamsFromValues(params).
		SetHeader("Accept", "application/json, text/javascript, */*; q=0.01").
		SetHeader("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6").
		SetHeader("Connection", "keep-alive").
		SetHeader("Referer", "http://kjyy.ccnu.edu.cn/clientweb/xcus/ic2/Default.aspx").
		SetHeader("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0").
		SetHeader("X-Requested-With", "XMLHttpRequest").
		Get(parsedUrl.String())

	if err != nil {
		log.Fatalf("Failed to execute request: %v", err)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		log.Fatalf("Failed to unmarshal response: %v", err)
	}

	msg, ok := response["msg"].(string)
	if !ok {
		log.Fatalf("Unexpected response format")
	}

	return len(msg) > len("<tbody date='2024-01-09 13:53' state='1082265730")+10
}
