package repository

import (
	"context"
	"fiber_web/admin/internal/entity"
	"fiber_web/admin/pkg/query"

	"fiber_web/admin/pkg/redis"
	"gorm.io/gorm"
)

type RoleRepository interface {
	Create(ctx context.Context, role *entity.Role) error
	FindByID(ctx context.Context, id uint) (*entity.Role, error)
	Update(ctx context.Context, role *entity.Role) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, param *query.Query) (*query.PageResult[entity.Role], error)
}

type roleRepository struct {
	db *gorm.DB
	cache *redis.Client
}

func NewRoleRepository(db *gorm.DB, cache *redis.Client) RoleRepository {
	return &roleRepository{db: db, cache: cache}
}

func (r *roleRepository) Create(ctx context.Context, role *entity.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *roleRepository) FindByID(ctx context.Context, id uint) (*entity.Role, error) {
	var role entity.Role
	err := r.db.WithContext(ctx).First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) Update(ctx context.Context, role *entity.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *roleRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.Role{}, id).Error
}

func (r *roleRepository) List(ctx context.Context, param *query.Query) (*query.PageResult[entity.Role], error) {
	return query.NewMySQLQuerier[entity.Role](r.db).FindPage(ctx, param)
}
