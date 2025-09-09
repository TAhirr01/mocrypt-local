package main

import (
	"context"
	"user_management_ms/repository"
	"user_management_ms/services"

	"os"
	"os/signal"
	"syscall"
	"time"
	"user_management_ms/config"
	"user_management_ms/controller"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

type service struct {
	//DB
	dbConnection *gorm.DB

	//Redis Client
	redisClient *redis.Client

	//OAuth2 Conf
	oauthConfig *oauth2.Config

	//WebAuthn Conf
	webAuthn *webauthn.WebAuthn

	// Repository
	userRepository   repository.IUserRepository
	googleRepository repository.IGoogleRepository

	// Service
	userService    services.IUserService
	jwtService     services.IJWTService
	googleService  services.IGoogleAuthService
	redisService   services.IRedisService
	passkeyService services.IPasskeyService

	// Controller
	authController       controller.IAuthController
	googleAuthController controller.IGoogleAuthController
	passkeyController    controller.IPasskeyController
}

// NOTE: Service Start
func (s *service) Start() {
	log.Info("Opening database connection...")
	s.dbConnection = config.OpenDatabaseConnection(config.Conf.Application.Datasource.PrimaryURL)
	config.Migrate(config.Conf.Application.Datasource.PrimaryURL)

	log.Info("Opening redis connection...")
	s.redisClient = config.ConnectToRedis(config.Conf.Application.Redis.Host)

	log.Info("OAuth2 config")
	s.oauthConfig = config.InitOAuth()

	log.Info("WebAuthn config")
	s.webAuthn = config.InitWebAuthn()
	// NOTE: Dependency Injections
	log.Info("WebAuthn configurated successfully")
	s.DependencyInjection()
	//TODO: coment and log

	// NOTE: Start Fiber server...
	app := NewServer(s.authController, s.googleAuthController, s.passkeyController).Start()

	log.Info("Server starting..")
	// NOTE: Server start with goroutine
	go func() {
		if err := app.Listen(config.Conf.Application.Server.Port); err != nil {
			log.Fatal("Server failed to start")
		}
	}()
	// NOTE: Keep OS signals for graceful shutdown
	s.gracefulShutdown(app)
}

// NOTE: Depency Injection Operation
func (s *service) DependencyInjection() {
	// NOTE: JWT services configured and initialize...
	s.jwtService = &services.JWTService{
		Secret:     []byte(config.Conf.Application.Security.Secret),
		Issuer:     config.Conf.Application.Security.Issuer,
		AccessTTL:  time.Duration(config.Conf.Application.Security.TokenValidityInSeconds) * time.Second,
		RefreshTTL: time.Duration(config.Conf.Application.Security.TokenValidityInSecondsForRememberMe) * time.Second,
	}
	// NOTE: Repositories Injections
	s.userRepository = repository.NewUserRepository()
	s.googleRepository = repository.NewGoogleRepository()
	// NOTE: Services Injections
	s.redisService = services.NewRedisService(s.redisClient)
	s.userService = services.NewUserService(s.dbConnection, s.userRepository, s.redisService, s.jwtService)
	s.googleService = services.NewGoogleAuthService(s.dbConnection, s.oauthConfig, s.googleRepository, s.jwtService, s.redisService)
	s.passkeyService = services.NewPasskeyService(s.webAuthn, s.dbConnection, s.userRepository, s.redisService)
	// NOTE: Controllers Injections
	s.authController = controller.NewAuthController(s.userService)
	s.googleAuthController = controller.NewGoogleAuthController(s.googleService)
	s.passkeyController = controller.NewPasskeyController(s.passkeyService)

}

// NOTE: Graceful shutdown operation
func (s *service) gracefulShutdown(app *fiber.App) {

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// NOTE:Server Shutdown when keep signal
	<-sigChan
	log.Info("Shutting down server...")
	// NOTE: Creating context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// NOTE: Shutdown Fiber server
	if err := app.Shutdown(); err != nil {
		log.Error("error while shutting down app", err)
	}

	// NOTE: Shutdown Database connection
	done := make(chan bool)
	go func() {
		config.CloseDatabaseConnection(s.dbConnection)
		done <- true
	}()

	select {
	case <-ctx.Done():
		log.Error("timeout while shutting down database", ctx.Err())
	case <-done:
		log.Info("database is gracefully shutdown", ctx.Err())
	}
}
