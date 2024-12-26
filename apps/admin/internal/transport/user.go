package transport

import (
	"fiber_web/apps/admin/internal/endpoint"
	"fiber_web/apps/admin/internal/middleware"
	"fiber_web/pkg/auth"
	"fiber_web/pkg/config"

	"github.com/gofiber/fiber/v2"
)

func RegisterUserRoutes(app fiber.Router, handler *endpoint.UserHandler, config *config.Config) {
	app.Post("/register", handler.Register)
	app.Post("/login", handler.Login)
	app.Get("/users", middleware.Pagination(), handler.ListUsers)
	app.Get("/test", middleware.Pagination(), handler.TestUser)
	app.Get("/users/me", middleware.Jwt(auth.NewJWTManager(config)), handler.GetProfile)
}
