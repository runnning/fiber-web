package usecase

import (
	"context"
	"fmt"
	"time"
	"fiber_web/admin/internal/entity"
	"fiber_web/admin/internal/repository"
	"fiber_web/admin/pkg/query"
)

// RoleUseCase 用例接口
type RoleUseCase interface {
	CreateRole(ctx context.Context, role *entity.Role) error
	GetRole(ctx context.Context, id uint) (*entity.Role, error)
	UpdateRole(ctx context.Context, role *entity.Role) error
	DeleteRole(ctx context.Context, id uint) error
	List(ctx context.Context, param *query.Query) (*query.PageResult[entity.Role], error)
}

// roleUseCase 用例实现
type roleUseCase struct {
	roleRepo repository.RoleRepository
}

// NewRoleUseCase 创建用例实例
func NewRoleUseCase(roleRepo repository.RoleRepository) RoleUseCase {
	return &roleUseCase{
		roleRepo: roleRepo,
	}
}

func (uc *roleUseCase) CreateRole(ctx context.Context, role *entity.Role) error {
	return uc.roleRepo.Create(ctx, role)
}

func (uc *roleUseCase) GetRole(ctx context.Context, id uint) (*entity.Role, error) {
	return uc.roleRepo.FindByID(ctx, id)
}

func (uc *roleUseCase) UpdateRole(ctx context.Context, role *entity.Role) error {
	return uc.roleRepo.Update(ctx, role)
}

func (uc *roleUseCase) DeleteRole(ctx context.Context, id uint) error {
	return uc.roleRepo.Delete(ctx, id)
}

func (uc *roleUseCase) List(ctx context.Context, param *query.Query) (*query.PageResult[entity.Role], error) {
	return uc.roleRepo.List(ctx, param)
}
