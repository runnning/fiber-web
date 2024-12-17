package initialize

import (
	"context"
	"fiber_web/bootstrap"
	"fiber_web/pkg/config"
	"fiber_web/pkg/server"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Manager struct {
	cfg    *config.Config
	server *server.FiberServer
	boot   *bootstrap.Bootstrapper
}

func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		cfg: cfg,
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
	app := m.server.App()
	infra := NewInfrastructure(m.cfg)
	infra.Register(m.boot)

	repo := NewRepository(infra)
	useCase := NewUseCase(repo)
	useCase.Register(m.boot)

	delivery := NewDelivery(useCase, infra, app)
	delivery.Register(m.boot)

	return m.boot.Bootstrap(ctx)
}

func (m *Manager) Run(ctx context.Context) error {
	// 启动服务器
	go func() {
		if err := m.server.Start(m.cfg.Server.Address); err != nil {
			log.Printf("Server error: %v\n", err)
		}
	}()

	// 等待终止信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		log.Println("Shutting down server...")
	case <-ctx.Done():
		log.Println("Server stopped due to error...")
	}

	return m.Shutdown()
}

func (m *Manager) Shutdown() error {
	// 设置关闭超时
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 关闭引导程序
	if err := m.boot.Shutdown(); err != nil {
		log.Printf("Error during bootstrap shutdown: %v\n", err)
	}

	// 关闭服务器
	if err := m.server.Shutdown(ctx); err != nil {
		log.Printf("Error during server shutdown: %v\n", err)
		return err
	}

	return nil
}
