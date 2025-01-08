package transport

import (
	"fiber_web/apps/admin/internal/endpoint"
	"fiber_web/apps/admin/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

func RegisterApiHttp(app fiber.Router, handlers *endpoint.Handlers) {
	app.Post("/register", handlers.User.Register)
	app.Post("/login", handlers.User.Login)
	app.Post("/refresh-token", middleware.Jwt(), handlers.User.RefreshToken)
	app.Get("/users", middleware.Jwt(), middleware.Pagination(), handlers.User.ListUsers)
	app.Get("/test", middleware.Jwt(), middleware.Pagination(), handlers.User.TestUser)
	app.Get("/users/me", middleware.Jwt(), handlers.User.GetProfile)
}
