package database

import (
	"fmt"
	"os"
	"strings"

	"github.com/openinfradev/tks-api/internal/delivery/api"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	indomain "github.com/openinfradev/tks-api/internal/domain"
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
	if err := db.AutoMigrate(&domain.User{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&domain.Role{}); err != nil {
		return err
	}

	// Organization
	if err := db.AutoMigrate(&domain.Organization{}); err != nil {
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

	// Role
	if err := db.AutoMigrate(&domain.Role{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&domain.Permission{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&domain.Endpoint{}); err != nil {
		return err
	}

	// Project
	if err := db.AutoMigrate(&indomain.Project{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&indomain.ProjectMember{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&indomain.ProjectNamespace{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&indomain.ProjectRole{}); err != nil {
		return err
	}

	return nil
}

func EnsureDefaultRows(db *gorm.DB) error {
	// Create default rows
	repoFactory := repository.Repository{
		Auth:          repository.NewAuthRepository(db),
		User:          repository.NewUserRepository(db),
		Cluster:       repository.NewClusterRepository(db),
		Organization:  repository.NewOrganizationRepository(db),
		AppGroup:      repository.NewAppGroupRepository(db),
		AppServeApp:   repository.NewAppServeAppRepository(db),
		CloudAccount:  repository.NewCloudAccountRepository(db),
		StackTemplate: repository.NewStackTemplateRepository(db),
		Alert:         repository.NewAlertRepository(db),
		Role:          repository.NewRoleRepository(db),
		Permission:    repository.NewPermissionRepository(db),
		Endpoint:      repository.NewEndpointRepository(db),
		Project:       repository.NewProjectRepository(db),
	}

	//
	eps, err := repoFactory.Endpoint.List(nil)
	if err != nil {
		return err
	}

	var storedEps = make(map[string]struct{})
	for _, ep := range eps {
		storedEps[ep.Name] = struct{}{}
	}
	for _, ep := range api.ApiMap {
		if _, ok := storedEps[ep.Name]; !ok {
			if err := repoFactory.Endpoint.Create(&domain.Endpoint{
				Name:  ep.Name,
				Group: ep.Group,
			}); err != nil {
				return err
			}
		}
	}

	// Audit
	if err := db.AutoMigrate(&repository.Audit{}); err != nil {
		return err
	}

	return nil
}
