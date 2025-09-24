package controller

import (
	"net/http"
	"user_management_ms/dtos/request"

	"user_management_ms/services"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

type IPasskeyController interface {
	RegisterStart(c *fiber.Ctx) error
	RegisterFinish(c *fiber.Ctx) error
	LoginStart(c *fiber.Ctx) error
	LoginFinish(c *fiber.Ctx) error
}

type PasskeyController struct {
	service services.IPasskeyService
}

func NewPasskeyController(service services.IPasskeyService) IPasskeyController {
	return &PasskeyController{service: service}
}

func (pc *PasskeyController) RegisterStart(c *fiber.Ctx) error {
	userId := c.Locals("userId")
	options, err := pc.service.RegisterStart(&request.StartPasskeyRegistrationRequest{UserId: uint(userId.(float64))})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(options)
}

func (pc *PasskeyController) RegisterFinish(c *fiber.Ctx) error {
	// 1. Parse user ID
	userId := c.Locals("userId")
	// 3. Convert Fiber (fasthttp) request to *http.Request
	req := new(http.Request)
	if err := fasthttpadaptor.ConvertRequest(c.Context(), req, true); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to convert request"})
	}

	// 4. Call service to finish registration
	if err := pc.service.RegisterFinish(uint(userId.(float64)), req); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{})
}

func (pc *PasskeyController) LoginStart(c *fiber.Ctx) error {
	options, sessionId, err := pc.service.LoginStart()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"sessionId": sessionId,
		"options":   options,
	})
}

func (pc *PasskeyController) LoginFinish(c *fiber.Ctx) error {
	sessionId := c.Params("sessionId")
	// Convert Fiber fasthttp request → *http.Request
	req := new(http.Request)
	if err := fasthttpadaptor.ConvertRequest(c.Context(), req, true); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to convert request"})
	}
	user, err := pc.service.LoginFinish(sessionId, req)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(200).JSON(fiber.Map{
		"user": user,
	})
}
