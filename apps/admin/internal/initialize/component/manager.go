package component

import (
	"context"
	"fiber_web/apps/admin/internal/bootstrap"
	"fiber_web/apps/admin/internal/initialize"
	"fiber_web/apps/admin/internal/initialize/api"
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
	AppTypeAPI AppType = "api"
)

// Manager 应用管理器
type Manager struct {
	cfg     *config.Config
	appType AppType
	server  *server.FiberServer
	boot    *bootstrap.Bootstrapper
}

func NewManager(cfg *config.Config, appType AppType) *Manager {
	return &Manager{
		cfg:     cfg,
		appType: appType,
	}
}

func (m *Manager) Initialize(ctx context.Context) error {
	// 初始化服务器
	m.server = server.NewFiberServer(
		server.WithReadTimeout(time.Second*30),
		server.WithWriteTimeout(time.Second*30),
		server.WithIdleTimeout(time.Second*30),
		server.WithEnv(m.cfg.App.Env),
		server.WithAppName(m.cfg.App.Name),
		server.WithServerHeader("Fiber"),
		server.WithBodyLimit(4*1024*1024),
	)

	// 初始化引导程序
	m.boot = bootstrap.New()

	// 初始化组件
	infra := initialize.NewInfrastructure(m.cfg)
	infra.Register(m.boot)

	repo := initialize.NewRepository(infra)
	useCase := initialize.NewUseCase(repo)
	useCase.Register(m.boot)

	// 根据应用类型初始化路由
	if m.appType == AppTypeAPI {
		delivery := api.NewDelivery(useCase, infra, m.server.App())
		delivery.Register(m.boot)
	}

	return m.boot.Bootstrap(ctx)
}

// Run 运行应用
func (m *Manager) Run(ctx context.Context) error {
	// 启动服务器
	go func() {
		if err := m.server.Start(m.cfg.Server.Address); err != nil {
			log.Printf("Server error: %v\n", err)
		}
	}()

	return m.waitForSignal(ctx)
}

// waitForSignal 等待终止信号
func (m *Manager) waitForSignal(ctx context.Context) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		log.Println("Shutting down server...")
	case <-ctx.Done():
		log.Println("Server stopped due to context cancellation...")
	}

	return m.Shutdown()
}

// Shutdown 关闭应用
func (m *Manager) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := m.boot.Shutdown(); err != nil {
		log.Printf("Error during bootstrap shutdown: %v\n", err)
	}

	if err := m.server.Shutdown(ctx); err != nil {
		log.Printf("Error during server shutdown: %v\n", err)
		return err
	}

	return nil
}
