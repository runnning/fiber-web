package initialize

import (
	"context"
	"fiber_web/bootstrap"
	"fiber_web/internal/endpoint"
	"fiber_web/internal/middleware"
	"fiber_web/pkg/auth"
	"fiber_web/pkg/validator"

	"github.com/gofiber/fiber/v2"
)

// Delivery initializes HTTP handlers
type Delivery struct {
	useCase   *UseCase
	infra     *Infrastructure
	app       *fiber.App
	validator *validator.Validator
}

// NewDelivery creates delivery initializer
func NewDelivery(useCase *UseCase, infra *Infrastructure, app *fiber.App) *Delivery {
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
		// Initialize handlers
		endpoint.NewUserHandler(d.useCase.UserUseCase, d.validator)
		// Initialize other handlers

		// Initialize middlewares
		//d.setupMiddlewares()

		// Initialize routes
		//d.app.Use(middleware.CommMiddleware)
		d.setupRoutes()

		return nil
	}, nil)
}

//func (d *Delivery) setupMiddlewares() {
//	// Setup global middlewares
//	d.app.Use(recover.New())
//	d.app.Use(logger.New())
//	d.app.Use(cors.New())
//}

func (d *Delivery) setupRoutes() {
	// Create user handler
	userHandler := endpoint.NewUserHandler(d.useCase.UserUseCase, d.validator)

	// API v1 routes with common middleware
	v1 := d.app.Group("/api/v1", middleware.CommMiddleware()...)

	// Public routes
	v1.Post("/register", userHandler.Register)
	v1.Post("/login", userHandler.Login)
	v1.Get("/test", userHandler.TestUser)
	v1.Get("/users", middleware.Pagination(), userHandler.ListUsers)

	// Protected routes
	v2 := v1.Use(middleware.Jwt(auth.NewJWTManager(d.infra.Config)))
	v2.Get("/users/me", userHandler.GetProfile)
}
