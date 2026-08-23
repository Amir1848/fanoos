package dbutils

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	// Loaded first (dependency order guarantees it initializes before this
	// package's init()): makes .env variables available to the DB_* reads below.
	_ "github.com/amir1848/fanoos/utils/envutils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func init() {
	initDB()
}

func initDB() {
	var err error

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Tehran",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "postgres"),
		getEnv("DB_NAME", "fanoos"),
		getEnv("DB_PORT", "5432"),
	)

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				// DB_LOG_LEVEL: silent | error | warn (default) | info.
				// Values are never logged (ParameterizedQueries) so
				// passwords/hashes don't leak into stdout.
				SlowThreshold:             200 * time.Millisecond,
				LogLevel:                  logLevel(getEnv("DB_LOG_LEVEL", "warn")),
				IgnoreRecordNotFoundError: true,
				ParameterizedQueries:       true,
				Colorful:                   true,
			},
		),
	})

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("Failed to get SQL DB from GORM: %v", err)
	}

	sqlDB.SetMaxOpenConns(25)                 // Limit the maximum open connections
	sqlDB.SetMaxIdleConns(10)                 // Limit the number of idle connections
	sqlDB.SetConnMaxLifetime(5 * time.Minute) // Reuse connections for 5 minutes

	log.Println("Database connected successfully!")
}

func GetDB() *gorm.DB {
	return DB
}

func Transaction(txFunc func(tx *gorm.DB) error) error {
	tx := GetDB().Begin()

	if tx.Error != nil {
		return tx.Error
	}

	err := txFunc(tx)

	if err != nil {
		tx.Rollback()
		log.Println("Transaction rolled back due to error:", err)
		return err
	}

	tx.Commit()
	log.Println("Transaction committed successfully")
	return nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func logLevel(name string) logger.LogLevel {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "info":
		return logger.Info
	default:
		return logger.Warn
	}
}
