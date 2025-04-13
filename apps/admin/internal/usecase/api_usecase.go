package usecase

import (
	"context"
	"fiber_web/apps/admin/internal/entity"
	"fiber_web/apps/admin/internal/repository"
	"fiber_web/pkg/query"
)

// ApiUseCase 用例接口
type ApiUseCase interface {
	CreateApi(ctx context.Context, api *entity.Api) error
	GetApi(ctx context.Context, id uint) (*entity.Api, error)
	UpdateApi(ctx context.Context, api *entity.Api) error
	DeleteApi(ctx context.Context, id uint) error
	List(ctx context.Context, param *query.Query) (*query.PageResult[entity.Api], error)
}

// apiUseCase 用例实现
type apiUseCase struct {
	apiRepo repository.ApiRepository
}

// NewApiUseCase 创建用例实例
func NewApiUseCase(apiRepo repository.ApiRepository) ApiUseCase {
	return &apiUseCase{
		apiRepo: apiRepo,
	}
}

func (uc *apiUseCase) CreateApi(ctx context.Context, api *entity.Api) error {
	return uc.apiRepo.Create(ctx, api)
}

func (uc *apiUseCase) GetApi(ctx context.Context, id uint) (*entity.Api, error) {
	return uc.apiRepo.FindByID(ctx, id)
}

func (uc *apiUseCase) UpdateApi(ctx context.Context, api *entity.Api) error {
	return uc.apiRepo.Update(ctx, api)
}

func (uc *apiUseCase) DeleteApi(ctx context.Context, id uint) error {
	return uc.apiRepo.Delete(ctx, id)
}

func (uc *apiUseCase) List(ctx context.Context, param *query.Query) (*query.PageResult[entity.Api], error) {
	return uc.apiRepo.List(ctx, param)
}
