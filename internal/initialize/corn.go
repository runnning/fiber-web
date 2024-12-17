package initialize

import (
	"fiber_web/pkg/cron"
	"time"

	"go.uber.org/zap"
)

var (
	defaultScheduler *cron.Scheduler
)

// InitScheduler 初始化定时任务调度器
func InitScheduler(logger *zap.Logger) error {
	defaultScheduler = cron.NewScheduler(logger)

	// 添加示例任务
	if err := defaultScheduler.AddTask(
		"example-task",
		"*/5 * * * * *", // 每5秒执行一次
		func() error {
			logger.Info("定时任务执行")
			return nil
		},
		10*time.Second,
	); err != nil {
		return err
	}

	// 启动调度器
	defaultScheduler.Start()
	return nil
}

// GetScheduler 获取默认调度器实例
func GetScheduler() *cron.Scheduler {
	return defaultScheduler
}

// StopScheduler 停止调度器
func StopScheduler() {
	if defaultScheduler != nil {
		defaultScheduler.Stop()
	}
}
