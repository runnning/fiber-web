package usecase

import (
	"context"
	"fmt"
	"time"
	"fiber_web/admin/internal/entity"
	"fiber_web/admin/internal/repository"
	"fiber_web/admin/pkg/query"
)

// MenuUseCase 用例接口
type MenuUseCase interface {
	CreateMenu(ctx context.Context, menu *entity.Menu) error
	GetMenu(ctx context.Context, id uint) (*entity.Menu, error)
	UpdateMenu(ctx context.Context, menu *entity.Menu) error
	DeleteMenu(ctx context.Context, id uint) error
	List(ctx context.Context, param *query.Query) (*query.PageResult[entity.Menu], error)
}

// menuUseCase 用例实现
type menuUseCase struct {
	menuRepo repository.MenuRepository
}

// NewMenuUseCase 创建用例实例
func NewMenuUseCase(menuRepo repository.MenuRepository) MenuUseCase {
	return &menuUseCase{
		menuRepo: menuRepo,
	}
}

func (uc *menuUseCase) CreateMenu(ctx context.Context, menu *entity.Menu) error {
	return uc.menuRepo.Create(ctx, menu)
}

func (uc *menuUseCase) GetMenu(ctx context.Context, id uint) (*entity.Menu, error) {
	return uc.menuRepo.FindByID(ctx, id)
}

func (uc *menuUseCase) UpdateMenu(ctx context.Context, menu *entity.Menu) error {
	return uc.menuRepo.Update(ctx, menu)
}

func (uc *menuUseCase) DeleteMenu(ctx context.Context, id uint) error {
	return uc.menuRepo.Delete(ctx, id)
}

func (uc *menuUseCase) List(ctx context.Context, param *query.Query) (*query.PageResult[entity.Menu], error) {
	return uc.menuRepo.List(ctx, param)
}
