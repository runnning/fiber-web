package repository

import (
	"context"
	"fiber_web/apps/admin/internal/entity"
	"fiber_web/pkg/query"

	"fiber_web/pkg/redis"

	"gorm.io/gorm"
)

type MenuRepository interface {
	Create(ctx context.Context, menu *entity.Menu) error
	FindByID(ctx context.Context, id uint) (*entity.Menu, error)
	Update(ctx context.Context, menu *entity.Menu) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, param *query.Query) (*query.PageResult[entity.Menu], error)
}

type menuRepository struct {
	db    *gorm.DB
	cache *redis.Client
}

func NewMenuRepository(db *gorm.DB, cache *redis.Client) MenuRepository {
	return &menuRepository{db: db, cache: cache}
}

func (r *menuRepository) Create(ctx context.Context, menu *entity.Menu) error {
	return r.db.WithContext(ctx).Create(menu).Error
}

func (r *menuRepository) FindByID(ctx context.Context, id uint) (*entity.Menu, error) {
	var menu entity.Menu
	err := r.db.WithContext(ctx).First(&menu, id).Error
	if err != nil {
		return nil, err
	}
	return &menu, nil
}

func (r *menuRepository) Update(ctx context.Context, menu *entity.Menu) error {
	return r.db.WithContext(ctx).Save(menu).Error
}

func (r *menuRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.Menu{}, id).Error
}

func (r *menuRepository) List(ctx context.Context, param *query.Query) (*query.PageResult[entity.Menu], error) {
	return query.NewMySQLQuerier[entity.Menu](r.db).FindPage(ctx, param)
}
