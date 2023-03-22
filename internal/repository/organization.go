package repository

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/openinfradev/tks-api/internal/helper"
	"github.com/openinfradev/tks-api/pkg/domain"
)

// Interfaces
type IOrganizationRepository interface {
	Fetch() (res []domain.Organization, err error)
	Get(organizationId string) (res domain.Organization, err error)
	Create(name string, creator uuid.UUID, description string) (string, error)
	Delete(organizationId string) (out string, err error)
	InitWorkflow(organizationId string, workflowId string) error
}

type OrganizationRepository struct {
	db *gorm.DB
}

func NewOrganizationRepository(db *gorm.DB) IOrganizationRepository {
	return &OrganizationRepository{
		db: db,
	}
}

// Models
type Organization struct {
	gorm.Model

	ID          string `gorm:"primarykey;type:varchar(36);not null"`
	Name        string
	Creator     uuid.UUID
	Description string
	Workflow    Workflow `gorm:"polymorphic:Ref;polymorphicValue:organization"`
	Status      string
}

type OrganizationUser struct {
	gorm.Model

	OrganizationId string
	Organization   Organization
	UserId         uuid.UUID
	User           User
	Creator        uuid.UUID
	Description    string
}

func (c *Organization) BeforeCreate(tx *gorm.DB) (err error) {
	c.ID = helper.GenerateOrganizationId()
	return nil
}

// Logics
func (r *OrganizationRepository) Fetch() (out []domain.Organization, err error) {
	var organizations []Organization
	out = []domain.Organization{}

	res := r.db.Find(&organizations)
	if res.Error != nil {
		return nil, res.Error
	}
	for _, organization := range organizations {
		outOrganization := r.reflect(organization)
		out = append(out, outOrganization)
	}
	return out, nil
}

func (r *OrganizationRepository) Get(id string) (domain.Organization, error) {
	var organization Organization
	res := r.db.Preload("Workflow").First(&organization, "id = ?", id)
	if res.RowsAffected == 0 || res.Error != nil {
		return domain.Organization{}, fmt.Errorf("Not found organization for %s", id)
	}
	resOrganization := r.reflect(organization)
	return resOrganization, nil
}

func (r *OrganizationRepository) Create(name string, creator uuid.UUID, description string) (string, error) {
	organization := Organization{Name: name, Creator: creator, Description: description}
	err := r.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Create(&organization)
		if res.Error != nil {
			return res.Error
		}
		return nil
	})

	return organization.ID, err
}

func (r *OrganizationRepository) Delete(organizationId string) (out string, err error) {
	err = r.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Delete(&Organization{}, "id = ?", organizationId)
		if res.Error != nil {
			return fmt.Errorf("could not delete organization for organizationId %s", organizationId)
		}
		return nil
	})
	return "Delete organization successfuly", nil
}

func (r *OrganizationRepository) InitWorkflow(organizationId string, workflowId string) error {
	workflow := Workflow{
		RefID:      organizationId,
		RefType:    "organization",
		WorkflowId: workflowId,
		StatusDesc: "INIT",
	}
	res := r.db.Create(&workflow)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (r *OrganizationRepository) reflect(organization Organization) domain.Organization {
	return domain.Organization{
		ID:          organization.ID,
		Name:        organization.Name,
		Description: organization.Description,
		Status:      organization.Status,
		Creator:     organization.Creator.String(),
		CreatedAt:   organization.CreatedAt,
		UpdatedAt:   organization.UpdatedAt,
	}
}
