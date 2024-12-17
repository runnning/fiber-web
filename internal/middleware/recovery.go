package middleware

import (
	"fiber_web/pkg/logger"
	"fiber_web/pkg/response"
	"fmt"
	"runtime"
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// Recovery 返回一个自定义的恢复中间件
func Recovery() fiber.Handler {
	// 返回一个新的中间件处理器
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				err, ok := r.(error)
				if !ok {
					err = fmt.Errorf("%v", r)
				}

				// 获取堆栈信息
				stack := make([]byte, 4<<10) // 4KB
				length := runtime.Stack(stack, false)

				// 提取关键的堆栈信息
				stackStr := string(stack[:length])
				stackLines := strings.Split(stackStr, "\n")
				var relevantStack []string
				for i, line := range stackLines {
					if i < 7 && len(line) > 0 { // 只保留前几行关键信息
						relevantStack = append(relevantStack, strings.TrimSpace(line))
					}
				}

				// 记录错误日志
				logger.Error("Recovered from panic",
					zap.Error(err),
					zap.String("request_id", c.Get("X-Request-ID")),
					zap.String("path", c.Path()),
					zap.String("method", c.Method()),
					zap.String("ip", c.IP()),
					zap.Strings("stack", relevantStack),
				)

				// 在开发环境下返回详细错误信息
				if c.Get("APP_ENV") == "development" {
					_ = response.ServerError(c, fmt.Errorf("%v\nStack:\n%s", err, strings.Join(relevantStack, "\n")))
				} else {
					// 生产环境只返回简单错误信息
					_ = response.ServerError(c, fmt.Errorf("internal server error"))
				}
			}
		}()

		return c.Next()
	}
}
