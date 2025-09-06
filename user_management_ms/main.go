package main

import (
	"user_management_ms/config"

	"github.com/alasgarovnamig/confhandler"
	"github.com/gofiber/fiber/v2/log"

	"os"
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

	log.Info("Starting server...")
	s := new(service)
	s.Start()
}
