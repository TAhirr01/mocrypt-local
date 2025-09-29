package controller

import (
	"log"
	"strconv"
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
	response, err := ac.googleService.LoginOrRegister(isNew, user)
	return c.Status(fiber.StatusOK).JSON(response)
}

func (ac *GoogleAuthController) GoogleRequestPhoneOTP(c *fiber.Ctx) error {
	userIdParam := c.Params("userId")
	userId, _ := strconv.Atoi(userIdParam)
	var req request.PhoneRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	log.Println(req.Phone)
	res, err := ac.googleService.StartGoogleRegistration(&request.StartGoogleRegistration{UserId: uint(userId), Phone: req.Phone})
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (ac *GoogleAuthController) GoogleVerifyRequestOTP(c *fiber.Ctx) error {
	userIdParam := c.Params("userId")
	userId, _ := strconv.Atoi(userIdParam)
	var req request.PhoneOTP
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	res, err := ac.googleService.VerifyPhoneOTP(&request.VerifyNumberOTPRequest{UserId: uint(userId), PhoneOTP: req.PhoneOTP})
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (ac *GoogleAuthController) GoogleVerifyLoginRequestOtp(c *fiber.Ctx) error {
	userIdParam := c.Params("userId")
	userId, _ := strconv.Atoi(userIdParam)
	var req request.EmailOTP
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	tokens, err := ac.googleService.VerifyGoogleLoginOtp(&request.VerifyEmailOTPRequest{UserId: uint(userId), EmailOTP: req.EmailOTP})
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(tokens)
}

func (ac *GoogleAuthController) CompleteGoogleRegistration(c *fiber.Ctx) error {
	userIdParam := c.Params("userId")
	userId, _ := strconv.Atoi(userIdParam)
	var compete request.BirthDateAndPassword
	if err := c.BodyParser(&compete); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	res, err := ac.googleService.CompleteGoogleRegistration(&request.CompleteGoogleRegistration{UserId: uint(userId), BirthDate: compete.BirthDate, Password: compete.Password})
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(res)
}
