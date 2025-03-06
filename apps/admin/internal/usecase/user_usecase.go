package usecase

import (
	"context"
	"fiber_web/apps/admin/internal/entity"
	"fiber_web/apps/admin/internal/repository"
	"fiber_web/pkg/query"
	"fmt"
	"time"
)

// UserUseCase 用户用例接口
type UserUseCase interface {
	CreateUser(ctx context.Context, user *entity.User) error
	GetUser(ctx context.Context, id uint) (*entity.User, error)
	UpdateUser(ctx context.Context, user *entity.User) error
	DeleteUser(ctx context.Context, id uint) error
	List(ctx context.Context, req *query.PageRequest) (*query.PageResponse[entity.User], error)
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

func (uc *userUseCase) List(ctx context.Context, req *query.PageRequest) (*query.PageResponse[entity.User], error) {
	// 处理查询参数和业务逻辑

	// 参数验证
	// if req.Page <= 0 {
	// 	req.Page = 1
	// }
	// if req.PageSize <= 0 || req.PageSize > 100 {
	// 	req.PageSize = 10 // 限制最大页面大小
	// }

	// 设置默认排序
	if req.OrderBy == "" {
		req.OrderBy = "id"
		req.Order = "DESC"
	}

	// 处理业务相关的过滤条件
	if status := req.GetFilter("status"); status != "" {
		// 验证状态值是否有效
		validStatus := map[string]bool{"0": true, "1": true, "2": true}
		if !validStatus[status] {
			return nil, fmt.Errorf("无效的状态值: %s", status)
		}
	}

	// 处理时间范围
	startTime := req.GetFilter("start_time")
	endTime := req.GetFilter("end_time")
	if startTime != "" || endTime != "" {
		// 可以在这里添加时间格式验证
		// 例如：验证时间格式是否正确
		if startTime != "" {
			if _, err := time.Parse("2006-01-02", startTime); err != nil {
				return nil, fmt.Errorf("开始时间格式错误: %s", startTime)
			}
		}
		if endTime != "" {
			if _, err := time.Parse("2006-01-02", endTime); err != nil {
				return nil, fmt.Errorf("结束时间格式错误: %s", endTime)
			}
		}
	}

	// 调用仓库层执行查询
	result, err := uc.userRepo.List(ctx, req)
	if err != nil {
		return nil, err
	}

	// 可以在这里对结果进行后处理
	// 例如：敏感信息过滤、数据转换等
	for i := range result.List {
		// 清除敏感信息
		result.List[i].Password = ""
	}

	return result, nil
}
