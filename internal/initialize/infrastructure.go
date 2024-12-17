package initialize

import (
	"context"
	"fiber_web/bootstrap"
	"fiber_web/pkg/auth"
	"fiber_web/pkg/config"
	"fiber_web/pkg/cron"
	"fiber_web/pkg/database"
	"fiber_web/pkg/logger"
	"fiber_web/pkg/queue"
	"fiber_web/pkg/redis"
)

// Infrastructure initializes infrastructure components
type Infrastructure struct {
	Config *config.Config
	DB     *database.Database
	Redis  *redis.Client
	NSQ    *queue.Producer
	Logger *logger.Logger
	Cron   *cron.Scheduler
}

// NewInfrastructure creates infrastructure initializer
func NewInfrastructure(cfg *config.Config) *Infrastructure {
	return &Infrastructure{
		Config: cfg,
	}
}

// Register registers infrastructure initialization
func (i *Infrastructure) Register(b *bootstrap.Bootstrapper) {
	// Logger initialization
	b.Register(func(ctx context.Context) error {
		if err := logger.InitLogger(i.Config.App.Env); err != nil {
			return err
		}
		i.Logger = logger.GetDefaultLogger()
		return nil
	}, func() error {
		// Sync logger
		return i.Logger.Sync()
	})

	// Database initialization
	b.Register(func(ctx context.Context) error {
		db, err := database.NewMySQL(i.Config)
		if err != nil {
			return err
		}
		i.DB = db
		return nil
	}, func() error {
		if i.DB != nil {
			return i.DB.Close()
		}
		return nil
	})

	// Redis initialization
	b.Register(func(ctx context.Context) error {
		client, err := redis.NewClient(i.Config)
		if err != nil {
			return err
		}
		i.Redis = client
		return nil
	}, func() error {
		if i.Redis != nil {
			return i.Redis.Close()
		}
		return nil
	})

	// NSQ initialization
	b.Register(func(ctx context.Context) error {
		producer, err := queue.NewProducer(i.Config)
		if err != nil {
			return err
		}
		i.NSQ = producer
		return nil
	}, func() error {
		if i.NSQ != nil {
			return i.NSQ.Stop()
		}
		return nil
	})

	b.Register(func(ctx context.Context) error {

		// Initialize cron scheduler
		if err := InitScheduler(logger.GetLogger()); err != nil {
			return err
		}
		i.Cron = GetScheduler()

		return nil
	}, func() error {
		// Stop cron scheduler
		StopScheduler()
		// Sync logger
		return i.Logger.Sync()
	})

	// rbac
	b.Register(func(ctx context.Context) error {
		_, err := auth.InitRbac(i.DB.DB())
		if err != nil {
			return err
		}
		return nil
	}, func() error {
		return i.Logger.Sync()
	})

}
