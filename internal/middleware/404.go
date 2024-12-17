package middleware

import (
	"fiber_web/pkg/logger"
	"fiber_web/pkg/response"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// NotFound 返回一个 404 处理中间件
func NotFound() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 获取所有注册的路由
		routes := c.App().GetRoutes()

		// 检查当前请求路径是否在路由列表中
		found := false
		for _, route := range routes {
			if route.Path == c.Path() && route.Method == c.Method() {
				found = true
				break
			}
		}

		// 如果路由未找到
		if !found {
			// 记录 404 日志
			logger.Warn("Route not found",
				zap.String("request_id", c.Get("X-Request-ID")),
				zap.String("path", c.Path()),
				zap.String("method", c.Method()),
				zap.String("ip", c.IP()),
				zap.String("user_agent", c.Get("User-Agent")),
			)

			// 构建错误消息
			message := "Route " + c.Method() + " " + c.Path() + " not found"

			// 返回 404 响应
			return response.NotFound(c, message)
		}

		return c.Next()
	}
}
