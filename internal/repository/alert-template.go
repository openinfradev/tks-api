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
type IAlertTemplateRepository interface {
	Get(ctx context.Context, alertTemplateId uuid.UUID) (model.AlertTemplate, error)
	GetByName(ctx context.Context, name string) (model.AlertTemplate, error)
	Fetch(ctx context.Context, pg *pagination.Pagination) ([]model.AlertTemplate, error)
	Create(ctx context.Context, dto model.AlertTemplate) (alertTemplateId uuid.UUID, err error)
	Update(ctx context.Context, dto model.AlertTemplate) (err error)
	Delete(ctx context.Context, dto model.AlertTemplate) (err error)
	UpdateOrganizations(ctx context.Context, alertTemplateId uuid.UUID, organizations []model.Organization) (err error)
}

type AlertTemplateRepository struct {
	db *gorm.DB
}

func NewAlertTemplateRepository(db *gorm.DB) IAlertTemplateRepository {
	return &AlertTemplateRepository{
		db: db,
	}
}

// Logics
func (r *AlertTemplateRepository) Get(ctx context.Context, alertTemplateId uuid.UUID) (out model.AlertTemplate, err error) {
	res := r.db.WithContext(ctx).Preload(clause.Associations).First(&out, "id = ?", alertTemplateId)
	if res.Error != nil {
		return out, res.Error
	}
	return
}

func (r *AlertTemplateRepository) GetByName(ctx context.Context, name string) (out model.AlertTemplate, err error) {
	res := r.db.WithContext(ctx).Preload(clause.Associations).First(&out, "name = ?", name)
	if res.Error != nil {
		return out, res.Error
	}
	return
}

func (r *AlertTemplateRepository) Fetch(ctx context.Context, pg *pagination.Pagination) (out []model.AlertTemplate, err error) {
	if pg == nil {
		pg = pagination.NewPagination(nil)
	}

	_, res := pg.Fetch(r.db.WithContext(ctx).Preload(clause.Associations), &out)
	if res.Error != nil {
		return nil, res.Error
	}
	return
}

func (r *AlertTemplateRepository) Create(ctx context.Context, dto model.AlertTemplate) (alertTemplateId uuid.UUID, err error) {
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

func (r *AlertTemplateRepository) Update(ctx context.Context, dto model.AlertTemplate) (err error) {
	res := r.db.WithContext(ctx).Model(&model.AlertTemplate{}).
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

func (r *AlertTemplateRepository) Delete(ctx context.Context, dto model.AlertTemplate) (err error) {
	res := r.db.WithContext(ctx).Delete(&model.AlertTemplate{}, "id = ?", dto.ID)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (r *AlertTemplateRepository) UpdateOrganizations(ctx context.Context, alertTemplateId uuid.UUID, organizations []model.Organization) (err error) {
	var alertTemplate = model.AlertTemplate{}
	res := r.db.WithContext(ctx).Preload("Organizations").First(&alertTemplate, "id = ?", alertTemplateId)
	if res.Error != nil {
		return res.Error
	}
	err = r.db.WithContext(ctx).Model(&alertTemplate).Association("Organizations").Unscoped().Replace(organizations)
	if err != nil {
		return err
	}

	return nil
}
