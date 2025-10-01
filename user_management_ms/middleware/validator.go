package middleware

import (
	"regexp"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var Validate *validator.Validate

// InitValidator initializes validator and custom rules
func InitValidator() {
	Validate = validator.New()

	// Custom phone validation
	Validate.RegisterValidation("phone", func(fl validator.FieldLevel) bool {
		re := regexp.MustCompile(`^\+?[0-9]{10,15}$`)
		return re.MatchString(fl.Field().String())
	})

	Validate.RegisterValidation("password", func(fl validator.FieldLevel) bool {
		password := fl.Field().String()
		if len(password) < 5 {
			return false
		}

		// Check for at least one uppercase letter
		hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
		// Check for at least one lowercase letter
		hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
		// Check for at least one digit
		hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
		// Check for at least one symbol
		hasSymbol := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>/?]`).MatchString(password)

		return hasUpper && hasLower && hasDigit && hasSymbol
	})
	Validate.RegisterValidation("pin", func(fl validator.FieldLevel) bool {
		pin := fl.Field().String()
		if len(pin) != 6 {
			return false
		}
		return true
	})
}
func translateValidationErrors(err validator.ValidationErrors) map[string]string {
	errorsMap := make(map[string]string)
	for _, e := range err {
		field := e.Field()
		tag := e.Tag()
		switch tag {
		case "pin":
			errorsMap[field] = field + "pin must be 6 digits long"
		case "required":
			errorsMap[field] = field + " is required"
		case "email":
			errorsMap[field] = field + " must be a valid email"
		case "phone":
			errorsMap[field] = field + " must be a valid phone number"
		case "password":
			errorsMap[field] = field + " must be at least 5 characters, with 1 uppercase, 1 number, and 1 symbol"
		default:
			errorsMap[field] = field + " is invalid"
		}
	}
	return errorsMap
}

// ValidateBody is Fiber middleware that validates request body
func ValidateBody[T any]() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body T

		// Parse JSON into struct
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid request body",
			})
		}

		// Validate struct
		if err := Validate.Struct(&body); err != nil {
			if errs, ok := err.(validator.ValidationErrors); ok {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"errors": translateValidationErrors(errs),
				})
			}
			// fallback for unexpected errors
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Store validated body in context for controller
		c.Locals("body", &body)
		return c.Next()
	}
}
