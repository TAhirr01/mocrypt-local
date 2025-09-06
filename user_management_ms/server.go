package main

import (
	"user_management_ms/config"
	"user_management_ms/controller"

	"github.com/gofiber/fiber/v2"
)

type Server struct {
	AuthController       controller.IAuthController
	GoogleAuthController controller.IGoogleAuthController
}

// NOTE: Server Constructor
func NewServer(
	AuthController controller.IAuthController,
	GoogleAuthController controller.IGoogleAuthController,
) *Server {
	return &Server{
		AuthController:       AuthController,
		GoogleAuthController: GoogleAuthController,
	}
}

// NOTE: Start Fiber Server
func (s *Server) Start() *fiber.App {
	// NOTE: Initialize Fiber Server
	app := fiber.New()

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
	authGroup.Get("/google/call-back", s.GoogleAuthController.GoogleCallback)
	authGroup.Get("/google/login", s.GoogleAuthController.GoogleLogin)
	authGroup.Post("/google/request-otp", s.GoogleAuthController.GoogleRequestPhoneOTP)
	authGroup.Post("/google/verify-otp/:email", s.GoogleAuthController.GoogleVerifyRequestOTP)
	authGroup.Post("/google/complete-registration", s.GoogleAuthController.CompleteGoogleRegistration)
	authGroup.Post("/google/login/verify-otp", s.GoogleAuthController.GoogleVerifyLoginRequestOtp)
	return app
}

//func (s *Server) configureAuthGroup(router fiber.Router) {
//	authGroup := router.Group("/auth")
//	//authGroup.Post("/request-otp", s.AuthController.RequestOTP)
//	authGroup.Post("/verify-otp", s.AuthController.VerifyOTP)
//	authGroup.Post("/complete-registration", s.AuthController.CompleteRegistration)
//}
