package middleware

import (
	"fiber_web/pkg/query"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// Pagination 分页中间件
func Pagination() fiber.Handler {
	return func(c *fiber.Ctx) error {
		page, _ := strconv.Atoi(c.Query("page", "1"))
		pageSize, _ := strconv.Atoi(c.Query("size", "10"))
		noCount := c.Query("no_count") == "true"

		c.Locals("pagination", query.NewPageOption(page, pageSize, noCount))
		return c.Next()
	}
}