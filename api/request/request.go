package request

// LoginRequest 登录请求参数
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// MFindVacantSeatsReq 查找今天的空座位
type MFindVacantSeatsReq struct {
	IsTomorrow bool   `json:"is_tomorrow"`
	StartTime  string `json:"start_time" binding:"required"`
	EndTime    string `json:"end_time" binding:"required"`
	KeyWord    string `json:"key_word,omitempty"`
}

type IsInLibraryReq struct {
	Username string `json:"username" binding:"required"`
}

type SeatToNameReq struct {
	SeatId string `json:"seat_id" binding:"required"`
}

type MGarbReq struct {
	StartTime  string `json:"start_time" binding:"required"`
	EndTime    string `json:"end_time" binding:"required"`
	IsTomorrow bool   `json:"is_tomorrow"`
	KeyWord    string `json:"key_word,omitempty"`
}

type ReserveReq struct {
	Data      string `json:"data" binding:"required"`
	StartTime string `json:"start_time" binding:"required"`
	EndTime   string `json:"end_time" binding:"required"`
	KeyWord   string `json:"key_word,omitempty"`
}
