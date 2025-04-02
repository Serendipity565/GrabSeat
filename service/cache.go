package service

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"net/url"
)

// Cache 缓存管理结构体
type Cache struct {
	deviceCache map[string]string
}

// Device 设备信息结构体
type Device struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ApiResponse 响应结构体
type ApiResponse struct {
	Ret  int `json:"ret"`
	Data struct {
		Objs []Device `json:"objs"` // 设备列表
	} `json:"data"`
}

// GetDevid 设置座位名称与ID对应的缓存
// TODO 暂时用不到，之后在实现具体功能
func (g *Grabber) GetDevid() {
	deviceMap := make(map[string]string)
	// 设置请求 URL
	baseURL, _ := url.Parse("http://kjyy.ccnu.edu.cn/ClientWeb/pro/ajax/device.aspx")

	for _, area := range g.areas {
		params := url.Values{}
		params.Set("byType", "devcls")
		params.Set("classkind", "8")
		params.Set("display", "fp")
		params.Set("md", "d")
		params.Set("room_id", area)
		params.Set("purpose", "")
		params.Set("selectOpenAty", "")
		params.Set("cld_name", "default")
		params.Set("act", "get_dev_coord")
		params.Set("_", "16698463729091")

		cookies := g.authClient.Jar.Cookies(baseURL)
		client := resty.New()
		resp, _ := client.SetCookies(cookies).R().
			SetQueryParamsFromValues(params).
			SetHeader("Accept", "application/json, text/javascript, */*; q=0.01").
			SetHeader("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6").
			SetHeader("Connection", "keep-alive").
			SetHeader("Referer", "http://kjyy.ccnu.edu.cn/clientweb/xcus/ic2/Default.aspx").
			SetHeader("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0").
			SetHeader("X-Requested-With", "XMLHttpRequest").
			Get(baseURL.String())

		var result ApiResponse
		err := json.Unmarshal(resp.Body(), &result)
		if err != nil {
			fmt.Printf("JSON 解析失败: %v\n", err)
		}

		json.Unmarshal(resp.Body(), &result)

		// 遍历设备列表，存入 map
		for _, device := range result.Data.Objs {
			deviceMap[device.Name] = device.ID
		}
	}
	return

}
