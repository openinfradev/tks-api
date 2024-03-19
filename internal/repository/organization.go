package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Interfaces
type IOrganizationRepository interface {
	Create(ctx context.Context, dto *model.Organization) (model.Organization, error)
	Fetch(ctx context.Context, pg *pagination.Pagination) (res *[]model.Organization, err error)
	Get(ctx context.Context, organizationId string) (res model.Organization, err error)
	Update(ctx context.Context, organizationId string, in model.Organization) (model.Organization, error)
	UpdatePrimaryClusterId(ctx context.Context, organizationId string, primaryClusterId string) error
	UpdateAdminId(ctx context.Context, organizationId string, adminId uuid.UUID) error
	UpdateStackTemplates(ctx context.Context, organizationId string, stackTemplates []model.StackTemplate) (err error)
	UpdatePolicyTemplates(ctx context.Context, organizationId string, policyTemplates []model.PolicyTemplate) (err error)
	Delete(ctx context.Context, organizationId string) (err error)
	InitWorkflow(ctx context.Context, organizationId string, workflowId string, status domain.OrganizationStatus) error
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

func (r *OrganizationRepository) Create(ctx context.Context, dto *model.Organization) (model.Organization, error) {
	organization := model.Organization{
		ID:          dto.ID,
		Name:        dto.Name,
		CreatorId:   dto.CreatorId,
		Description: dto.Description,
		Status:      domain.OrganizationStatus_PENDING,
		Phone:       dto.Phone,
	}
	res := r.db.WithContext(ctx).Create(&organization)
	if res.Error != nil {
		log.Errorf(ctx, "error is :%s(%T)", res.Error.Error(), res.Error)
		return model.Organization{}, res.Error
	}

	return organization, nil
}

func (r *OrganizationRepository) Fetch(ctx context.Context, pg *pagination.Pagination) (out *[]model.Organization, err error) {
	if pg == nil {
		pg = pagination.NewPagination(nil)
	}

	_, res := pg.Fetch(r.db.WithContext(ctx).Preload(clause.Associations), &out)
	if res.Error != nil {
		return nil, res.Error
	}
	return
}

func (r *OrganizationRepository) Get(ctx context.Context, id string) (out model.Organization, err error) {
	res := r.db.WithContext(ctx).Preload(clause.Associations).
		First(&out, "id = ?", id)
	if res.Error != nil {
		log.Errorf(ctx, "error is :%s(%T)", res.Error.Error(), res.Error)
		return model.Organization{}, res.Error
	}
	return
}

func (r *OrganizationRepository) Update(ctx context.Context, organizationId string, in model.Organization) (out model.Organization, err error) {
	res := r.db.WithContext(ctx).Model(&model.Organization{}).
		Where("id = ?", organizationId).
		Updates(map[string]interface{}{
			"name":        in.Name,
			"description": in.Description,
		})

	if res.Error != nil {
		log.Errorf(ctx, "error is :%s(%T)", res.Error.Error(), res.Error)
		return model.Organization{}, res.Error
	}
	res = r.db.Model(&model.Organization{}).Where("id = ?", organizationId).Find(&out)
	if res.Error != nil {
		log.Errorf(ctx, "error is :%s(%T)", res.Error.Error(), res.Error)
		return model.Organization{}, res.Error
	}
	return
}

func (r *OrganizationRepository) UpdatePrimaryClusterId(ctx context.Context, organizationId string, primaryClusterId string) error {
	res := r.db.WithContext(ctx).Model(&model.Organization{}).
		Where("id = ?", organizationId).
		Updates(map[string]interface{}{
			"primary_cluster_id": primaryClusterId,
		})

	if res.Error != nil {
		log.Errorf(ctx, "error is :%s(%T)", res.Error.Error(), res.Error)
		return res.Error
	}
	return nil
}

func (r *OrganizationRepository) UpdateAdminId(ctx context.Context, organizationId string, adminId uuid.UUID) (err error) {
	res := r.db.WithContext(ctx).Model(&model.Organization{}).
		Where("id = ?", organizationId).
		Updates(map[string]interface{}{
			"admin_id": adminId,
		})

	if res.Error != nil {
		log.Errorf(ctx, "error is :%s(%T)", res.Error.Error(), res.Error)
		return res.Error
	}
	return nil
}

func (r *OrganizationRepository) Delete(ctx context.Context, organizationId string) error {
	res := r.db.WithContext(ctx).Delete(&model.Organization{}, "id = ?", organizationId)
	if res.Error != nil {
		log.Errorf(ctx, "error is :%s(%T)", res.Error.Error(), res.Error)
		return res.Error
	}

	return nil
}

func (r *OrganizationRepository) InitWorkflow(ctx context.Context, organizationId string, workflowId string, status domain.OrganizationStatus) error {
	res := r.db.WithContext(ctx).Model(&model.Organization{}).
		Where("ID = ?", organizationId).
		Updates(map[string]interface{}{"Status": status, "WorkflowId": workflowId})
	if res.Error != nil {
		log.Errorf(ctx, "error is :%s(%T)", res.Error.Error(), res.Error)
		return res.Error
	}
	return nil
}

func (r *OrganizationRepository) UpdateStackTemplates(ctx context.Context, organizationId string, stackTemplates []model.StackTemplate) (err error) {
	var organization = model.Organization{}
	res := r.db.WithContext(ctx).Preload("StackTemplates").First(&organization, "id = ?", organizationId)
	if res.Error != nil {
		return res.Error
	}

	err = r.db.WithContext(ctx).Model(&organization).Association("StackTemplates").Replace(stackTemplates)
	if err != nil {
		return err
	}

	return nil
}

func (r *OrganizationRepository) UpdatePolicyTemplates(ctx context.Context, organizationId string, policyTemplates []model.PolicyTemplate) (err error) {
	// [TODO]

	return nil
}
