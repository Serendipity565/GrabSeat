package service

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
func GetDevid() {

}
