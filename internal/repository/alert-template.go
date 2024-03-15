package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/pkg/log"
)

// Interfaces
type IAlertTemplateRepository interface {
	Get(alertTemplateId uuid.UUID) (model.AlertTemplate, error)
	GetByName(name string) (model.AlertTemplate, error)
	Fetch(pg *pagination.Pagination) ([]model.AlertTemplate, error)
	Create(dto model.AlertTemplate) (alertTemplateId uuid.UUID, err error)
	Update(dto model.AlertTemplate) (err error)
	Delete(dto model.AlertTemplate) (err error)
	UpdateOrganizations(alertTemplateId uuid.UUID, organizations []model.Organization) (err error)
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
func (r *AlertTemplateRepository) Get(alertTemplateId uuid.UUID) (out model.AlertTemplate, err error) {
	res := r.db.Preload(clause.Associations).First(&out, "id = ?", alertTemplateId)
	if res.Error != nil {
		return out, res.Error
	}
	return
}

func (r *AlertTemplateRepository) GetByName(name string) (out model.AlertTemplate, err error) {
	res := r.db.Preload(clause.Associations).First(&out, "name = ?", name)
	if res.Error != nil {
		return out, res.Error
	}
	return
}

func (r *AlertTemplateRepository) Fetch(pg *pagination.Pagination) (out []model.AlertTemplate, err error) {
	if pg == nil {
		pg = pagination.NewPagination(nil)
	}

	_, res := pg.Fetch(r.db.Preload(clause.Associations), &out)
	if res.Error != nil {
		return nil, res.Error
	}
	return
}

func (r *AlertTemplateRepository) Create(dto model.AlertTemplate) (alertTemplateId uuid.UUID, err error) {
	res := r.db.Create(&dto)
	if res.Error != nil {
		return uuid.Nil, res.Error
	}

	err = r.db.Model(&dto).Association("MetricParameters").Replace(dto.MetricParameters)
	if err != nil {
		log.Error(err)
	}

	return dto.ID, nil
}

func (r *AlertTemplateRepository) Update(dto model.AlertTemplate) (err error) {
	res := r.db.Model(&model.AlertTemplate{}).
		Where("id = ?", dto.ID).
		Updates(map[string]interface{}{
			"Name":        dto.Name,
			"Description": dto.Description,
			"MetricQuery": dto.MetricQuery,
		})
	if res.Error != nil {
		return res.Error
	}

	if err = r.db.Model(&dto).Association("MetricParameters").Unscoped().Replace(dto.MetricParameters); err != nil {
		return err
	}
	return nil
}

func (r *AlertTemplateRepository) Delete(dto model.AlertTemplate) (err error) {
	res := r.db.Delete(&model.AlertTemplate{}, "id = ?", dto.ID)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (r *AlertTemplateRepository) UpdateOrganizations(alertTemplateId uuid.UUID, organizations []model.Organization) (err error) {
	log.Info("AAA")
	log.Info(organizations)

	var alertTemplate = model.AlertTemplate{}
	res := r.db.Preload("Organizations").First(&alertTemplate, "id = ?", alertTemplateId)
	if res.Error != nil {
		return res.Error
	}
	err = r.db.Model(&alertTemplate).Association("Organizations").Unscoped().Replace(organizations)
	if err != nil {
		return err
	}

	return nil
}
