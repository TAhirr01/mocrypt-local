package controller

import (
	"log"
	"user_management_ms/dtos/request"
	"user_management_ms/services"

	"github.com/gofiber/fiber/v2"
	"github.com/hashicorp/go-uuid"
)

type IGoogleAuthController interface {
	GoogleLogin(c *fiber.Ctx) error
	GoogleCallback(c *fiber.Ctx) error
	GoogleRequestPhoneOTP(c *fiber.Ctx) error
	CompleteGoogleRegistration(c *fiber.Ctx) error
	GoogleVerifyRequestOTP(c *fiber.Ctx) error
	GoogleVerifyLoginRequestOtp(c *fiber.Ctx) error
}

type GoogleAuthController struct {
	googleService services.IGoogleAuthService
}

func NewGoogleAuthController(googleService services.IGoogleAuthService) IGoogleAuthController {
	return &GoogleAuthController{googleService: googleService}
}

func (ac *GoogleAuthController) GoogleLogin(c *fiber.Ctx) error {
	state, _ := uuid.GenerateUUID()
	url := ac.googleService.LoginGoogle(state)

	return c.Redirect(url, fiber.StatusTemporaryRedirect)
}
func (ac *GoogleAuthController) GoogleCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	if code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "code is required",
		})
	}
	// 1. Exchange code for token
	token, err := ac.googleService.ExchangeGoogleToken(code)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	idToken, ok := token.Extra("id_token").(string)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "id_token is required",
		})
	}
	// 2. Verify Google ID token
	userInfo, err := ac.googleService.VerifyGoogleIDToken(idToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	// 3. Create or link Google user
	user, isNew, err := ac.googleService.CreteNewGoogleUser(userInfo.Email, userInfo.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	// 4. Response based on whether it's a new user or existing
	if isNew {
		// New user → needs to complete registration
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":             "new_user",
			"message":            "Phone number verification is needed",
			"phone_verification": user.PhoneVerified,
			"email":              user.Email,
			"google_id":          user.GoogleID,
		})
	}
	if isNew == false && !user.PhoneVerified {
		if _, err := ac.googleService.SendPhoneVerificationOtp(&request.OTPRequestPhone{Phone: user.Phone}); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":             "existing_user",
			"message":            "user already exists send to verify phone ",
			"phone_verification": user.PhoneVerified,
			"phone":              user.Phone,
			"email":              user.Email,
			"google_id":          user.GoogleID,
		})
	}
	if isNew == false && user.GoogleID != "" {
		if _, err := ac.googleService.SendEmailLoginOtp(&request.OTPRequestEmail{Email: user.Email}); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":             "existing_user",
			"message":            "user already exists send to verification ",
			"phone_verification": user.PhoneVerified,
			"phone":              user.Phone,
			"email":              user.Email,
			"google_id":          user.GoogleID,
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{})
}

func (ac *GoogleAuthController) GoogleRequestPhoneOTP(c *fiber.Ctx) error {
	email := c.Query("email")
	log.Println(email)
	var req request.PhoneRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	log.Println(req.Phone)
	res, err := ac.googleService.StartGoogleRegistration(&request.StartGoogleRegistration{Email: email, Phone: req.Phone})
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (ac *GoogleAuthController) GoogleVerifyRequestOTP(c *fiber.Ctx) error {
	email := c.Params("email")
	phone := c.Query("phone")
	var req request.PhoneOTP
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	res, err := ac.googleService.VerifyPhoneOTP(&request.VerifyNumberOTPRequest{Email: email, Phone: phone, PhoneOTP: req.PhoneOTP})
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (ac *GoogleAuthController) GoogleVerifyLoginRequestOtp(c *fiber.Ctx) error {
	email := c.Query("email")
	var req request.EmailOTP
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	tokens, err := ac.googleService.VerifyGoogleLoginOtp(&request.VerifyEmailOTPRequest{Email: email, EmailOTP: req.EmailOTP})
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{})
	}
	return c.Status(fiber.StatusOK).JSON(tokens)
}

func (ac *GoogleAuthController) CompleteGoogleRegistration(c *fiber.Ctx) error {
	email := c.Query("email")
	var compete request.BirthDateAndPassword
	if err := c.BodyParser(&compete); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	res, err := ac.googleService.CompleteGoogleRegistration(&request.CompleteGoogleRegistration{Email: email, BirthDate: compete.BirthDate, Password: compete.Password})
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(res)
}
