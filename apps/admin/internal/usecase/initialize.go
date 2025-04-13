package usecase

import (
	"fiber_web/apps/admin/internal/repository"
)

// InitUseCases 初始化所有用例
func InitUseCases(repos *repository.Repositories) *UseCases {
	return &UseCases{
		UserUserCase: NewUserUseCase(repos.UserRepository),
		// 在这里添加其他用例的初始化
		AdminUserUseCase: NewAdminUserUseCase(repos.AdminUserRepository),
		ApiUseCase:       NewApiUseCase(repos.ApiRepository),
		MenuUseCase:      NewMenuUseCase(repos.MenuRepository),
		RoleUseCase:      NewRoleUseCase(repos.RoleRepository),
	}
}

// UseCases 用例集合
type UseCases struct {
	UserUserCase UserUseCase
	// 在这里添加其他用例
	AdminUserUseCase AdminUserUseCase
	ApiUseCase       ApiUseCase
	MenuUseCase      MenuUseCase
	RoleUseCase      RoleUseCase
}
