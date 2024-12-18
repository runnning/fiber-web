package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/fiber/v2/utils"
)

// RequestID 返回一个设置请求ID的中间件
func RequestID() fiber.Handler {
	return requestid.New(requestid.Config{
		// 使用默认的 UUID 生成器
		Generator: func() string {
			return utils.UUID()
		},
		// 设置请求ID的响应头
		Header: "X-Request-ID",
		// 如果请求头中已存在请求ID，则使用已有的
		ContextKey: "request_id",
	})
}
