package repository

import (
	"context"
	"fiber_web/apps/admin/internal/entity"
	"fiber_web/pkg/query"

	"fiber_web/pkg/redis"
	"gorm.io/gorm"
)

// UserRepository 用户仓库接口
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	FindByID(ctx context.Context, id uint) (*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, req *query.PageRequest, queryBuilder query.QueryBuilder) (*query.PageResponse[entity.User], error)
}

// userRepository 用户仓库实现
type userRepository struct {
	db    *gorm.DB
	cache *redis.Client
}

// NewUserRepository 创建用户仓库
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

func (r *userRepository) List(ctx context.Context, req *query.PageRequest, queryBuilder query.QueryBuilder) (*query.PageResponse[entity.User], error) {
	var users []entity.User

	// 使用传入的查询构建器
	// 如果查询构建器为空，创建一个新的
	if queryBuilder == nil {
		// 创建一个新的查询构建器
		factory := query.NewMySQLQueryFactory(r.db)
		queryBuilder = factory.NewQuery()

		// 设置模型
		db := r.db.WithContext(ctx).Model(&entity.User{})
		queryBuilder.WhereRaw(db)
	}

	// 创建数据提供者
	provider := query.NewMySQLProvider[entity.User](r.db)

	// 执行分页查询
	return query.Paginate(ctx, queryBuilder, provider, req, &users)
}
