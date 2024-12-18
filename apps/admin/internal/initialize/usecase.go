package initialize

import (
	"context"
	"fiber_web/apps/admin/internal/bootstrap"
	"fiber_web/apps/admin/internal/usecase"
)

// UseCase initializes use cases
type UseCase struct {
	repo        *Repository
	UserUseCase usecase.UserUseCase
	// Add other use cases here
}

// NewUseCase creates use case initializer
func NewUseCase(repo *Repository) *UseCase {
	return &UseCase{
		repo: repo,
	}
}

// Register registers use case initialization
func (u *UseCase) Register(b *bootstrap.Bootstrapper) {
	b.Register(func(ctx context.Context) error {
		u.UserUseCase = usecase.NewUserUseCase(u.repo.UserRepo)
		// Initialize other use cases
		return nil
	}, nil)
}
