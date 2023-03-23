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
	Delete(organizationId string) (err error)
	Update(organizationId string, in domain.UpdateOrganizationRequest) (err error)
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

	ID          string `gorm:"primarykey"`
	Name        string
	Description string
	PhoneNumber string
	WorkflowId  string
	Status      domain.OrganizationStatus
	StatusDesc  string
	Creator     uuid.UUID
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
	res := r.db.First(&organization, "id = ?", id)
	if res.RowsAffected == 0 || res.Error != nil {
		return domain.Organization{}, fmt.Errorf("Not found organization for %s", id)
	}
	resOrganization := r.reflect(organization)
	return resOrganization, nil
}

func (r *OrganizationRepository) Create(name string, creator uuid.UUID, description string) (string, error) {
	organization := Organization{Name: name, Creator: creator, Description: description}
	res := r.db.Create(&organization)
	if res.Error != nil {
		return "", res.Error
	}
	return organization.ID, nil
}

func (r *OrganizationRepository) Delete(organizationId string) (err error) {
	res := r.db.Delete(&Organization{}, "id = ?", organizationId)
	if res.Error != nil {
		return fmt.Errorf("could not delete organization for organizationId %s", organizationId)
	}
	return nil
}

func (r *OrganizationRepository) Update(organizationId string, in domain.UpdateOrganizationRequest) (err error) {
	res := r.db.Model(&Organization{}).
		Where("id = ?", organizationId).
		Updates(map[string]interface{}{"Description": in.Description, "PhoneNumber": in.PhoneNumber})
	if res.Error != nil {
		return fmt.Errorf("could not delete organization for organizationId %s", organizationId)
	}
	return nil
}

func (r *OrganizationRepository) InitWorkflow(organizationId string, workflowId string) error {
	res := r.db.Model(&Organization{}).
		Where("ID = ?", organizationId).
		Updates(map[string]interface{}{"Status": domain.OrganizationStatus_PENDING, "WorkflowId": workflowId})

	if res.Error != nil || res.RowsAffected == 0 {
		return fmt.Errorf("nothing updated in organization with id %s", organizationId)
	}
	return nil
}

func (r *OrganizationRepository) reflect(organization Organization) domain.Organization {
	return domain.Organization{
		ID:          organization.ID,
		Name:        organization.Name,
		Description: organization.Description,
		PhoneNumber: organization.PhoneNumber,
		Status:      organization.Status.String(),
		Creator:     organization.Creator.String(),
		CreatedAt:   organization.CreatedAt,
		UpdatedAt:   organization.UpdatedAt,
	}
}
