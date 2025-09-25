package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// GlobalRateLimiter returns a pre-configured limiter middleware
func GlobalRateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        10,               // 10 requests
		Expiration: 30 * time.Second, // per 30s window
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Too many requests, slow down.",
			})
		},
	})
}

// RouteRateLimiter allows you to set custom limits per route
func RouteRateLimiter(max int, window time.Duration) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        max,
		Expiration: window,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Rate limit exceeded",
			})
		},
	})
}
