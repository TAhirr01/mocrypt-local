package controller

import (
	"user_management_ms/dtos/request"
	"user_management_ms/services"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type IAuthController interface {
	LoginLocal(c *fiber.Ctx) error
	RegisterRequestOTP(c *fiber.Ctx) error
	VerifyRegisterOTP(c *fiber.Ctx) error
	CompleteRegistration(c *fiber.Ctx) error
	ResendOTP(c *fiber.Ctx) error
	VerifyLoginOTP(c *fiber.Ctx) error
	RefreshToken(c *fiber.Ctx) error
	Setup2FA(c *fiber.Ctx) error
	Verify2FA(c *fiber.Ctx) error
	SetPIN(c *fiber.Ctx) error
	VerifyPIN(c *fiber.Ctx) error
}

var validate = validator.New()

type AuthController struct {
	userService services.IUserService
}

func NewAuthController(service services.IUserService) IAuthController {
	return &AuthController{userService: service}
}

func (ac *AuthController) RegisterRequestOTP(c *fiber.Ctx) error {

	var req request.OTPRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if err := validate.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	response, err := ac.userService.RegisterRequestOTP(&req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(response)
}

func (ac *AuthController) VerifyRegisterOTP(c *fiber.Ctx) error {
	email := c.Query("email")
	phone := c.Query("phone")
	var req request.EmailAndPhoneOTP
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if err := validate.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	response, err := ac.userService.VerifyRegisterOTP(&request.VerifyOTPRequest{Email: email, Phone: phone, EmailOTP: req.EmailOTP, PhoneOTP: req.PhoneOTP})
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(response)
}

func (ac *AuthController) CompleteRegistration(c *fiber.Ctx) error {
	email := c.Query("email")
	var req request.BirthDateAndPassword
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if err := validate.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	response, err := ac.userService.CompleteRegistration(&request.CompleteRegisterRequest{Email: email, Password: req.Password, BirthDate: req.BirthDate})
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(response)
}

func (ac *AuthController) ResendOTP(c *fiber.Ctx) error {
	var req request.OTPRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{})
	}
	if err := validate.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{})
	}
	response, err := ac.userService.SendOTP(&req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{})
	}
	return c.Status(fiber.StatusOK).JSON(response)
}

func (ac *AuthController) LoginLocal(c *fiber.Ctx) error {
	var req request.LoginLocalRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if err := validate.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	response, err := ac.userService.LoginLocal(&req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(response)
}

func (ac *AuthController) VerifyLoginOTP(c *fiber.Ctx) error {
	email := c.Query("email")
	phone := c.Query("phone")
	var req request.EmailAndPhoneOTP
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if err := validate.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	response, err := ac.userService.VerifyLoginOTP(&request.VerifyOTPRequest{Email: email, Phone: phone, EmailOTP: req.EmailOTP, PhoneOTP: req.PhoneOTP})
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(response)
}

func (ac *AuthController) RefreshToken(c *fiber.Ctx) error {
	var req *request.RefreshTokenReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	response, err := ac.userService.RefreshToken(req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(response)
}

func (ac *AuthController) Setup2FA(c *fiber.Ctx) error {
	email := c.Query("email")
	phone := c.Query("phone")

	png, err := ac.userService.Setup2FA(email, phone)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	c.Set("Content-Type", "image/png")
	return c.Send(png)
}

func (ac *AuthController) Verify2FA(c *fiber.Ctx) error {
	email := c.Query("email")
	phone := c.Query("phone")

	var body struct {
		Code string `json:"code"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	verified, err := ac.userService.Verify2FA(email, phone, body.Code)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if !verified {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid 2FA code"})
	}

	return c.JSON(fiber.Map{"message": "2FA verified, access granted"})
}

func (ac *AuthController) SetPIN(c *fiber.Ctx) error {
	req := request.PINRequest{}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	err := ac.userService.SetPIN(req.Email, req.Phone, req.PIN)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "PIN set successfully"})
}

func (ac *AuthController) VerifyPIN(c *fiber.Ctx) error {
	req := request.PINRequest{}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	valid, err := ac.userService.VerifyPIN(req.Email, req.Phone, req.PIN)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if !valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid PIN"})
	}

	return c.JSON(fiber.Map{"message": "PIN verified"})
}

//func (ac *AuthController) CheckLogin(c *fiber.Ctx) error {
//	var req *request.CheckLogin
//	if err := c.BodyParser(&req); err != nil {
//		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
//			"error": err.Error(),
//		})
//	}
//	if ac.userService.ShouldForceFullAuth(req, 30*24*time.Hour) {
//		return c.JSON(fiber.Map{"login_type": "full"})
//	}
//	return c.JSON(fiber.Map{"login_type": "passkey"})
//}
