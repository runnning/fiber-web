package endpoint

import (
	"fiber_web/apps/admin/internal/usecase"
	"fiber_web/pkg/validator"
)

// InitHandlers 初始化所有处理器
func InitHandlers(
	uses *usecase.UseCases,
	validator *validator.Validator) *Handlers {
	return &Handlers{
		UserHandler:      NewUserHandler(uses.UserUserCase, validator),
		AdminUserHandler: NewAdminUserHandler(uses.AdminUserUseCase, validator),
		ApiHandler:       NewApiHandler(uses.ApiUseCase, validator),
		MenuHandler:      NewMenuHandler(uses.MenuUseCase, validator),
		RoleHandler:      NewRoleHandler(uses.RoleUseCase, validator),
	}
}

// Handlers 集中管理所有的HTTP处理器
type Handlers struct {
	UserHandler      *UserHandler
	AdminUserHandler *AdminUserHandler
	ApiHandler       *ApiHandler
	MenuHandler      *MenuHandler
	RoleHandler      *RoleHandler
}
