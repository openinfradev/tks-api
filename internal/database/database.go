package database

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/openinfradev/tks-api/internal/repository"
)

var gormDB *gorm.DB

func InitDB() (*gorm.DB, error) {
	// Connect to gormDB
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s port=%d dbname=tks sslmode=disable TimeZone=Asia/Seoul",
		viper.GetString("dbhost"),
		viper.GetString("dbuser"),
		viper.GetString("dbpassword"),
		viper.GetInt("dbport"),
	)

	// Set log level for gorm
	var level logger.LogLevel
	switch strings.ToUpper(os.Getenv("LOG_LEVEL")) {
	case "DEBUG":
		level = logger.Info
	case "TEST":
		level = logger.Silent
	default:
		level = logger.Silent
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(level),
	})
	if err != nil {
		return nil, err
	}

	if err := MigrateSchema(db); err != nil {
		return nil, err
	}

	return db, nil
}

func MigrateSchema(db *gorm.DB) error {
	if err := db.AutoMigrate(&repository.History{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&repository.Contract{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&repository.Cluster{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&repository.AppGroup{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&repository.AppServeApp{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&repository.AppServeAppTask{}); err != nil {
		return err
	}
	return nil
}
