package initialize

import (
	"context"
	"fiber_web/pkg/config"
	"fiber_web/pkg/cron"
	"fiber_web/pkg/database"
	"fiber_web/pkg/logger"
	"fiber_web/pkg/queue"
	"fiber_web/pkg/redis"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

// Infra 基础设施
type Infra struct {
	Config *config.Config
	DB     *database.Database
	Redis  *redis.Client
	NSQ    *queue.Producer
	Logger *logger.Logger
	Cron   *cron.Scheduler
	mu     sync.RWMutex
}

func NewInfra(cfg *config.Config) *Infra {
	return &Infra{Config: cfg}
}

// Init 实现 Component 接口
func (i *Infra) Init(ctx context.Context) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	// 初始化日志
	if err := logger.InitLogger(i.Config.App.Env); err != nil {
		return err
	}
	i.Logger = logger.GetDefaultLogger()
	i.Logger.Info("Logger initialized")

	// 初始化数据库
	db, err := database.NewMySQL(i.Config)
	if err != nil {
		return err
	}
	i.DB = db
	i.Logger.Info("Database initialized")

	// 初始化 Redis
	redis, err := redis.NewClient(i.Config)
	if err != nil {
		return err
	}
	i.Redis = redis
	i.Logger.Info("Redis initialized")

	// 初始化 NSQ
	producer, err := queue.NewProducer(i.Config)
	if err != nil {
		return err
	}
	i.NSQ = producer
	i.Logger.Info("NSQ initialized")

	// 启动 Cron
	i.Cron = cron.NewScheduler(logger.GetLogger())
	i.Logger.Info("Cron initialized")

	return nil
}

// Start 实现 Component 接口
func (i *Infra) Start(ctx context.Context) error {
	i.Cron.Start()
	return nil
}

// Stop 实现 Component 接口
func (i *Infra) Stop(ctx context.Context) error {
	return i.Shutdown()
}

// Shutdown 关闭基础设施
func (i *Infra) Shutdown() error {
	i.mu.Lock()
	defer i.mu.Unlock()

	var errs []error

	// 关闭 Cron
	if i.Cron != nil {
		i.Cron.Stop()
		i.Logger.Info("Cron scheduler stopped")
	}

	// 关闭 NSQ
	if i.NSQ != nil {
		if err := i.NSQ.Stop(); err != nil {
			i.Logger.Error("Failed to stop NSQ", zap.Error(err))
			errs = append(errs, err)
		} else {
			i.Logger.Info("NSQ producer stopped")
		}
	}

	// 关闭 Redis
	if i.Redis != nil {
		if err := i.Redis.Close(); err != nil {
			i.Logger.Error("Failed to close Redis", zap.Error(err))
			errs = append(errs, err)
		} else {
			i.Logger.Info("Redis connection closed")
		}
	}

	// 关闭数据库
	if i.DB != nil {
		if err := i.DB.Close(); err != nil {
			i.Logger.Error("Failed to close DB", zap.Error(err))
			errs = append(errs, err)
		} else {
			i.Logger.Info("Database connection closed")
		}
	}

	// 同步日志
	if i.Logger != nil {
		if err := i.Logger.Sync(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}
	return nil
}
