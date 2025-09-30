package controller

import (
	"encoding/base64"
	"log"
	"strconv"
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
	QrLoginRequest(c *fiber.Ctx) error
	ApproveLoginRequest(c *fiber.Ctx) error
	CheckLoginRequest(c *fiber.Ctx) error
}

var validate = validator.New()

type AuthController struct {
	userService services.IUserService
	pin         services.IPinService
	qrLogin     services.IQRLoginService
	otp         services.IOtp
}

func NewAuthController(service services.IUserService, pin services.IPinService, otp services.IOtp, qrlogin services.IQRLoginService) IAuthController {
	return &AuthController{userService: service, otp: otp, pin: pin, qrLogin: qrlogin}
}

func (ac *AuthController) CheckLoginRequest(c *fiber.Ctx) error {
	sessionId := c.Params("sessionId")
	response, err := ac.qrLogin.CheckLoginQr(sessionId)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(response)
}

func (ac *AuthController) ApproveLoginRequest(c *fiber.Ctx) error {
	userId := c.Locals("userId")

	var req struct {
		SessionId string `json:"sessionId"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if err := ac.qrLogin.ApproveLoginQr(uint(userId.(float64)), req.SessionId); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	log.Println(userId)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "QR login approved",
	})
}

func (ac *AuthController) QrLoginRequest(c *fiber.Ctx) error {
	png, sessionId, err := ac.qrLogin.RequestLoginQr()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"base64png": base64.StdEncoding.EncodeToString(png),
		"sessionId": sessionId,
	})
}

func (ac *AuthController) RegisterRequestOTP(c *fiber.Ctx) error {

	var req request.StartRegistration
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
	userIdParam := c.Params("userId")
	userId, err := strconv.Atoi(userIdParam)
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
	response, err := ac.otp.VerifyOTPs(&request.VerifyOTPRequest{UserId: uint(userId), EmailOTP: req.EmailOTP, PhoneOTP: req.PhoneOTP})
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(response)
}

func (ac *AuthController) CompleteRegistration(c *fiber.Ctx) error {
	userIdParam := c.Params("userId")
	userId, err := strconv.Atoi(userIdParam)
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
	response, err := ac.userService.AddPasswordBirthdate(&request.CompleteRegisterRequest{UserId: uint(userId), Password: req.Password, BirthDate: req.BirthDate})
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
	response, err := ac.otp.ResendRegisterOtp(&req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
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
	response, err := ac.userService.LoginLocal(&req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(response)
}

func (ac *AuthController) VerifyLoginOTP(c *fiber.Ctx) error {
	userIdParam := c.Params("userId")
	userId, err := strconv.Atoi(userIdParam)
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
	log.Println(userId)
	response, err := ac.userService.VerifyLoginOTP(&request.VerifyOTPRequest{UserId: uint(userId), EmailOTP: req.EmailOTP, PhoneOTP: req.PhoneOTP})
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
	userId := c.Locals("userId")

	resp, err := ac.userService.Setup2FA(uint(userId.(float64)))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"secret":  resp.Secret,
		"qr_code": base64.StdEncoding.EncodeToString(resp.QRCode),
	})
}

func (ac *AuthController) Verify2FA(c *fiber.Ctx) error {
	userId := c.Locals("userId")

	var body struct {
		Code string `json:"code"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	verified, err := ac.userService.Verify2FA(uint(userId.(float64)), body.Code)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if !verified {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid 2FA code"})
	}

	return c.JSON(fiber.Map{"message": "2FA verified, access granted"})
}

func (ac *AuthController) SetPIN(c *fiber.Ctx) error {
	userIdParam := c.Params("userId")
	userId, _ := strconv.Atoi(userIdParam)

	req := struct {
		Pin string `json:"pin"`
	}{}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	if err := ac.pin.SetPIN(uint(userId), req.Pin); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "PIN set successfully"})
}

func (ac *AuthController) VerifyPIN(c *fiber.Ctx) error {
	userIdParam := c.Params("userId")
	userId, _ := strconv.Atoi(userIdParam)

	req := struct {
		Pin string `json:"pin"`
	}{}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	tokens, valid, err := ac.pin.VerifyPIN(uint(userId), req.Pin)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if !valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid PIN"})
	}

	return c.Status(200).JSON(tokens)
}
