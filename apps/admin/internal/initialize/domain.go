package initialize

import (
	"context"
	"fiber_web/apps/admin/internal/repository"
	"fiber_web/apps/admin/internal/usecase"
	"fiber_web/pkg/auth"
	"sync"
)

// Domain 业务领域
type Domain struct {
	infra *Infra
	Repos *Repositories
	Uses  *UseCases
	mu    sync.RWMutex
}

// Repositories 仓储层集合
type Repositories struct {
	User repository.UserRepository
	// 添加其他仓储...
}

// UseCases 用例层集合
type UseCases struct {
	User usecase.UserUseCase
	// 添加其他用例...
}

// Init 实现 Component 接口
func (d *Domain) Init(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// 初始化仓储层
	d.Repos.User = repository.NewUserRepository(d.infra.DB.DB(), d.infra.Redis)
	d.infra.Logger.Info("Repositories initialized")

	// 初始化用例层
	d.Uses.User = usecase.NewUserUseCase(d.Repos.User)
	d.infra.Logger.Info("UseCases initialized")

	// 初始化权限
	if _, err := auth.InitRbac(d.infra.DB.DB()); err != nil {
		return err
	}
	d.infra.Logger.Info("RBAC initialized")

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
		Repos: &Repositories{},
		Uses:  &UseCases{},
	}
}
