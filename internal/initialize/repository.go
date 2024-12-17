package initialize

import (
	"context"
	"fiber_web/bootstrap"
	"fiber_web/internal/repository"
)

// Repository initializes repositories
type Repository struct {
	infra    *Infrastructure
	UserRepo repository.UserRepository
	// Add other repositories here
}

// NewRepository creates repository initializer
func NewRepository(infra *Infrastructure) *Repository {
	return &Repository{
		infra: infra,
	}
}

// Register registers repository initialization
func (r *Repository) Register(b *bootstrap.Bootstrapper) {
	b.Register(func(ctx context.Context) error {
		r.UserRepo = repository.NewUserRepository(r.infra.DB.DB(), r.infra.Redis)
		// Initialize other repositories
		return nil
	}, nil)
}
