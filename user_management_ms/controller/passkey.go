package controller

import (
	"log"
	"net/http"
	"strconv"
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
	log.Println("Start register start ")
	userID, err := strconv.Atoi(c.Params("userId"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid user id"})
	}

	options, err := pc.service.RegisterStart(&request.StartPasskeyRegistrationRequest{UserId: uint(userID)})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	log.Println("Finish register start ")
	return c.JSON(options)
}

func (pc *PasskeyController) RegisterFinish(c *fiber.Ctx) error {
	// 1. Parse user ID
	log.Println("Start register finish")
	userID, err := strconv.Atoi(c.Params("userId"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid user id"})
	}

	// 3. Convert Fiber (fasthttp) request to *http.Request
	req := new(http.Request)
	if err := fasthttpadaptor.ConvertRequest(c.Context(), req, true); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to convert request"})
	}

	// 4. Call service to finish registration
	if err := pc.service.RegisterFinish(uint(userID), req); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	log.Println("Finish register finish")
	return c.JSON(fiber.Map{"status": "ok"})
}

func (pc *PasskeyController) LoginStart(c *fiber.Ctx) error {
	log.Println("start login start-controller")
	userID, err := strconv.Atoi(c.Params("userId"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid user id"})
	}

	options, err := pc.service.LoginStart(uint(userID))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	log.Println("Finish login start-controller ")
	return c.JSON(options)
}

func (pc *PasskeyController) LoginFinish(c *fiber.Ctx) error {
	log.Println("start login finish-controller")
	userID, err := strconv.Atoi(c.Params("userId"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid user id"})
	}
	// Convert Fiber fasthttp request → *http.Request
	req := new(http.Request)
	if err := fasthttpadaptor.ConvertRequest(c.Context(), req, true); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to convert request"})
	}

	if err := pc.service.LoginFinish(uint(userID), req); err != nil {
		return c.Status(401).JSON(fiber.Map{"error": err.Error()})
	}
	log.Println("Finish login finish-controller ")
	return c.JSON(fiber.Map{"status": "ok"})
}
