package middleware

import (
	"errors"
	"fiber_web/pkg/auth"
	"fiber_web/pkg/logger"
	"go.uber.org/zap"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func Jwt() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing authorization header",
			})
		}

		// Check if the header has the Bearer prefix
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid authorization header format",
			})
		}
		claims, err := auth.GetJWTManager().ValidateToken(parts[1])
		if err != nil {
			status := fiber.StatusUnauthorized
			message := "invalid token"

			if errors.Is(err, auth.ErrExpiredToken) {
				message = "token has expired"
			}

			return c.Status(status).JSON(fiber.Map{
				"error": message,
			})
		}

		// Store claims in context for later use
		c.Locals("claims", claims)
		return c.Next()
	}
}

func Rbac() fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, ok := c.Locals("claims").(*auth.Claims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "unauthorized",
			})
		}

		allowed, err := auth.GetEnforcer().HasPermission(claims.Role, c.Path(), c.Method())
		if err != nil {
			logger.Error("Failed to check permission",
				zap.String("role", claims.Role),
				zap.String("path", c.Path()),
				zap.String("method", c.Method()),
				zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to check permission",
			})
		}

		if !allowed {
			logger.Warn("Permission denied",
				zap.String("role", claims.Role),
				zap.String("path", c.Path()),
				zap.String("method", c.Method()))
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "permission denied",
			})
		}
		return c.Next()
	}
}
