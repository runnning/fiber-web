package component

import (
	"context"
	"fiber_web/apps/admin/internal/bootstrap"
	"fiber_web/apps/admin/internal/initialize"
	"fiber_web/apps/admin/internal/initialize/api"
	"fiber_web/pkg/config"
	"fiber_web/pkg/server"
	"github.com/gofiber/fiber/v2"
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

// DeliveryStrategy 路由策略接口
type DeliveryStrategy interface {
	Register(boot *bootstrap.Bootstrapper, useCase *initialize.UseCase, infra *initialize.Infrastructure, app *fiber.App)
}

// APIDelivery API路由策略
type APIDelivery struct{}

func (d *APIDelivery) Register(boot *bootstrap.Bootstrapper, useCase *initialize.UseCase, infra *initialize.Infrastructure, app *fiber.App) {
	delivery := api.NewDelivery(useCase, infra, app)
	delivery.Register(boot)
}

// DeliveryFactory 路由策略工厂
type DeliveryFactory struct{}

func (f *DeliveryFactory) CreateStrategy(appType AppType) DeliveryStrategy {
	switch appType {
	case AppTypeAPI:
		return &APIDelivery{}
	default:
		return nil
	}
}

// Manager 应用管理器
type Manager struct {
	cfg      *config.Config
	appType  AppType
	server   *server.FiberServer
	boot     *bootstrap.Bootstrapper
	factory  *DeliveryFactory
	strategy DeliveryStrategy
}

// NewManager 创建应用管理器
func NewManager(cfg *config.Config, appType AppType) *Manager {
	factory := &DeliveryFactory{}
	return &Manager{
		cfg:      cfg,
		appType:  appType,
		factory:  factory,
		strategy: factory.CreateStrategy(appType),
	}
}

// Initialize 初始化应用
func (m *Manager) Initialize(ctx context.Context) error {
	if err := m.initServer(); err != nil {
		return err
	}

	if err := m.initBootstrap(); err != nil {
		return err
	}

	if err := m.initApplication(ctx); err != nil {
		return err
	}

	return nil
}

// initServer 初始化服务器
func (m *Manager) initServer() error {
	m.server = server.NewFiberServer(
		server.WithReadTimeout(time.Second*30),
		server.WithWriteTimeout(time.Second*30),
		server.WithIdleTimeout(time.Second*30),
		server.WithEnv(m.cfg.App.Env),
		server.WithAppName(m.cfg.App.Name),
		server.WithServerHeader("Fiber"),
		server.WithBodyLimit(4*1024*1024),
	)
	return nil
}

// initBootstrap 初始化引导程序
func (m *Manager) initBootstrap() error {
	m.boot = bootstrap.New()
	return nil
}

// initApplication 初始化应用组件
func (m *Manager) initApplication(ctx context.Context) error {
	infra := initialize.NewInfrastructure(m.cfg)
	infra.Register(m.boot)

	repo := initialize.NewRepository(infra)
	useCase := initialize.NewUseCase(repo)
	useCase.Register(m.boot)

	if m.strategy != nil {
		m.strategy.Register(m.boot, useCase, infra, m.server.App())
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
