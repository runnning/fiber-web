package usecase

import (
	"context"
	"fiber_web/apps/admin/internal/entity"
	"fiber_web/apps/admin/internal/repository"
	"fiber_web/pkg/query"
)

// UserUseCase 用户用例接口
type UserUseCase interface {
	CreateUser(ctx context.Context, user *entity.User) error
	GetUser(ctx context.Context, id uint) (*entity.User, error)
	UpdateUser(ctx context.Context, user *entity.User) error
	DeleteUser(ctx context.Context, id uint) error
	List(ctx context.Context, opts ...query.Option) (*query.Result[[]entity.User], error)
}

// userUseCase 用户用例实现
type userUseCase struct {
	userRepo repository.UserRepository
}

// NewUserUseCase 创建用户用例实例
func NewUserUseCase(userRepo repository.UserRepository) UserUseCase {
	return &userUseCase{userRepo: userRepo}
}

func (uc *userUseCase) CreateUser(ctx context.Context, user *entity.User) error {
	return uc.userRepo.Create(ctx, user)
}

func (uc *userUseCase) GetUser(ctx context.Context, id uint) (*entity.User, error) {
	return uc.userRepo.FindByID(ctx, id)
}

func (uc *userUseCase) UpdateUser(ctx context.Context, user *entity.User) error {
	return uc.userRepo.Update(ctx, user)
}

func (uc *userUseCase) DeleteUser(ctx context.Context, id uint) error {
	return uc.userRepo.Delete(ctx, id)
}

func (uc *userUseCase) List(ctx context.Context, opts ...query.Option) (*query.Result[[]entity.User], error) {
	return uc.userRepo.List(ctx, opts...)
}
