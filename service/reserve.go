package service

import (
	"time"
)

// BeforeDate 返回前一天的 18:00
func BeforeDate(date string) (time.Time, error) {
	// 解析输入的日期字符串
	layout := "2006-01-02"
	t, err := time.Parse(layout, date)
	if err != nil {
		return time.Time{}, err
	}

	// 计算前一天的 18:00
	prevDay := t.AddDate(0, 0, -1)
	targetTime := time.Date(prevDay.Year(), prevDay.Month(), prevDay.Day(), 18, 0, 0, 0, prevDay.Location())

	return targetTime, nil
}

// Reserve 预约明天座位
// 采用 goroutine + kafka 的方式进行预约
// 先往 sleep 一定时间，然后往 kafka 发送预约请求
func Reserve(DevId string, delay time.Duration) {
	go func() {
		time.Sleep(delay)
		// 预约座位

	}()
}
