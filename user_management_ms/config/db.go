package config

import (
	"errors"
	def_log "log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlserver"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	gorm_logger "gorm.io/gorm/logger"
)

func OpenDatabaseConnection(url string) *gorm.DB {
	//Migrate()
	log.Info("Opening database connection")

	gormLogger := gorm_logger.New(
		def_log.New(os.Stdout, "\r\n", def_log.LstdFlags),
		gorm_logger.Config{
			LogLevel:                  gorm_logger.Info,
			Colorful:                  true,
			SlowThreshold:             time.Second,
			IgnoreRecordNotFoundError: true,
		},
	)

	// NOTE: Open database connection
	db, err := gorm.Open(sqlserver.Open(url), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		log.Panic("Failed to open database connection")
	} else {
		log.Info("Successfully opened database connection")
	}

	log.Info("retrieving database instance from GORM")
	sqlDB, err := db.DB()
	if err != nil {
		log.Panic("failed to retrieve database instance from GORM")
	} else {
		log.Info("successfully retrieved database instance from GORM")
	}

	log.Info("configuring database connection pool settings...")
	sqlDB.SetMaxIdleConns(Conf.Application.Datasource.MaxIdleConnections) // Aynı anda boşta bekleyebilecek maksimum bağlantı sayısı
	sqlDB.SetMaxOpenConns(Conf.Application.Datasource.MaxOpenConnections) // Aynı anda açık olabilecek maksimum bağlantı sayısı
	sqlDB.SetConnMaxLifetime(time.Minute * time.Duration(Conf.Application.Datasource.ConnectionMaxLifetime))

	log.Info("database connection pool successfully configured")
	return db
}

func Migrate(url string) {

	log.Info("configuring migration instance settings...")
	// NOTE: Migration instance creating...
	m, err := migrate.New(
		Conf.Application.Migration,
		url,
	)
	if err != nil {
		log.Panic("failed to create migration instance")
	} else {

		log.Info("migration instance successfully configured")
	}

	log.Info("migration applying...")
	// NOTE: Apply migration operations
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Panic("failed to run migration: ", err)
		panic(err.Error())
	}

	log.Info("database migrated successfully")
}

func CloseDatabaseConnection(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		//logger.Logger.
		//	WithField("application_display_name", Conf.Application.DisplayName).
		//	WithField("error", err).Error("Error occurred while closing the database connection")
		return
	}
	// NOTE: Close database connection
	if err := sqlDB.Close(); err != nil {
		log.Error("Failed to close the database connection")
	} else {
		log.Info("Database connection closed successfully")
	}
}
