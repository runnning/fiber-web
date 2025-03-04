package ctx

import (
	"fiber_web/pkg/query"
	"github.com/gofiber/fiber/v2"
)

// GetPagination 从上下文获取分页参数
func GetPagination(c *fiber.Ctx) *query.PageOption {
	p, ok := c.Locals("pagination").(*query.PageOption)
	if !ok {
		return query.NewPageOption(1, 10, true)
	}
	return p
}
