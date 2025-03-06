package repository

import (
	"context"
	"fiber_web/apps/admin/internal/entity"
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
	List(ctx context.Context, req *query.PageRequest) (*query.PageResponse[entity.User], error)
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

func (r *userRepository) List(ctx context.Context, req *query.PageRequest) (*query.PageResponse[entity.User], error) {
	var users []entity.User
	db := r.db.WithContext(ctx).Model(&entity.User{})

	// 处理搜索条件
	if search := req.GetFilter("search"); search != "" {
		db = query.BuildSearchQuery(db, search, []string{"name", "email"})
	}

	// 处理状态过滤
	if status := req.GetFilter("status"); status != "" {
		db = db.Where("status = ?", status)
	}

	// 处理时间范围
	startTime := req.GetFilter("start_time")
	endTime := req.GetFilter("end_time")
	if startTime != "" || endTime != "" {
		db = query.BuildTimeRangeQuery(db, "created_at", startTime, endTime)
	}

	// 构建查询
	builder := query.NewMySQLQuery(db)

	// 添加其他条件
	if role := req.GetFilter("role"); role != "" {
		builder.AddCondition("role", query.OpEq, role)
	}

	// 创建数据提供者
	provider := query.NewMySQLProvider[entity.User](r.db)

	// 执行分页查询
	return query.Paginate(ctx, builder, provider, req, &users)
}
