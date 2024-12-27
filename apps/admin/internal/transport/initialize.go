package transport

import (
	"fiber_web/apps/admin/internal/endpoint"
	"fiber_web/apps/admin/internal/middleware"
	"fiber_web/apps/admin/internal/usecase"
	"fiber_web/pkg/config"
	"fiber_web/pkg/validator"

	"github.com/gofiber/fiber/v2"
)

// RouterInitializer 路由初始化器
type RouterInitializer struct {
	app  *fiber.App
	uses *usecase.UseCases
}

// NewRouterInitializer 创建路由初始化器
func NewRouterInitializer(app *fiber.App, uses *usecase.UseCases) *RouterInitializer {
	return &RouterInitializer{
		app:  app,
		uses: uses,
	}
}

// InitAPIRoutes 初始化 API 路由
func (r *RouterInitializer) InitAPIRoutes() error {
	validator := validator.New(&validator.Config{Language: config.Data.App.Language})
	handlers := endpoint.InitHandlers(r.uses, validator)
	v1 := r.app.Group("/api/v1", middleware.CommMiddleware(config.Data.App.Env)...)
	{
		// 注册路由
		RegisterApiHttp(v1, handlers)
	}

	return nil
}

// InitAdminRoutes 初始化管理后台路由
func (r *RouterInitializer) InitAdminRoutes() error {
	// TODO: 实现管理后台路由
	return nil
}
