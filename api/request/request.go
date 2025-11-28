package request

// LoginRequest 登录请求参数
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// FindVacantSeatsReq 查找今天的空座位
type FindVacantSeatsReq struct {
	IsTomorrow *bool  `json:"is_tomorrow" binding:"required"`
	StartTime  string `json:"start_time" binding:"required"`
	EndTime    string `json:"end_time" binding:"required"`
	KeyWord    string `json:"key_word,omitempty"` // 可选参数，模糊搜索关键词
}

type IsInLibraryReq struct {
	StudentName string `json:"student_name" binding:"required"` // 学生姓名
}

type SeatToNameReq struct {
	SeatName   string `json:"seat_name" binding:"required"` // 座位名称,例如 N1224
	IsTomorrow *bool  `json:"is_tomorrow" binding:"required"`
}

type GarbReq struct {
	StartTime  string `json:"start_time" binding:"required"`
	EndTime    string `json:"end_time" binding:"required"`
	IsTomorrow *bool  `json:"is_tomorrow" binding:"required"`
	SeatID     string `json:"seat_id,omitempty"` // 可选参数,指定座位ID,不指定则自动选择
}

type ReserveReq struct {
	Data      string `json:"data" binding:"required"`
	StartTime string `json:"start_time" binding:"required"`
	EndTime   string `json:"end_time" binding:"required"`
	KeyWord   string `json:"key_word,omitempty"`
}
