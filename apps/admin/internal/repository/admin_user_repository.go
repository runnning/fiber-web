package repository

import (
	"context"
	"fiber_web/apps/admin/internal/entity"
	"fiber_web/pkg/query"

	"fiber_web/pkg/redis"

	"gorm.io/gorm"
)

type AdminUserRepository interface {
	Create(ctx context.Context, admin_user *entity.AdminUser) error
	FindByID(ctx context.Context, id uint) (*entity.AdminUser, error)
	Update(ctx context.Context, admin_user *entity.AdminUser) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, param *query.Query) (*query.PageResult[entity.AdminUser], error)
}

type admin_userRepository struct {
	db    *gorm.DB
	cache *redis.Client
}

func NewAdminUserRepository(db *gorm.DB, cache *redis.Client) AdminUserRepository {
	return &admin_userRepository{db: db, cache: cache}
}

func (r *admin_userRepository) Create(ctx context.Context, admin_user *entity.AdminUser) error {
	return r.db.WithContext(ctx).Create(admin_user).Error
}

func (r *admin_userRepository) FindByID(ctx context.Context, id uint) (*entity.AdminUser, error) {
	var admin_user entity.AdminUser
	err := r.db.WithContext(ctx).First(&admin_user, id).Error
	if err != nil {
		return nil, err
	}
	return &admin_user, nil
}

func (r *admin_userRepository) Update(ctx context.Context, admin_user *entity.AdminUser) error {
	return r.db.WithContext(ctx).Save(admin_user).Error
}

func (r *admin_userRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.AdminUser{}, id).Error
}

func (r *admin_userRepository) List(ctx context.Context, param *query.Query) (*query.PageResult[entity.AdminUser], error) {
	return query.NewMySQLQuerier[entity.AdminUser](r.db).FindPage(ctx, param)
}
