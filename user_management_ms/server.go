package main

import (
	"user_management_ms/config"
	"user_management_ms/controller"
	"user_management_ms/dtos/request"
	"user_management_ms/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"go.uber.org/zap"
)

type Server struct {
	AuthController       controller.IAuthController
	GoogleAuthController controller.IGoogleAuthController
	WebAuthnController   controller.IPasskeyController
	Logger               *zap.Logger
}

// NOTE: Server Constructor
func NewServer(
	AuthController controller.IAuthController,
	GoogleAuthController controller.IGoogleAuthController,
	WebAuthnController controller.IPasskeyController,
	Logger *zap.Logger,
) *Server {
	return &Server{
		AuthController:       AuthController,
		GoogleAuthController: GoogleAuthController,
		WebAuthnController:   WebAuthnController,
		Logger:               Logger,
	}
}

// NOTE: Start Fiber Server
func (s *Server) Start() *fiber.App {
	// NOTE: Initialize Fiber Server
	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5500",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// NOTE: Define API paths (context path and grouping by version)
	contextPath := app.Group(config.Conf.Application.Server.ContextPath)
	apiVersion := contextPath.Group(config.Conf.Application.Server.ApiVersion)

	authGroup := apiVersion.Group("/auth")
	authGroup.Use(middleware.LoggingMiddleware(s.Logger))
	authGroup.Post("/request-otp", middleware.ValidateBody[request.StartRegistration](), s.AuthController.RegisterRequestOTP)
	authGroup.Post("/verify-otp/:userId", middleware.ValidateBody[request.EmailAndPhoneOTP](), s.AuthController.VerifyRegisterOTP)
	authGroup.Post("/resend-otp", middleware.ValidateBody[request.OTPRequest](), s.AuthController.ResendOTP)
	authGroup.Post("/complete-registration/:userId", middleware.ValidateBody[request.BirthDateAndPassword](), s.AuthController.CompleteRegistration)
	authGroup.Post("/login", middleware.ValidateBody[request.LoginLocalRequest](), s.AuthController.LoginLocal)
	authGroup.Post("/verify-login-otp/:userId", middleware.ValidateBody[request.EmailAndPhoneOTP](), s.AuthController.VerifyLoginOTP)
	authGroup.Post("/refresh-token", middleware.ValidateBody[request.RefreshTokenReq](), s.AuthController.RefreshToken)
	authGroup.Get("/setup-2fa", middleware.AuthMiddleware(), s.AuthController.Setup2FA)
	authGroup.Post("/verify-2fa", middleware.ValidateBody[request.VerifyTwoFa](), middleware.AuthMiddleware(), s.AuthController.Verify2FA)
	authGroup.Post("/pin/set/:userId", middleware.ValidateBody[request.PinReq](), s.AuthController.SetPIN)
	authGroup.Post("/pin/verify/:userId", middleware.ValidateBody[request.PinReq](), s.AuthController.VerifyPIN)
	authGroup.Post("/qr", s.AuthController.QrLoginRequest)
	authGroup.Post("/qr/approve", middleware.AuthMiddleware(), s.AuthController.ApproveLoginRequest)
	authGroup.Post("/qr/:sessionId/status", s.AuthController.CheckLoginRequest)

	authGroup.Get("/google/call-back", s.GoogleAuthController.GoogleCallback)
	authGroup.Get("/google/login", s.GoogleAuthController.GoogleLogin)
	authGroup.Post("/google/request-otp/:userId", middleware.ValidateBody[request.PhoneRequest](), s.GoogleAuthController.GoogleRequestPhoneOTP)
	authGroup.Post("/google/verify-otp/:userId", middleware.ValidateBody[request.PhoneOTP](), s.GoogleAuthController.GoogleVerifyRequestOTP)
	authGroup.Post("/google/complete-registration/:userId", middleware.ValidateBody[request.BirthDateAndPassword](), s.GoogleAuthController.CompleteGoogleRegistration)
	authGroup.Post("/google/login/verify-otp/:userId", middleware.ValidateBody[request.EmailOTP](), s.GoogleAuthController.GoogleVerifyLoginRequestOtp)
	authGroup.Post("/google/login/resend-otp/:userId", s.GoogleAuthController.ResendOTP)

	authGroup.Post("/register/start", middleware.AuthMiddleware(), s.WebAuthnController.RegisterStart)
	authGroup.Post("/register/finish", middleware.AuthMiddleware(), s.WebAuthnController.RegisterFinish)
	authGroup.Post("/login/start", s.WebAuthnController.LoginStart)
	authGroup.Post("/login/finish/:sessionId", s.WebAuthnController.LoginFinish)
	return app
}
