package usecase

import (
	"context"
	"fiber_web/apps/admin/internal/entity"
	"fiber_web/apps/admin/internal/repository"
	"fiber_web/pkg/query"
)

// UserUseCase 用户用例接口
type UserUseCase interface {
	CreateUser(ctx context.Context, user *entity.User) error
	GetUser(ctx context.Context, id uint) (*entity.User, error)
	UpdateUser(ctx context.Context, user *entity.User) error
	DeleteUser(ctx context.Context, id uint) error
	List(ctx context.Context, req *query.PageRequest, queryBuilder query.QueryBuilder) (*query.PageResponse[entity.User], error)
}

// userUseCase 用户用例实现
type userUseCase struct {
	userRepo repository.UserRepository
}

// NewUserUseCase 创建用户用例实例
func NewUserUseCase(userRepo repository.UserRepository) UserUseCase {
	return &userUseCase{userRepo: userRepo}
}

func (uc *userUseCase) CreateUser(ctx context.Context, user *entity.User) error {
	return uc.userRepo.Create(ctx, user)
}

func (uc *userUseCase) GetUser(ctx context.Context, id uint) (*entity.User, error) {
	return uc.userRepo.FindByID(ctx, id)
}

func (uc *userUseCase) UpdateUser(ctx context.Context, user *entity.User) error {
	return uc.userRepo.Update(ctx, user)
}

func (uc *userUseCase) DeleteUser(ctx context.Context, id uint) error {
	return uc.userRepo.Delete(ctx, id)
}

func (uc *userUseCase) List(ctx context.Context, req *query.PageRequest, queryBuilder query.QueryBuilder) (*query.PageResponse[entity.User], error) {
	// 处理查询参数和业务逻辑

	// 参数验证
	//if req.Page <= 0 {
	//	req.Page = 1
	//}
	//if req.PageSize <= 0 || req.PageSize > 100 {
	//	req.PageSize = 10 // 限制最大页面大小
	//}

	// 设置默认排序
	if req.OrderBy == "" {
		req.OrderBy = "id"
		req.Order = "DESC"
	}

	// 在这里可以添加业务逻辑相关的处理
	// 例如：权限检查、数据过滤等

	// 对查询构建器进行业务逻辑相关的修改
	if queryBuilder != nil {
		// 例如：根据用户角色添加额外的查询条件
		// 如果当前用户不是管理员，可能需要限制只能查看特定状态的用户
		// queryBuilder.WhereSimple("status", query.OpEq, "active")

		// 或者添加默认的排序条件
		// queryBuilder.OrderBy("created_at", "DESC")

		// 或者添加安全相关的条件，如排除敏感用户
		// queryBuilder.WhereSimple("is_sensitive", query.OpEq, false)
	}

	// 调用仓库层
	return uc.userRepo.List(ctx, req, queryBuilder)
}
