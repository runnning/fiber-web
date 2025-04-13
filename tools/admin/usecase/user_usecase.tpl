package usecase

import (
	"context"
	"fmt"
	"time"
	"fiber_web/admin/internal/entity"
	"fiber_web/admin/internal/repository"
	"fiber_web/admin/pkg/query"
)

// UserUseCase 用例接口
type UserUseCase interface {
	CreateUser(ctx context.Context, user *entity.User) error
	GetUser(ctx context.Context, id uint) (*entity.User, error)
	UpdateUser(ctx context.Context, user *entity.User) error
	DeleteUser(ctx context.Context, id uint) error
	List(ctx context.Context, param *query.Query) (*query.PageResult[entity.User], error)
}

// userUseCase 用例实现
type userUseCase struct {
	userRepo repository.UserRepository
}

// NewUserUseCase 创建用例实例
func NewUserUseCase(userRepo repository.UserRepository) UserUseCase {
	return &userUseCase{
		userRepo: userRepo,
	}
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

func (uc *userUseCase) List(ctx context.Context, param *query.Query) (*query.PageResult[entity.User], error) {
	return uc.userRepo.List(ctx, param)
}
