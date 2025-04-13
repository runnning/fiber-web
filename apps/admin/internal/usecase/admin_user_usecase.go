package usecase

import (
	"context"
	"fiber_web/apps/admin/internal/entity"
	"fiber_web/apps/admin/internal/repository"
	"fiber_web/pkg/query"
)

// AdminUserUseCase 用例接口
type AdminUserUseCase interface {
	CreateAdminUser(ctx context.Context, admin_user *entity.AdminUser) error
	GetAdminUser(ctx context.Context, id uint) (*entity.AdminUser, error)
	UpdateAdminUser(ctx context.Context, admin_user *entity.AdminUser) error
	DeleteAdminUser(ctx context.Context, id uint) error
	List(ctx context.Context, param *query.Query) (*query.PageResult[entity.AdminUser], error)
}

// admin_userUseCase 用例实现
type admin_userUseCase struct {
	admin_userRepo repository.AdminUserRepository
}

// NewAdminUserUseCase 创建用例实例
func NewAdminUserUseCase(admin_userRepo repository.AdminUserRepository) AdminUserUseCase {
	return &admin_userUseCase{
		admin_userRepo: admin_userRepo,
	}
}

func (uc *admin_userUseCase) CreateAdminUser(ctx context.Context, admin_user *entity.AdminUser) error {
	return uc.admin_userRepo.Create(ctx, admin_user)
}

func (uc *admin_userUseCase) GetAdminUser(ctx context.Context, id uint) (*entity.AdminUser, error) {
	return uc.admin_userRepo.FindByID(ctx, id)
}

func (uc *admin_userUseCase) UpdateAdminUser(ctx context.Context, admin_user *entity.AdminUser) error {
	return uc.admin_userRepo.Update(ctx, admin_user)
}

func (uc *admin_userUseCase) DeleteAdminUser(ctx context.Context, id uint) error {
	return uc.admin_userRepo.Delete(ctx, id)
}

func (uc *admin_userUseCase) List(ctx context.Context, param *query.Query) (*query.PageResult[entity.AdminUser], error) {
	return uc.admin_userRepo.List(ctx, param)
}
