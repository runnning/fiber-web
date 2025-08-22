package middleware

import (
	"time"

	"fiber_web/pkg/logger"

	"github.com/gofiber/fiber/v2"
)

// Logger 返回一个使用 zap 的日志中间件
func Logger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		path := c.Path()
		method := c.Method()

		// 处理请求
		err := c.Next()

		// 记录请求日志
		latency := time.Since(start)
		status := c.Response().StatusCode()
		reqID := c.GetRespHeader("X-Request-ID")

		logger.Info("HTTP Request",
			logger.String("request_id", reqID),
			logger.String("method", method),
			logger.String("path", path),
			logger.Int("status", status),
			logger.Duration("latency", latency),
			logger.String("ip", c.IP()),
			logger.String("user_agent", c.Get("User-Agent")),
		)

		return err
	}
}
