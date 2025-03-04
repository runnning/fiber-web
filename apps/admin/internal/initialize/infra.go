package initialize

import (
	"context"
	"fiber_web/pkg/auth"
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
	//Config *config.Config
	DB              *database.DBManager
	Redis           *redis.RedisManager
	MongoDB         *database.MongoManager
	DefaultProducer *queue.Producer
	Logger          *logger.Logger
	Cron            *cron.Scheduler
	mu              sync.RWMutex
}

func NewInfra() *Infra {
	return &Infra{}
}

// Init 实现 Component 接口
func (i *Infra) Init(ctx context.Context) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	// 初始化日志
	if err := logger.InitLogger(&config.Data.Log, logger.WithAsync(3, 1024)); err != nil {
		return err
	}
	i.Logger = logger.GetLogger()
	i.Logger.Info("Logger initialized")

	// 初始化数据库
	dbManager, err := database.NewDBManager(&config.Data.Database)
	if err != nil {
		return err
	}
	i.DB = dbManager
	i.Logger.Info("Database initialized")

	// 初始化MongoDB
	mongoManager, err := database.NewMongoManager(&config.Data.MongoDB)
	if err != nil {
		return err
	}
	i.MongoDB = mongoManager
	i.Logger.Info("MongoDB initialized")

	// 初始化 Redis
	redisManager, err := redis.NewRedisManager(&config.Data.Redis)
	if err != nil {
		return err
	}
	i.Redis = redisManager
	i.Logger.Info("Redis initialized")

	// 初始化 NSQ
	defaultProducer, err := queue.NewProducer(&config.Data.NSQ, &queue.DefaultOptions)
	if err != nil {
		return err
	}
	i.DefaultProducer = defaultProducer
	i.Logger.Info("NSQ initialized")

	// 初始化权限
	defaultDB, err := i.DB.GetDB("default")
	if err != nil {
		return err
	}
	if err = auth.InitRbac(defaultDB.DB()); err != nil {
		return err
	}
	i.Logger.Info("RBAC initialized")

	// 初始化jwt
	auth.InitJWTManager(&config.Data.JWT)
	i.Logger.Info("jwt initialized")

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

	i.Logger.Info("jwt stopped")

	// 关闭 NSQ
	if i.DefaultProducer != nil {
		i.DefaultProducer.Stop()
		i.Logger.Info("NSQ producer stopped")
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

	// 关闭MongoDB
	if i.MongoDB != nil {
		if err := i.MongoDB.Close(); err != nil {
			i.Logger.Error("Failed to close MongoDB", zap.Error(err))
			errs = append(errs, err)
		} else {
			i.Logger.Info("MongoDB connection closed")
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
