package repository

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/serializer"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
)

// Interfaces
type IStackTemplateRepository interface {
	Get(stackTemplateId uuid.UUID) (domain.StackTemplate, error)
	Fetch(pg *pagination.Pagination) ([]domain.StackTemplate, error)
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
	Organization   domain.Organization `gorm:"foreignKey:OrganizationId"`
	Name           string              `gorm:"index"`
	Description    string              `gorm:"index"`
	Template       string
	TemplateType   string
	Version        string
	CloudService   string
	Platform       string
	KubeVersion    string
	KubeType       string
	Services       datatypes.JSON
	CreatorId      *uuid.UUID  `gorm:"type:uuid"`
	Creator        domain.User `gorm:"foreignKey:CreatorId"`
	UpdatorId      *uuid.UUID  `gorm:"type:uuid"`
	Updator        domain.User `gorm:"foreignKey:UpdatorId"`
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

// [TODO] organizationId 별로 생성하지 않고, 하나의 stackTemplate 을 모든 organization 에서 재사용한다. ( 5월 한정, 추후 rearchitecture 필요)
func (r *StackTemplateRepository) Fetch(pg *pagination.Pagination) (out []domain.StackTemplate, err error) {
	var stackTemplates []StackTemplate

	if pg == nil {
		pg = pagination.NewPagination(nil)
	}

	_, res := pg.Fetch(r.db, &stackTemplates)
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
		TemplateType:   dto.TemplateType,
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

func reflectStackTemplate(stackTemplate StackTemplate) (out domain.StackTemplate) {
	if err := serializer.Map(stackTemplate.Model, &out); err != nil {
		log.Error(err)
	}
	if err := serializer.Map(stackTemplate, &out); err != nil {
		log.Error(err)
	}
	out.Services = stackTemplate.Services
	return
}
