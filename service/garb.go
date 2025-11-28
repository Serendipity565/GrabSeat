package service

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/Serendipity565/GrabSeat/api/response"
	"github.com/Serendipity565/GrabSeat/errs"
	"github.com/Serendipity565/GrabSeat/pkg/logger"
	"github.com/Serendipity565/GrabSeat/service/crawler"
)

var (
	Areas = []string{"101699191", "101699189", "101699187", "101699179"}
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
	dateTime := time.Now()
	if isTomorrow {
		dateTime = dateTime.Add(time.Hour * 24)
	}
	year, month, day := dateTime.Date()
	for _, area := range Areas {
		bodyBytes, err := crawler.FetchSearchUrl(client, area, year, int(month), day, "8:00", "22:00")
		if err != nil {
			return nil, errs.CrawlerServerError(err)
		}
		var bodyData response.SearchResp
		if err = json.Unmarshal(bodyBytes, &bodyData); err != nil {
			return nil, errs.InternalServerError(err)
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
	dateTime := time.Now()
	year, month, day := dateTime.Date()
	for _, area := range Areas {
		bodyBytes, err := crawler.FetchSearchUrl(client, area, year, int(month), day, "8:00", "22:00")
		if err != nil {
			return nil, errs.CrawlerServerError(err)
		}
		var bodyData response.SearchResp
		if err = json.Unmarshal(bodyBytes, &bodyData); err != nil {
			return nil, errs.InternalServerError(err)
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
	// 找不到返回为空
	return nil, nil
}

// SeatToName 座位号转姓名: 查看该座位的预约信息，可以看到预约人是谁
func (g *grabberService) SeatToName(client *http.Client, seatName string, isTomorrow bool) ([]response.Ts, error) {
	dateTime := time.Now()
	if isTomorrow {
		dateTime = dateTime.Add(time.Hour * 24)
	}
	year, month, day := dateTime.Date()
	for _, area := range Areas {
		bodyBytes, err := crawler.FetchSearchUrl(client, area, year, int(month), day, "8:00", "22:00")
		if err != nil {
			return nil, errs.CrawlerServerError(err)
		}
		var bodyData response.SearchResp
		if err = json.Unmarshal(bodyBytes, &bodyData); err != nil {
			return nil, errs.InternalServerError(err)
		}
		for _, locationInfo := range bodyData.Data {
			if locationInfo.Title == seatName {
				return locationInfo.Ts, nil
			}
		}
	}
	return nil, nil
}

// Grab 预约座位
func (g *grabberService) Grab(client *http.Client, seatID, startTime, endTime string, isTomorrow bool) (bool, error) {
	dateTime := time.Now()
	if isTomorrow {
		dateTime = dateTime.Add(time.Hour * 24)
	}
	year, month, day := dateTime.Date()

	bodyBytes, err := crawler.FetchGrabUrl(client, seatID, year, int(month), day, startTime, endTime)
	if err != nil {
		return false, errs.CrawlerServerError(err)
	}
	var respMap map[string]interface{}
	err = json.Unmarshal(bodyBytes, &respMap)
	if err != nil {
		return false, errs.InternalServerError(err)
	}
	// success {"ret":1,"act":"set_resv","msg":"操作成功！","data":null,"ext":null}
	if msg, ok := respMap["msg"].(string); ok && strings.Contains(msg, "操作成功") {
		return true, nil
	} else {
		return false, errs.GrabSeatError(errors.New(msg))
	}
}

// GrabSuccess 预约是否成功
func (g *grabberService) GrabSuccess(client *http.Client) (bool, error) {
	bodyBytes, err := crawler.FetchPersonUrl(client)
	if err != nil {
		return false, errs.CrawlerServerError(err)
	}

	var respMap map[string]interface{}
	if err = json.Unmarshal(bodyBytes, &respMap); err != nil {
		return false, errs.InternalServerError(err)
	}

	// success {
	//    "ret": 1,
	//    "act": "get_History_resv",
	//    "msg": "<tbody date='2025-11-24 14:00' state='4482' over='false'><tr class='head'><td colspan='6'><h3></h3><span><span class='orange uni_trans'>预约成功</span></span><span class='pull-right'><span class='grey'>2025-11-24 14:00</span></span></td></tr><tr class='content'><td><div class='box'><a>N1245</a><div class='grey'>南湖分馆一楼</div</div></td><td>姜高峰</td><td style='max-width:300px'><span class='grey'>个人预约</span></td><td><div><div><span class='grey'>开始:</span> <span class='text-primary'>11-24 19:00</span></div><div><span class='grey'>结束:</span> <span class='text-primary'>11-24 22:00</span></div></div></td><td><div><span style='color:green' class='uni_trans'>预约成功</span>,<span style='color:orange' class='uni_trans'>未生效</span>,<span style='color:green' class='uni_trans'>审核通过</span></div><div style='font-size:12px;color:#777;'></div></td><td class='text-center' style='vertical-align: middle;'><a class='click' rsvId='175901668' onclick='delRsv(this);'>取消</a></td></tr></tbody>",
	//    "data": null,
	//    "ext": null
	//}
	// failure {"ret":1,"act":"get_History_resv","msg":"<tbody><tr><td colspan='6' class='text-center'>没有数据</td></tr></tbody>","data":null,"ext":null}
	if msg, ok := respMap["msg"].(string); ok && strings.Contains(msg, "<tbody") {
		return true, nil
	} else {
		return false, errs.GetHistoryError(errors.New(msg))
	}
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
		} else {
			return nil, errs.UnauthorizedError(errors.New("cookie 已失效，请重新登录"))
		}
	}
	g.mu.RUnlock()

	// 升级到写锁 double-check
	g.mu.Lock()
	defer g.mu.Unlock()

	// double check
	entry, ok = g.cookiePool[username]
	if ok && entry != nil && time.Now().Before(entry.expire) {
		validate, _ := g.validateClient(entry.client)
		if validate {
			return entry.client, nil
		} else {
			return nil, errs.UnauthorizedError(errors.New("cookie 已失效，请重新登录"))
		}
	}

	// 需要创建或刷新
	newClient, err := g.getLibraryClient(username, password)
	if err != nil {
		return nil, errs.CreateClientError(err)
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
func (g *grabberService) getLibraryClient(username, password string) (*http.Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, errs.InternalServerError(err)
	}
	client := &http.Client{
		Jar:       jar,
		Timeout:   15 * time.Second,
		Transport: &http.Transport{},
	}

	bodyBytes, err := crawler.FetchLibraryLoginUrl(client, username, password)
	if err != nil {
		return nil, errs.CrawlerServerError(err)
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, errs.InternalServerError(err)
	}
	errMsg := strings.TrimSpace(doc.Find("div#msg.errors").Text())
	if strings.Contains(errMsg, "您输入的用户名或密码有误") {
		return nil, errs.UserIdOrPasswordError(errors.New(errMsg))
	}
	// 返回已带 cookie 的 client，调用方可以直接使用 client.Jar.Cookies(...)
	return client, nil
}

// validateClient 状态验证，判断 cookie 是否有效
func (g *grabberService) validateClient(client *http.Client) (bool, error) {
	bodyBytes, err := crawler.FetchPersonUrl(client)
	if err != nil {
		return false, errs.CrawlerServerError(err)
	}

	// 尝试解析为 JSON，检查 msg 字段
	var m map[string]interface{}
	err = json.Unmarshal(bodyBytes, &m)
	if err != nil {
		return false, errs.InternalServerError(err)
	}
	if msg, ok := m["msg"].(string); ok {
		if strings.Contains(msg, "未登录") || strings.Contains(msg, "登录超时") || strings.Contains(msg, "session =null") {
			return false, errs.UnauthorizedError(errors.New(msg))
		}
	}

	return true, nil
}
