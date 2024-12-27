package initialize

import (
	"context"
	"fiber_web/apps/admin/internal/bootstrap"
	"fiber_web/pkg/config"
	"fiber_web/pkg/server"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// AppType 应用类型
type AppType string

const (
	AppTypeAPI   AppType = "api"
	AppTypeAdmin AppType = "admin"
	// 添加其他应用类型...
)

// Component 组件管理器
type Component struct {
	appType        AppType
	server         *server.FiberServer
	boot           *bootstrap.Bootstrapper
	infra          *Infra
	app            *App
	lifecycleHooks LifecycleHooks
}

func NewComponent(appType AppType) *Component {
	return &Component{
		appType:        appType,
		lifecycleHooks: NewLifecycleHooks(appType),
	}
}

// Initialize 初始化组件
func (c *Component) Initialize(ctx context.Context) error {
	// 初始化服务器
	c.server = server.NewFiberServer(
		server.WithReadTimeout(time.Second*30),
		server.WithWriteTimeout(time.Second*30),
		server.WithIdleTimeout(time.Second*30),
		server.WithEnv(config.Data.App.Env),
		server.WithAppName(config.Data.App.Name),
		server.WithServerHeader("Fiber"),
		server.WithBodyLimit(4>>20),
		server.WithDisableStartupMessage(false),
	)

	c.boot = bootstrap.New()

	// 注册生命周期钩子
	c.lifecycleHooks.RegisterHooks(c.boot, c.appType)

	// 按顺序添加组件
	c.infra = NewInfra()
	c.boot.AddComponent(c.infra)

	domain := NewDomain(c.infra)
	c.boot.AddComponent(domain)

	c.app = NewApp(c.infra, domain, c.server, c.boot, c.appType)
	c.boot.AddComponent(c.app)

	return c.boot.Bootstrap(ctx)
}

// Run 运行应用
func (c *Component) Run(ctx context.Context) error {
	// 启动所有组件
	if err := c.boot.Start(ctx); err != nil {
		return err
	}

	// 启动服务器
	go func() {
		// 这里不需要检查错误，因为正常关闭时也会返回错误
		if err := c.server.Start(config.Data.Server.Address); err != nil {
			// 只有在非正常关闭时才记录错误
			if err.Error() != "server closed" {
				log.Printf("Server error: %v\n", err)
			}
		}
	}()

	return c.waitForSignal(ctx)
}

// waitForSignal 等待终止信号
func (c *Component) waitForSignal(ctx context.Context) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		log.Println("Shutting down server...")
	case <-ctx.Done():
		log.Println("Server stopped due to context cancellation...")
	}

	return c.Shutdown()
}

// Shutdown 关闭应用
func (c *Component) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. 先关闭 HTTP 服务器
	if err := c.server.Shutdown(ctx); err != nil {
		log.Printf("Error during server shutdown: %v\n", err)
		return err
	}

	// 2. 关闭所有组件（包括基础设施）
	if err := c.boot.Shutdown(); err != nil {
		log.Printf("Error during bootstrap shutdown: %v\n", err)
		return err
	}

	return nil
}

func (c *Component) GetLifecycleHooks() LifecycleHooks {
	return c.lifecycleHooks
}
