package middleware

import (
	"fiber_web/pkg/logger"
	"fiber_web/pkg/response"
	"strings"

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
		requestPath := c.Path()
		requestMethod := c.Method()

		for _, route := range routes {
			// 如果方法不匹配，继续下一个
			if route.Method != requestMethod {
				continue
			}

			// 检查路径是否完全匹配
			if route.Path == requestPath {
				found = true
				break
			}

			// 检查是否是参数路由
			routeParts := strings.Split(route.Path, "/")
			requestParts := strings.Split(requestPath, "/")

			// 如果路径段数不同，继续下一个
			if len(routeParts) != len(requestParts) {
				continue
			}

			// 逐段比较路径
			pathMatch := true
			for i := 0; i < len(routeParts); i++ {
				// 如果是参数段（以:或*开头）或者段完全匹配
				if strings.HasPrefix(routeParts[i], ":") ||
					strings.HasPrefix(routeParts[i], "*") ||
					routeParts[i] == requestParts[i] {
					continue
				}
				pathMatch = false
				break
			}

			if pathMatch {
				found = true
				break
			}
		}

		// 如果路由未找到
		if !found {
			// 记录 404 日志
			logger.Warn("Route not found",
				zap.String("request_id", c.Get("X-Request-ID")),
				zap.String("path", requestPath),
				zap.String("method", requestMethod),
				zap.String("ip", c.IP()),
				zap.String("user_agent", c.Get("User-Agent")),
			)

			// 构建错误消息
			message := "Route " + requestMethod + " " + requestPath + " not found"

			// 返回 404 响应
			return response.NotFound(c, message)
		}

		return c.Next()
	}
}
