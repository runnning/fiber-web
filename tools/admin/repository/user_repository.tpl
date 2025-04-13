package repository

import (
	"context"
	"fiber_web/admin/internal/entity"
	"fiber_web/admin/pkg/query"

	"fiber_web/admin/pkg/redis"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	FindByID(ctx context.Context, id uint) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, param *query.Query) (*query.PageResult[entity.User], error)
}

type userRepository struct {
	db *gorm.DB
	cache *redis.Client
}

func NewUserRepository(db *gorm.DB, cache *redis.Client) UserRepository {
	return &userRepository{db: db, cache: cache}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) FindByID(ctx context.Context, id uint) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.User{}, id).Error
}

func (r *userRepository) List(ctx context.Context, param *query.Query) (*query.PageResult[entity.User], error) {
	return query.NewMySQLQuerier[entity.User](r.db).FindPage(ctx, param)
}
