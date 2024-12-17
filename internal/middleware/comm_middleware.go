package middleware

import (
	"github.com/gofiber/fiber/v2"
)

func CommMiddleware(app *fiber.App) {
	// 添加请求ID中间件
	app.Use(RequestID())

	// Recover from panics
	app.Use(Recovery())

	// CORS
	app.Use(CORS())

	// Logger
	app.Use(Logger())

	app.Use(NotFound())
}
