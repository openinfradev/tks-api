package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/openinfradev/tks-api/pkg/domain"
)

// Interfaces
type IStackTemplateRepository interface {
	Get(stackTemplateId uuid.UUID) (domain.StackTemplate, error)
	Fetch(organizationId string) ([]domain.StackTemplate, error)
	Create(dto domain.StackTemplate) (stackTemplateId uuid.UUID, err error)
	Update(dto domain.StackTemplate) (err error)
	Delete(dto domain.StackTemplate) (err error)
}

type StackTemplateRepository struct {
	db *gorm.DB
}

func NewStackTemplateRepository(db *gorm.DB) IStackTemplateRepository {
	return &StackTemplateRepository{
		db: db,
	}
}

// Models
type StackTemplate struct {
	gorm.Model

	ID             uuid.UUID `gorm:"primarykey"`
	OrganizationId string
	Organization   Organization `gorm:"foreignKey:OrganizationId"`
	Name           string
	Description    string
	Version        string
	CloudService   string
	Platform       string
	Template       string
	CreatorId      *uuid.UUID `gorm:"type:uuid"`
	Creator        User       `gorm:"foreignKey:CreatorId"`
	UpdatorId      *uuid.UUID `gorm:"type:uuid"`
	Updator        User       `gorm:"foreignKey:UpdatorId"`
}

func (c *StackTemplate) BeforeCreate(tx *gorm.DB) (err error) {
	c.ID = uuid.New()
	return nil
}

// Logics
func (r *StackTemplateRepository) Get(stackTemplateId uuid.UUID) (out domain.StackTemplate, err error) {
	var stackTemplate StackTemplate
	res := r.db.Preload(clause.Associations).First(&stackTemplate, "id = ?", stackTemplateId)
	if res.Error != nil {
		return domain.StackTemplate{}, res.Error
	}
	out = reflectStackTemplate(stackTemplate)
	return
}

func (r *StackTemplateRepository) Fetch(organizationId string) (out []domain.StackTemplate, err error) {
	var stackTemplates []StackTemplate
	res := r.db.Preload(clause.Associations).Find(&stackTemplates, "organization_id = ?", organizationId)
	if res.Error != nil {
		return nil, res.Error
	}

	for _, stackTemplate := range stackTemplates {
		out = append(out, reflectStackTemplate(stackTemplate))
	}
	return
}

func (r *StackTemplateRepository) Create(dto domain.StackTemplate) (stackTemplateId uuid.UUID, err error) {
	stackTemplate := StackTemplate{
		OrganizationId: dto.OrganizationId,
		Name:           dto.Name,
		Description:    dto.Description,
		CloudService:   dto.CloudService,
		Platform:       dto.Platform,
		Template:       dto.Template,
		CreatorId:      &dto.CreatorId,
		UpdatorId:      nil}
	res := r.db.Create(&stackTemplate)
	if res.Error != nil {
		return uuid.Nil, res.Error
	}
	return stackTemplate.ID, nil
}

func (r *StackTemplateRepository) Update(dto domain.StackTemplate) (err error) {
	res := r.db.Model(&StackTemplate{}).
		Where("id = ?", dto.ID).
		Updates(map[string]interface{}{
			"Description": dto.Description,
			"UpdatorId":   dto.UpdatorId})
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (r *StackTemplateRepository) Delete(dto domain.StackTemplate) (err error) {
	res := r.db.Delete(&StackTemplate{}, "id = ?", dto.ID)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func reflectStackTemplate(stackTemplate StackTemplate) domain.StackTemplate {
	return domain.StackTemplate{
		ID:             stackTemplate.ID,
		OrganizationId: stackTemplate.OrganizationId,
		Name:           stackTemplate.Name,
		Description:    stackTemplate.Description,
		CloudService:   stackTemplate.CloudService,
		Platform:       stackTemplate.Platform,
		Version:        stackTemplate.Version,
		Creator:        reflectUser(stackTemplate.Creator),
		Updator:        reflectUser(stackTemplate.Updator),
		CreatedAt:      stackTemplate.CreatedAt,
		UpdatedAt:      stackTemplate.UpdatedAt,
	}
}
