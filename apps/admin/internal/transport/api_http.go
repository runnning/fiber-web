package transport

import (
	"fiber_web/apps/admin/internal/endpoint"
	"fiber_web/apps/admin/internal/middleware"
	"github.com/gofiber/fiber/v2"
	"time"
)

func RegisterApiHttp(app fiber.Router, handlers *endpoint.Handlers) {
	app.Post("/register", handlers.UserHandler.Register)
	app.Post("/login", handlers.UserHandler.Login)
	app.Post("/refresh-token", middleware.Jwt(), middleware.RateLimit(3, time.Minute), handlers.UserHandler.RefreshToken)
	app.Get("/users", middleware.Pagination(), handlers.UserHandler.List)
	app.Get("/test", middleware.Jwt(), middleware.Pagination(), handlers.UserHandler.TestUser)
	app.Get("/users/me", middleware.Jwt(), handlers.UserHandler.GetProfile)
}
