package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// CommMiddleware 返回通用中间件处理器列表
func CommMiddleware(env string) []fiber.Handler {
	return []fiber.Handler{
		// 添加请求ID中间件
		RequestID(),
		// Recover from panics
		Recovery(env),
		// CORS
		CORS(),
		// Logger
		Logger(),
		// NotFound
		NotFound(),
	}
}
