package initialize

import (
	"context"
	"fiber_web/apps/admin/internal/bootstrap"
	"fiber_web/apps/admin/internal/transport"
	"fiber_web/pkg/server"
	"fmt"
)

// App 应用初始化器
type App struct {
	infra   *Infra
	domain  *Domain
	servers map[string]*server.FiberServer
	boot    *bootstrap.Bootstrapper
	appType AppType
}

func NewApp(infra *Infra, domain *Domain, servers map[string]*server.FiberServer, boot *bootstrap.Bootstrapper, appType AppType) *App {
	return &App{
		infra:   infra,
		domain:  domain,
		servers: servers,
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

// servers 返回主服务器
func (a *App) server(name string) *server.FiberServer {
	s, ok := a.servers[name]
	if !ok {
		panic("server not found")
	}
	return s
}

func (a *App) initRoutes(ctx context.Context) error {
	switch a.appType {
	case AppTypeAPI:
		return transport.NewRouterInitializer(a.server("admin").App(), a.domain.Uses).InitAPIRoutes()
	case AppTypeAdmin:
		return transport.NewRouterInitializer(a.server("admin").App(), a.domain.Uses).InitAdminRoutes()
	default:
		return fmt.Errorf("unsupported app type: %s", a.appType)
	}
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
