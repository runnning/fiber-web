package endpoint

import (
	"fiber_web/apps/admin/internal/endpoint/user"
	"fiber_web/apps/admin/internal/usecase"
	"fiber_web/pkg/validator"
)

// Handlers 集中管理所有的HTTP处理器
type Handlers struct {
	User *user.UserHandler
}

// InitHandlers 初始化所有处理器
func InitHandlers(User usecase.UserUseCase, validator *validator.Validator) *Handlers {
	return &Handlers{
		User: user.NewUserHandler(User, validator),
	}
}
