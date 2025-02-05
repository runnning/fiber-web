package middleware

import (
	"fiber_web/pkg/response"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// RateLimit 返回一个限流中间件
// max: 在指定时间窗口内的最大请求数
// expiration: 时间窗口大小
func RateLimit(max int, expiration time.Duration) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        max,
		Expiration: expiration,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP() // 使用IP作为限流的key
		},
		LimitReached: func(c *fiber.Ctx) error {
			return response.Error(c, fiber.StatusTooManyRequests, "请求太频繁，请稍后再试")
		},
	})
}
