package main

import (
	"notification-ms/config"
	"notification-ms/services"
	"os"

	"github.com/alasgarovnamig/confhandler"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

func main() {
	var configPath string
	env := os.Getenv("CONFIG_PATH")
	if env == "" {
		configPath = "./resources/application.yaml"
	} else {
		configPath = env
	}
	// NOTE: Graceful shutdown when panic time
	defer func() {
		if r := recover(); r != nil {
			os.Exit(1)
		}
	}()
	// NOTE: Configuration initialize...
	log.Info("Loading configuration...")
	err := confhandler.LoadConfigToStruct(configPath, &config.Conf)
	if err != nil {
		log.Panic("Error loading configuration file")
	}
	// NOTE: Logged successfully loaded config...
	log.Info("Configuration loaded successfully")

	emailService := services.NewEmailService()
	go emailService.ConsumeVerifyUserEvents()
	app := fiber.New()

	if err := app.Listen(config.Conf.Application.Server.Port); err != nil {
		log.Panicf("Failed to start server: %v", err)
	}

}
