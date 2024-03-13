package repository

import (
	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
	"gorm.io/gorm"
)

// Interfaces
type IOrganizationRepository interface {
	Create(organizationId string, name string, creator uuid.UUID, phone string, description string) (model.Organization, error)
	Fetch(pg *pagination.Pagination) (res *[]model.Organization, err error)
	Get(organizationId string) (res model.Organization, err error)
	Update(organizationId string, in domain.UpdateOrganizationRequest) (model.Organization, error)
	UpdatePrimaryClusterId(organizationId string, primaryClusterId string) error
	Delete(organizationId string) (err error)
	InitWorkflow(organizationId string, workflowId string, status domain.OrganizationStatus) error
}

type OrganizationRepository struct {
	db *gorm.DB
}

func NewOrganizationRepository(db *gorm.DB) IOrganizationRepository {
	return &OrganizationRepository{
		db: db,
	}
}

//
//// Models
//type Organization struct {
//	gorm.Model
//
//	ID               string `gorm:"primarykey;type:varchar(36);not null"`
//	Name             string
//	Description      string
//	Phone            string
//	WorkflowId       string
//	Status           model.OrganizationStatus
//	StatusDesc       string
//	Creator          uuid.UUID
//	PrimaryClusterId string // allow null
//}

//func (c *Organization) BeforeCreate(tx *gorm.DB) (err error) {
//	c.ID = helper.GenerateOrganizationId()
//	return nil
//}

func (r *OrganizationRepository) Create(organizationId string, name string, creator uuid.UUID, phone string,
	description string) (model.Organization, error) {
	organization := model.Organization{
		ID:          organizationId,
		Name:        name,
		Creator:     creator.String(),
		Phone:       phone,
		Description: description,
		Status:      domain.OrganizationStatus_PENDING,
	}
	res := r.db.Create(&organization)
	if res.Error != nil {
		log.Errorf("error is :%s(%T)", res.Error.Error(), res.Error)
		return model.Organization{}, res.Error
	}

	return organization, nil
}

func (r *OrganizationRepository) Fetch(pg *pagination.Pagination) (out *[]model.Organization, err error) {
	if pg == nil {
		pg = pagination.NewPagination(nil)
	}

	_, res := pg.Fetch(r.db, &out)
	if res.Error != nil {
		return nil, res.Error
	}
	return
}

func (r *OrganizationRepository) Get(id string) (out model.Organization, err error) {
	res := r.db.First(&out, "id = ?", id)
	if res.Error != nil {
		log.Errorf("error is :%s(%T)", res.Error.Error(), res.Error)
		return model.Organization{}, res.Error
	}

	return
}

func (r *OrganizationRepository) Update(organizationId string, in domain.UpdateOrganizationRequest) (out model.Organization, err error) {
	res := r.db.Model(&model.Organization{}).
		Where("id = ?", organizationId).
		Updates(map[string]interface{}{
			"name":        in.Name,
			"description": in.Description,
			"phone":       in.Phone,
		})

	if res.Error != nil {
		log.Errorf("error is :%s(%T)", res.Error.Error(), res.Error)
		return model.Organization{}, res.Error
	}
	res = r.db.Model(&model.Organization{}).Where("id = ?", organizationId).Find(&out)
	if res.Error != nil {
		log.Errorf("error is :%s(%T)", res.Error.Error(), res.Error)
		return model.Organization{}, res.Error
	}

	return
}

func (r *OrganizationRepository) UpdatePrimaryClusterId(organizationId string, primaryClusterId string) error {
	res := r.db.Model(&model.Organization{}).
		Where("id = ?", organizationId).
		Updates(map[string]interface{}{
			"primary_cluster_id": primaryClusterId,
		})

	if res.Error != nil {
		log.Errorf("error is :%s(%T)", res.Error.Error(), res.Error)
		return res.Error
	}
	return nil
}

func (r *OrganizationRepository) Delete(organizationId string) error {
	res := r.db.Delete(&model.Organization{}, "id = ?", organizationId)
	if res.Error != nil {
		log.Errorf("error is :%s(%T)", res.Error.Error(), res.Error)
		return res.Error
	}

	return nil
}

func (r *OrganizationRepository) InitWorkflow(organizationId string, workflowId string, status domain.OrganizationStatus) error {
	res := r.db.Model(&model.Organization{}).
		Where("ID = ?", organizationId).
		Updates(map[string]interface{}{"Status": status, "WorkflowId": workflowId})
	if res.Error != nil {
		log.Errorf("error is :%s(%T)", res.Error.Error(), res.Error)
		return res.Error
	}
	return nil
}
