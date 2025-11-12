package service

import (
	"github.com/robfig/cron/v3"
)

type Ticker struct {
	c *cron.Cron
}

func NewTicker() *Ticker {
	return &Ticker{
		c: cron.New(cron.WithSeconds()),
	}
}

// StartDailyTask 定时器服务
// cron 本身是异步的，不需要阻塞
//func (t Ticker) StartDailyTask() {
//	// 每天18:00执行
//	spec := "0 0 18 * * *"
//	_, errs := t.c.AddFunc(spec, task)
//	if errs != nil {
//		fmt.Println("定时任务添加失败:", errs)
//		return
//	}
//
//	t.c.Start()
//	fmt.Println("✅ 定时任务已启动，每天18:00执行")
//}
