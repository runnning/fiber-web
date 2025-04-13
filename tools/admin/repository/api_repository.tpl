package repository

import (
	"context"
	"fiber_web/admin/internal/entity"
	"fiber_web/admin/pkg/query"

	"fiber_web/admin/pkg/redis"
	"gorm.io/gorm"
)

type ApiRepository interface {
	Create(ctx context.Context, api *entity.Api) error
	FindByID(ctx context.Context, id uint) (*entity.Api, error)
	Update(ctx context.Context, api *entity.Api) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, param *query.Query) (*query.PageResult[entity.Api], error)
}

type apiRepository struct {
	db *gorm.DB
	cache *redis.Client
}

func NewApiRepository(db *gorm.DB, cache *redis.Client) ApiRepository {
	return &apiRepository{db: db, cache: cache}
}

func (r *apiRepository) Create(ctx context.Context, api *entity.Api) error {
	return r.db.WithContext(ctx).Create(api).Error
}

func (r *apiRepository) FindByID(ctx context.Context, id uint) (*entity.Api, error) {
	var api entity.Api
	err := r.db.WithContext(ctx).First(&api, id).Error
	if err != nil {
		return nil, err
	}
	return &api, nil
}

func (r *apiRepository) Update(ctx context.Context, api *entity.Api) error {
	return r.db.WithContext(ctx).Save(api).Error
}

func (r *apiRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.Api{}, id).Error
}

func (r *apiRepository) List(ctx context.Context, param *query.Query) (*query.PageResult[entity.Api], error) {
	return query.NewMySQLQuerier[entity.Api](r.db).FindPage(ctx, param)
}
