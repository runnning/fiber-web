package api

import (
	"context"
	"fiber_web/apps/admin/internal/bootstrap"
	"fiber_web/apps/admin/internal/endpoint"
	"fiber_web/apps/admin/internal/initialize"
	"fiber_web/apps/admin/internal/middleware"
	"fiber_web/pkg/auth"
	"fiber_web/pkg/validator"

	"github.com/gofiber/fiber/v2"
)

// Delivery initializes HTTP handlers
type Delivery struct {
	useCase   *initialize.UseCase
	infra     *initialize.Infrastructure
	app       *fiber.App
	validator *validator.Validator
}

// NewDelivery creates delivery initializer
func NewDelivery(useCase *initialize.UseCase, infra *initialize.Infrastructure, app *fiber.App) *Delivery {
	return &Delivery{
		useCase:   useCase,
		infra:     infra,
		app:       app,
		validator: validator.New(&validator.Config{Language: infra.Config.App.Language}),
	}
}

// Register registers delivery initialization
func (d *Delivery) Register(b *bootstrap.Bootstrapper) {
	b.Register(func(ctx context.Context) error {
		// Create user handler
		userHandler := endpoint.NewUserHandler(d.useCase.UserUseCase, d.validator)

		// API v1 routes with common middleware
		v1 := d.app.Group("/api/v1", middleware.CommMiddleware(d.infra.Config.App.Env)...)
		{
			// Public routes
			v1.Post("/register", userHandler.Register)
			v1.Post("/login", userHandler.Login)
			v1.Get("/test", userHandler.TestUser)
			v1.Get("/users", middleware.Pagination(), userHandler.ListUsers)

			// Protected routes
			v1.Get("/users/me", middleware.Jwt(auth.NewJWTManager(d.infra.Config)), userHandler.GetProfile)
		}
		return nil
	}, nil)
}
