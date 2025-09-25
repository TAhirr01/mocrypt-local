package middleware

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Status messages
var statusMessages = map[int]string{
	200: "Ok",
	201: "Created",
	400: "Bad Request",
	403: "Forbidden",
	404: "Not Found",
	401: "Unauthorized",
	429: "Too many requests",
	500: "Internal Server Error",
}

// Map HTTP status codes to zap log levels
var statusToLevel = map[int]zapcore.Level{
	200: zap.InfoLevel,
	201: zap.InfoLevel,
	400: zap.WarnLevel,
	403: zap.WarnLevel,
	401: zap.WarnLevel,
	404: zap.ErrorLevel,
	429: zap.InfoLevel,
	500: zap.ErrorLevel,
}

func LoggingMiddleware(logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()
		duration := time.Since(start)
		statusCode := c.Response().StatusCode()
		responseBody := c.Response().Body()

		responseErr := struct {
			ResponseErr string `json:"error"`
		}{}
		if jsonErr := json.Unmarshal(responseBody, &responseErr); jsonErr != nil {
			responseErr.ResponseErr = ""
		}

		fields := []zap.Field{
			zap.String("err", responseErr.ResponseErr),
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", statusCode),
			zap.Duration("duration", duration),
		}

		level, ok := statusToLevel[statusCode]
		if !ok {
			level = zap.InfoLevel
		}

		message, ok := statusMessages[statusCode]
		if !ok {
			message = fmt.Sprintf("Unknown status %d", statusCode)
		}

		switch level {
		case zap.DebugLevel:
			logger.Debug(message, fields...)
		case zap.InfoLevel:
			logger.Info(message, fields...)
		case zap.WarnLevel:
			logger.Warn(message, fields...)
		case zap.ErrorLevel:
			logger.Error(message, fields...)
		case zap.DPanicLevel:
			logger.DPanic(message, fields...)
		case zap.PanicLevel:
			logger.Panic(message, fields...)
		case zap.FatalLevel:
			logger.Fatal(message, fields...)
		}

		return err
	}
}
