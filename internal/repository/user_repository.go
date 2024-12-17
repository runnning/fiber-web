package repository

import (
	"context"
	"fiber_web/internal/entity"
	"fiber_web/pkg/query"
	"fiber_web/pkg/redis"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	FindByID(ctx context.Context, id uint) (*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, opts ...query.Option) (*query.Result[[]entity.User], error)
}

type userRepository struct {
	db    *gorm.DB
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

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
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

func (r *userRepository) List(ctx context.Context, opts ...query.Option) (*query.Result[[]entity.User], error) {
	var users []entity.User
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.User{})
	db = query.WithOptions(db, opts...)

	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	if err := db.Find(&users).Error; err != nil {
		return nil, err
	}

	var page, pageSize int
	for _, opt := range opts {
		if po, ok := opt.(query.PageOption); ok {
			page = po.Page
			pageSize = po.PageSize
			break
		}
	}

	return query.NewResult(users, total, page, pageSize), nil
}
