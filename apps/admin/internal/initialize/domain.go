package initialize

import (
	"context"
	"fiber_web/apps/admin/internal/repository"
	"fiber_web/apps/admin/internal/usecase"
	"sync"
)

// Domain 业务领域
type Domain struct {
	infra *Infra
	Repos *repository.Repositories
	Uses  *usecase.UseCases
	mu    sync.RWMutex
}

// Init 实现 Component 接口
func (d *Domain) Init(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// 初始化仓储层
	defaultDB, err := d.infra.DB.GetDB("default")
	if err != nil {
		return err
	}
	defaultRedis, err := d.infra.Redis.GetClient("default")
	if err != nil {
		return err
	}
	d.Repos = repository.InitRepositories(defaultDB.DB(), defaultRedis)
	d.infra.Logger.Info("Repositories initialized")

	// 初始化用例层
	d.Uses = usecase.InitUseCases(d.Repos)
	d.infra.Logger.Info("UseCases initialized")

	return nil
}

// Start 实现 Component 接口
func (d *Domain) Start(ctx context.Context) error {
	return nil
}

// Stop 实现 Component 接口
func (d *Domain) Stop(ctx context.Context) error {
	return nil
}

func NewDomain(infra *Infra) *Domain {
	return &Domain{
		infra: infra,
	}
}
