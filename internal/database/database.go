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
	"github.com/openinfradev/tks-api/pkg/domain"
)

func InitDB() (*gorm.DB, error) {
	// Connect to gormDB
	dsn := fmt.Sprintf(
		"host=%s dbname=%s user=%s password=%s port=%d sslmode=disable TimeZone=Asia/Seoul",
		viper.GetString("dbhost"),
		viper.GetString("dbname"),
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

	if viper.GetInt("migrate-db") == 1 {
		if err := migrateSchema(db); err != nil {
			return nil, err
		}
	}

	return db, nil
}

func migrateSchema(db *gorm.DB) error {
	// Auth
	if err := db.AutoMigrate(&repository.CacheEmailCode{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&repository.User{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&repository.Role{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&repository.Policy{}); err != nil {
		return err
	}

	// Organization
	if err := db.AutoMigrate(&repository.Organization{}); err != nil {
		return err
	}

	// CloudAccount
	if err := db.AutoMigrate(&repository.CloudAccount{}); err != nil {
		return err
	}

	// StackTemplate
	if err := db.AutoMigrate(&repository.StackTemplate{}); err != nil {
		return err
	}

	// Cluster
	if err := db.AutoMigrate(&repository.Cluster{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&repository.ClusterFavorite{}); err != nil {
		return err
	}

	// Services
	if err := db.AutoMigrate(&repository.AppGroup{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&repository.Application{}); err != nil {
		return err
	}

	// AppServe
	if err := db.AutoMigrate(&domain.AppServeApp{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&domain.AppServeAppTask{}); err != nil {
		return err
	}

	// Alert
	if err := db.AutoMigrate(&repository.Alert{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&repository.AlertAction{}); err != nil {
		return err
	}

	// Project
	if err := db.AutoMigrate(&domain.Project{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&domain.ProjectRole{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&domain.ProjectMember{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&domain.ProjectNamesapce{}); err != nil {
		return err
	}

	return nil
}
