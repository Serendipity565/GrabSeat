package response

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

type Occupant struct {
	Title string `json:"title"`
	Name  string `json:"name"`
	State string `json:"state"`
	Start string `json:"start"`
	End   string `json:"end"`
}

type SearchResp struct {
	Data []Seat `json:"data"`
}
