package main

import (
	"user_management_ms/config"
	"user_management_ms/controller"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

type Server struct {
	AuthController       controller.IAuthController
	GoogleAuthController controller.IGoogleAuthController
	WebAuthnController   controller.IPasskeyController
}

// NOTE: Server Constructor
func NewServer(
	AuthController controller.IAuthController,
	GoogleAuthController controller.IGoogleAuthController,
	WebAuthnController controller.IPasskeyController,
) *Server {
	return &Server{
		AuthController:       AuthController,
		GoogleAuthController: GoogleAuthController,
		WebAuthnController:   WebAuthnController,
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

	//s.configureAuthGroup(apiVersion)
	authGroup := apiVersion.Group("/auth")
	authGroup.Post("/request-otp", s.AuthController.RegisterRequestOTP)
	authGroup.Post("/verify-otp", s.AuthController.VerifyRegisterOTP)
	authGroup.Post("/resend-otp", s.AuthController.ResendOTP)
	authGroup.Post("/complete-registration", s.AuthController.CompleteRegistration)
	authGroup.Post("/login", s.AuthController.LoginLocal)
	authGroup.Post("/verify-login-otp", s.AuthController.VerifyLoginOTP)
	authGroup.Post("/refresh-token", s.AuthController.RefreshToken)
	authGroup.Get("/setup-2fa", s.AuthController.Setup2FA)
	authGroup.Post("/verify-2fa", s.AuthController.Verify2FA)

	authGroup.Get("/google/call-back", s.GoogleAuthController.GoogleCallback)
	authGroup.Get("/google/login", s.GoogleAuthController.GoogleLogin)
	authGroup.Post("/google/request-otp", s.GoogleAuthController.GoogleRequestPhoneOTP)
	authGroup.Post("/google/verify-otp/:email", s.GoogleAuthController.GoogleVerifyRequestOTP)
	authGroup.Post("/google/complete-registration", s.GoogleAuthController.CompleteGoogleRegistration)
	authGroup.Post("/google/login/verify-otp", s.GoogleAuthController.GoogleVerifyLoginRequestOtp)

	authGroup.Post("/register/start/:userId", s.WebAuthnController.RegisterStart)
	authGroup.Post("/register/finish/:userId", s.WebAuthnController.RegisterFinish)
	authGroup.Post("/login/start/:userId", s.WebAuthnController.LoginStart)

	//authGroup.Post("/login/finish/:userId", s.WebAuthnController.LoginFinish)
	return app
}

//func (s *Server) configureAuthGroup(router fiber.Router) {
//	authGroup := router.Group("/auth")
//	//authGroup.Post("/request-otp", s.AuthController.RequestOTP)
//	authGroup.Post("/verify-otp", s.AuthController.VerifyOTP)
//	authGroup.Post("/complete-registration", s.AuthController.CompleteRegistration)
//}
