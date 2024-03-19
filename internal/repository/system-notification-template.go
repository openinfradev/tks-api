package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/pkg/log"
)

// Interfaces
type ISystemNotificationTemplateRepository interface {
	Get(ctx context.Context, systemNotificationTemplateId uuid.UUID) (model.SystemNotificationTemplate, error)
	GetByName(ctx context.Context, name string) (model.SystemNotificationTemplate, error)
	Fetch(ctx context.Context, pg *pagination.Pagination) ([]model.SystemNotificationTemplate, error)
	Create(ctx context.Context, dto model.SystemNotificationTemplate) (systemNotificationTemplateId uuid.UUID, err error)
	Update(ctx context.Context, dto model.SystemNotificationTemplate) (err error)
	Delete(ctx context.Context, dto model.SystemNotificationTemplate) (err error)
	UpdateOrganizations(ctx context.Context, systemNotificationTemplateId uuid.UUID, organizations []model.Organization) (err error)
}

type SystemNotificationTemplateRepository struct {
	db *gorm.DB
}

func NewSystemNotificationTemplateRepository(db *gorm.DB) ISystemNotificationTemplateRepository {
	return &SystemNotificationTemplateRepository{
		db: db,
	}
}

// Logics
func (r *SystemNotificationTemplateRepository) Get(ctx context.Context, systemNotificationTemplateId uuid.UUID) (out model.SystemNotificationTemplate, err error) {
	res := r.db.WithContext(ctx).Preload(clause.Associations).First(&out, "id = ?", systemNotificationTemplateId)
	if res.Error != nil {
		return out, res.Error
	}
	return
}

func (r *SystemNotificationTemplateRepository) GetByName(ctx context.Context, name string) (out model.SystemNotificationTemplate, err error) {
	res := r.db.WithContext(ctx).Preload(clause.Associations).First(&out, "name = ?", name)
	if res.Error != nil {
		return out, res.Error
	}
	return
}

func (r *SystemNotificationTemplateRepository) Fetch(ctx context.Context, pg *pagination.Pagination) (out []model.SystemNotificationTemplate, err error) {
	if pg == nil {
		pg = pagination.NewPagination(nil)
	}

	_, res := pg.Fetch(r.db.WithContext(ctx).Preload(clause.Associations), &out)
	if res.Error != nil {
		return nil, res.Error
	}
	return
}

func (r *SystemNotificationTemplateRepository) Create(ctx context.Context, dto model.SystemNotificationTemplate) (systemNotificationTemplateId uuid.UUID, err error) {
	dto.ID = uuid.New()
	res := r.db.WithContext(ctx).Create(&dto)
	if res.Error != nil {
		return uuid.Nil, res.Error
	}

	err = r.db.WithContext(ctx).Model(&dto).Association("MetricParameters").Replace(dto.MetricParameters)
	if err != nil {
		log.Error(ctx, err)
	}

	return dto.ID, nil
}

func (r *SystemNotificationTemplateRepository) Update(ctx context.Context, dto model.SystemNotificationTemplate) (err error) {
	res := r.db.WithContext(ctx).Model(&model.SystemNotificationTemplate{}).
		Where("id = ?", dto.ID).
		Updates(map[string]interface{}{
			"Name":        dto.Name,
			"Description": dto.Description,
			"MetricQuery": dto.MetricQuery,
		})
	if res.Error != nil {
		return res.Error
	}

	if err = r.db.WithContext(ctx).Model(&dto).Association("MetricParameters").Unscoped().Replace(dto.MetricParameters); err != nil {
		return err
	}
	return nil
}

func (r *SystemNotificationTemplateRepository) Delete(ctx context.Context, dto model.SystemNotificationTemplate) (err error) {
	res := r.db.WithContext(ctx).Delete(&model.SystemNotificationTemplate{}, "id = ?", dto.ID)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (r *SystemNotificationTemplateRepository) UpdateOrganizations(ctx context.Context, systemNotificationTemplateId uuid.UUID, organizations []model.Organization) (err error) {
	var systemNotificationTemplate = model.SystemNotificationTemplate{}
	res := r.db.WithContext(ctx).Preload("Organizations").First(&systemNotificationTemplate, "id = ?", systemNotificationTemplateId)
	if res.Error != nil {
		return res.Error
	}
	err = r.db.WithContext(ctx).Model(&systemNotificationTemplate).Association("Organizations").Replace(organizations)
	if err != nil {
		return err
	}

	return nil
}
