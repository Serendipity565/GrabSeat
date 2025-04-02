package response

import (
	"GrabSeat/service"
)

// MFindVacantSeatsResp 模糊查找今天的空座位
type MFindVacantSeatsResp struct {
	Seats []service.Seat `json:"seats"`
}

type LoginResp struct {
	Token string `json:"token"`
}

type SeatToNameResp struct {
	Ts []service.Ts `json:"ts"`
}

type Response struct {
	Msg  string      `json:"msg"`
	Code int         `json:"code"`
	Data interface{} `json:"data"`
}
