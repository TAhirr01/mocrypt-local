package middleware

import (
	"log"
	"runtime/debug"

	"github.com/gofiber/fiber/v2"
)

func RecoveryMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		defer func() {
			if err := recover(); err != nil {
				msg := "Caught panic: %v, Stack Trace: %s"
				log.Printf(msg, err, string(debug.Stack()))

				er := fiber.StatusInternalServerError
				c.Status(er)
			}
		}()
		return c.Next()
	}
}
