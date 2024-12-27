package transport

import (
	"fiber_web/apps/admin/internal/endpoint"
	"fiber_web/apps/admin/internal/middleware"
	"fiber_web/pkg/auth"
	"fiber_web/pkg/config"

	"github.com/gofiber/fiber/v2"
)

func RegisterApiHttp(app fiber.Router, handlers *endpoint.Handlers) {
	app.Post("/register", handlers.User.Register)
	app.Post("/login", handlers.User.Login)
	app.Get("/users", middleware.Pagination(), handlers.User.ListUsers)
	app.Get("/test", middleware.Pagination(), handlers.User.TestUser)
	app.Get("/users/me", middleware.Jwt(auth.NewJWTManager(&config.Data.JWT)), handlers.User.GetProfile)
}
