package middleware

import (
	"errors"
	"fiber_web/pkg/auth"
	"fiber_web/pkg/logger"
	"fiber_web/pkg/response"
	"fiber_web/pkg/utils/errorx"

	"strings"

	"github.com/gofiber/fiber/v2"
)

func Jwt() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return response.Unauthorized(c, "missing authorization header")
		}

		// Check if the header has the Bearer prefix
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return response.Unauthorized(c, "invalid authorization header format")
		}

		claims, err := auth.GetJWTManager().ValidateToken(parts[1])
		if err != nil {
			message := "invalid token"

			if errors.Is(err, auth.ErrExpiredToken) {
				message = "token has expired"
			}

			return response.Unauthorized(c, message)
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
			return response.Unauthorized(c, "unauthorized")
		}

		allowed, err := auth.GetEnforcer().HasPermission(claims.Role, c.Path(), c.Method())
		if err != nil {
			logger.ErrorLog("Failed to check permission",
				logger.String("role", claims.Role),
				logger.String("path", c.Path()),
				logger.String("method", c.Method()),
				logger.ErrorField(err))
			return response.ServerError(c, errorx.NewSystemError("failed to check permission"))
		}

		if !allowed {
			logger.Warn("Permission denied",
				logger.String("role", claims.Role),
				logger.String("path", c.Path()),
				logger.String("method", c.Method()))
			return response.Forbidden(c, "permission denied")
		}
		return c.Next()
	}
}
