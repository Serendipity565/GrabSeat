package service

import (
	"fmt"
	"github.com/robfig/cron/v3"
)

// StartDailyTask 定时器服务
// cron 本身是异步的，不需要阻塞
func StartDailyTask(task func()) {
	c := cron.New(cron.WithSeconds()) // 开启秒级支持

	// 每天18:00执行
	spec := "0 0 18 * * *"
	_, err := c.AddFunc(spec, task)
	if err != nil {
		fmt.Println("定时任务添加失败:", err)
		return
	}

	c.Start()
	fmt.Println("✅ 定时任务已启动，每天18:00执行")
}
