package middleware

import (
	"strings"
	"time"
	"user_management_ms/config"
	"user_management_ms/services"

	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		secret := config.Conf.Application.Security.Secret
		issuer := config.Conf.Application.Security.Issuer
		acctm := config.Conf.Application.Security.TokenValidityInSeconds
		reftm := config.Conf.Application.Security.TokenValidityInSecondsForRememberMe

		jwt := services.NewJWTService([]byte(secret), issuer, time.Duration(acctm), time.Duration(reftm))

		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing or invalid token",
			})
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.ParseJWT(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Token parse error",
			})
		}

		if !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		claims, err := jwt.GetClaims(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}
		c.Locals("userId", claims["sub"])

		return c.Next()
	}
}
