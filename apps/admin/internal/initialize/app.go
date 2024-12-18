package initialize

import (
	"context"
	"fiber_web/apps/admin/internal/bootstrap"
	"fiber_web/apps/admin/internal/endpoint"
	"fiber_web/apps/admin/internal/middleware"
	"fiber_web/pkg/auth"
	"fiber_web/pkg/server"
	"fiber_web/pkg/validator"
	"fmt"
)

// App 应用初始化器
type App struct {
	infra   *Infra
	domain  *Domain
	server  *server.FiberServer
	boot    *bootstrap.Bootstrapper
	appType AppType
}

func NewApp(infra *Infra, domain *Domain, server *server.FiberServer, boot *bootstrap.Bootstrapper, appType AppType) *App {
	return &App{
		infra:   infra,
		domain:  domain,
		server:  server,
		boot:    boot,
		appType: appType,
	}
}

// Init 实现 Component 接口
func (a *App) Init(ctx context.Context) error {
	a.infra.Logger.Info("Initializing application routes")
	// 根据应用类型初始化路由
	if err := a.initRoutes(ctx); err != nil {
		return fmt.Errorf("failed to initialize routes: %w", err)
	}
	a.infra.Logger.Info("Application routes initialized")
	return nil
}

func (a *App) initRoutes(ctx context.Context) error {
	switch a.appType {
	case AppTypeAPI:
		return a.initAPIRoutes(ctx)
	case AppTypeAdmin:
		return a.initAdminRoutes(ctx)
	default:
		return fmt.Errorf("unsupported app type: %s", a.appType)
	}
}

func (a *App) initAPIRoutes(ctx context.Context) error {
	validator := validator.New(&validator.Config{Language: a.infra.Config.App.Language})
	userHandler := endpoint.NewUserHandler(a.domain.Uses.User, validator)
	v1 := a.server.App().Group("/api/v1", middleware.CommMiddleware(a.infra.Config.App.Env)...)
	{
		v1.Post("/register", userHandler.Register)
		v1.Post("/login", userHandler.Login)
		v1.Get("/users", middleware.Pagination(), userHandler.ListUsers)
		v1.Get("/test", middleware.Pagination(), userHandler.TestUser)
		v1.Get("/users/me", middleware.Jwt(auth.NewJWTManager(a.infra.Config)), userHandler.GetProfile)
	}

	return nil
}

func (a *App) initAdminRoutes(ctx context.Context) error {
	// TODO: 实现管理后台路由
	return nil
}

// Start 实现 Component 接口
func (a *App) Start(ctx context.Context) error {
	// 不需要重复打印路由信息，fiber.go 会处理
	return nil
}

// Stop 实现 Component 接口
func (a *App) Stop(ctx context.Context) error {
	return nil
}
