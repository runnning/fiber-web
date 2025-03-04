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
		orderBy := c.Query("orderBy", "")
		order := c.Query("order", "")

		req := query.NewPageRequest(page, pageSize)
		req.OrderBy = orderBy
		req.Order = order

		c.Locals("pagination", req)
		return c.Next()
	}
}
