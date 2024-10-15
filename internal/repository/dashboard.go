package repository

import (
	"context"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type IDashboardRepository interface {
	CreateDashboard(ctx context.Context, d *model.Dashboard) (string, error)
	GetDashboardById(ctx context.Context, organizationId string, dashboardId string) (*model.Dashboard, error)
	GetDashboardByUserId(ctx context.Context, organizationId string, userId string, dashboardKey string) (*model.Dashboard, error)
	UpdateDashboard(ctx context.Context, d *model.Dashboard) error
}

type DashboardRepository struct {
	db *gorm.DB
}

func NewDashboardRepository(db *gorm.DB) IDashboardRepository {
	return &DashboardRepository{
		db: db,
	}
}

func (dr DashboardRepository) CreateDashboard(ctx context.Context, d *model.Dashboard) (string, error) {
	res := dr.db.WithContext(ctx).Create(&d)
	if res.Error != nil {
		log.Error(ctx, res.Error)
		return "", res.Error
	}

	return d.ID.String(), nil
}

func (dr DashboardRepository) GetDashboardById(ctx context.Context, organizationId string, dashboardId string) (d *model.Dashboard, err error) {
	res := dr.db.WithContext(ctx).Limit(1).
		Where("organization_id = ? and id = ?", organizationId, dashboardId).
		First(&d)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			log.Info(ctx, "Cannot find dashboard")
			return nil, nil
		} else {
			log.Error(ctx, res.Error)
			return nil, res.Error
		}
	}
	return d, nil
}

func (dr DashboardRepository) GetDashboardByUserId(ctx context.Context, organizationId string, userId string, dashboardKey string) (d *model.Dashboard, err error) {
	res := dr.db.WithContext(ctx).Limit(1).
		Where("organization_id = ? and user_id = ? and key = ?", organizationId, userId, dashboardKey).
		First(&d)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			log.Info(ctx, "Cannot find dashboard")
			return nil, nil
		} else {
			log.Error(ctx, res.Error)
			return nil, res.Error
		}
	}
	return d, nil
}

func (dr DashboardRepository) UpdateDashboard(ctx context.Context, d *model.Dashboard) error {
	res := dr.db.WithContext(ctx).Model(&d).
		Updates(model.Dashboard{Content: d.Content})
	if res.Error != nil {
		log.Error(ctx, res.Error)
		return res.Error
	}
	return nil
}
