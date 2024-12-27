package usecase

import (
	"fiber_web/apps/admin/internal/repository"
)

// InitUseCases 初始化所有用例
func InitUseCases(repos *repository.Repositories) *UseCases {
	return &UseCases{
		User: NewUserUseCase(repos.User),
		// 在这里添加其他用例的初始化
	}
}

// UseCases 用例集合
type UseCases struct {
	User UserUseCase
	// 在这里添加其他用例
}
