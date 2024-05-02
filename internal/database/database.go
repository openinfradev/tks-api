package database

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/openinfradev/tks-api/internal/pagination"

	"github.com/openinfradev/tks-api/internal/delivery/api"

	internal_gorm "github.com/openinfradev/tks-api/internal/gorm"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/repository"
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
	newLogger := internal_gorm.NewGormLogger().LogMode(level)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
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
	if err := db.AutoMigrate(&model.CacheEmailCode{},
		&model.ExpiredTokenTime{},
		&model.Role{},
		&model.CloudAccount{},
		&model.StackTemplate{},
		&model.Organization{},
		&model.User{},
		&model.Cluster{},
		&model.ClusterFavorite{},
		&model.AppGroup{},
		&model.Application{},
		&model.AppServeApp{},
		&model.AppServeAppTask{},
		&model.SystemNotification{},
		&model.SystemNotificationAction{},
		&model.SystemNotificationMetricParameter{},
		&model.SystemNotificationTemplate{},
		&model.SystemNotificationRule{},
		&model.SystemNotificationCondition{},
		&model.Project{},
		&model.ProjectMember{},
		&model.ProjectNamespace{},
		&model.ProjectRole{},
		&model.Audit{},
		&model.PolicyTemplateSupportedVersion{},
		&model.PolicyTemplate{},
		&model.Policy{},
		&model.Dashboard{},
	); err != nil {
		return err
	}

	if err := db.AutoMigrate(&model.Permission{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&model.Endpoint{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&model.PermissionEndpoint{}); err != nil {
		return err
	}
	return nil
}

func EnsureDefaultRows(db *gorm.DB) error {
	// Create default rows
	repoFactory := repository.Repository{
		Auth:                       repository.NewAuthRepository(db),
		User:                       repository.NewUserRepository(db),
		Cluster:                    repository.NewClusterRepository(db),
		Organization:               repository.NewOrganizationRepository(db),
		AppGroup:                   repository.NewAppGroupRepository(db),
		AppServeApp:                repository.NewAppServeAppRepository(db),
		CloudAccount:               repository.NewCloudAccountRepository(db),
		StackTemplate:              repository.NewStackTemplateRepository(db),
		SystemNotification:         repository.NewSystemNotificationRepository(db),
		SystemNotificationRule:     repository.NewSystemNotificationRuleRepository(db),
		SystemNotificationTemplate: repository.NewSystemNotificationTemplateRepository(db),
		Role:                       repository.NewRoleRepository(db),
		Permission:                 repository.NewPermissionRepository(db),
		Endpoint:                   repository.NewEndpointRepository(db),
		Project:                    repository.NewProjectRepository(db),
		Dashboard:                  repository.NewDashboardRepository(db),
	}

	//

	ctx := context.Background()
	pg := pagination.NewPagination(nil)
	pg.Limit = 1000
	eps, err := repoFactory.Endpoint.List(ctx, pg)
	if err != nil {
		return err
	}

	var storedEps = make(map[string]struct{})
	for _, ep := range eps {
		storedEps[ep.Name] = struct{}{}
	}
	for _, ep := range api.MapWithEndpoint {
		if _, ok := storedEps[ep.Name]; !ok {
			if err := repoFactory.Endpoint.Create(ctx, &model.Endpoint{
				Name:  ep.Name,
				Group: ep.Group,
			}); err != nil {
				return err
			}
		}
	}

	return nil
}
