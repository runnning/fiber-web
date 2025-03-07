package ctx

import (
	"fiber_web/pkg/query"

	"github.com/gofiber/fiber/v2"
)

// GetPagination 从上下文获取分页参数
func GetPagination(c *fiber.Ctx) *query.Pagination {
	p, ok := c.Locals("pagination").(*query.Pagination)
	if !ok {
		return query.NewPagination(1, 10)
	}
	return p
}
