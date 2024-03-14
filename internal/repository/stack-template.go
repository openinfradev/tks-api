package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
)

// Interfaces
type IStackTemplateRepository interface {
	Get(stackTemplateId uuid.UUID) (model.StackTemplate, error)
	GetByName(name string) (model.StackTemplate, error)
	Fetch(pg *pagination.Pagination) ([]model.StackTemplate, error)
	FetchWithOrganization(organizationId string, pg *pagination.Pagination) (out []model.StackTemplate, err error)
	Create(dto model.StackTemplate) (stackTemplateId uuid.UUID, err error)
	Update(dto model.StackTemplate) (err error)
	Delete(dto model.StackTemplate) (err error)
	UpdateOrganizations(stackTemplateId uuid.UUID, organizationIds []model.Organization) (err error)
}

type StackTemplateRepository struct {
	db *gorm.DB
}

func NewStackTemplateRepository(db *gorm.DB) IStackTemplateRepository {
	return &StackTemplateRepository{
		db: db,
	}
}

// Logics
func (r *StackTemplateRepository) Get(stackTemplateId uuid.UUID) (out model.StackTemplate, err error) {
	res := r.db.Preload(clause.Associations).First(&out, "id = ?", stackTemplateId)
	if res.Error != nil {
		return model.StackTemplate{}, res.Error
	}
	return
}

func (r *StackTemplateRepository) GetByName(name string) (out model.StackTemplate, err error) {
	res := r.db.First(&out, "name = ?", name)
	if res.Error != nil {
		return out, res.Error
	}
	return
}

func (r *StackTemplateRepository) Fetch(pg *pagination.Pagination) (out []model.StackTemplate, err error) {
	if pg == nil {
		pg = pagination.NewPagination(nil)
	}

	_, res := pg.Fetch(r.db.Preload(clause.Associations), &out)
	if res.Error != nil {
		return nil, res.Error
	}
	return
}

func (r *StackTemplateRepository) FetchWithOrganization(organizationId string, pg *pagination.Pagination) (out []model.StackTemplate, err error) {
	if pg == nil {
		pg = pagination.NewPagination(nil)
	}

	_, res := pg.Fetch(
		r.db.Preload(clause.Associations).
			Joins("JOIN stack_template_organizations ON stack_template_organizations.stack_template_id = stack_templates.id AND stack_template_organizations.organization_id = ?", organizationId),
		&out)
	if res.Error != nil {
		return nil, res.Error
	}
	return
}

func (r *StackTemplateRepository) Create(dto model.StackTemplate) (stackTemplateId uuid.UUID, err error) {
	res := r.db.Create(&dto)
	if res.Error != nil {
		return uuid.Nil, res.Error
	}
	return dto.ID, nil
}

func (r *StackTemplateRepository) Update(dto model.StackTemplate) (err error) {
	res := r.db.Model(&model.StackTemplate{}).
		Where("id = ?", dto.ID).
		Updates(map[string]interface{}{
			"Template":     dto.Template,
			"TemplateType": dto.TemplateType,
			"Version":      dto.Version,
			"CloudService": dto.CloudService,
			"Platform":     dto.Platform,
			"KubeVersion":  dto.KubeVersion,
			"KubeType":     dto.KubeType,
			"Services":     dto.Services,
			"Description":  dto.Description,
			"UpdatorId":    dto.UpdatorId})
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (r *StackTemplateRepository) Delete(dto model.StackTemplate) (err error) {
	res := r.db.Delete(&model.StackTemplate{}, "id = ?", dto.ID)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (r *StackTemplateRepository) UpdateOrganizations(stackTemplateId uuid.UUID, organizations []model.Organization) (err error) {
	var stackTemplate = model.StackTemplate{}
	res := r.db.Preload("Organizations").First(&stackTemplate, "id = ?", stackTemplateId)
	if res.Error != nil {
		return res.Error
	}
	err = r.db.Model(&stackTemplate).Association("Organizations").Replace(organizations)
	if err != nil {
		return err
	}

	return nil
}
