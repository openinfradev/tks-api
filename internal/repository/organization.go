package repository

import (
	"context"
	"fmt"

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
	AddStackTemplates(ctx context.Context, organizationId string, stackTemplates []model.StackTemplate) (err error)
	RemoveStackTemplates(ctx context.Context, organizationId string, stackTemplates []model.StackTemplate) (err error)
	AddSystemNotificationTemplates(ctx context.Context, organizationId string, systemNotificationTemplates []model.SystemNotificationTemplate) (err error)
	RemoveSystemNotificationTemplates(ctx context.Context, organizationId string, systemNotificationTemplates []model.SystemNotificationTemplate) (err error)
	Delete(ctx context.Context, organizationId string) (err error)
	InitWorkflow(ctx context.Context, organizationId string, workflowId string, status domain.OrganizationStatus) error
	AddPermittedPolicyTemplatesByID(ctx context.Context, organizationId string, policyTemplates []model.PolicyTemplate) (err error)
	UpdatePermittedPolicyTemplatesByID(ctx context.Context, organizationId string, policyTemplates []model.PolicyTemplate) (err error)
	DeletePermittedPolicyTemplatesByID(ctx context.Context, organizationId string, policyTemplateids []uuid.UUID) (err error)
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

	db := r.db.WithContext(ctx).Preload(clause.Associations).Model(&model.Organization{})

	// [TODO] more pretty!
	adminQuery := ""
	for _, filter := range pg.Filters {
		if filter.Relation == "Admin" {
			if adminQuery != "" {
				adminQuery = adminQuery + " OR "
			}

			switch filter.Column {
			case "name":
				adminQuery = adminQuery + fmt.Sprintf("users.name ilike '%%%s%%'", filter.Values[0])
			case "account_id":
				adminQuery = adminQuery + fmt.Sprintf("users.account_id ilike '%%%s%%'", filter.Values[0])
			case "email":
				adminQuery = adminQuery + fmt.Sprintf("users.email ilike '%%%s%%'", filter.Values[0])
			}
		}
	}
	db = db.Joins("join users on users.id::text = organizations.admin_id::text").
		Where(adminQuery)

	_, res := pg.Fetch(db, &out)
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

func (r *OrganizationRepository) AddStackTemplates(ctx context.Context, organizationId string, stackTemplates []model.StackTemplate) (err error) {
	var organization = model.Organization{}
	res := r.db.WithContext(ctx).Preload("StackTemplates").First(&organization, "id = ?", organizationId)
	if res.Error != nil {
		return res.Error
	}

	err = r.db.WithContext(ctx).Model(&organization).Association("StackTemplates").Append(stackTemplates)
	if err != nil {
		return err
	}

	return nil
}

func (r *OrganizationRepository) RemoveStackTemplates(ctx context.Context, organizationId string, stackTemplates []model.StackTemplate) (err error) {
	var organization = model.Organization{}
	res := r.db.WithContext(ctx).Preload("StackTemplates").First(&organization, "id = ?", organizationId)
	if res.Error != nil {
		return res.Error
	}

	err = r.db.WithContext(ctx).Model(&organization).Association("StackTemplates").Delete(stackTemplates)
	if err != nil {
		return err
	}

	return nil
}

func (r *OrganizationRepository) AddSystemNotificationTemplates(ctx context.Context, organizationId string, templates []model.SystemNotificationTemplate) (err error) {
	var organization = model.Organization{}
	res := r.db.WithContext(ctx).Preload("SystemNotificationTemplates").First(&organization, "id = ?", organizationId)
	if res.Error != nil {
		return res.Error
	}

	err = r.db.WithContext(ctx).Model(&organization).Association("SystemNotificationTemplates").Append(templates)
	if err != nil {
		return err
	}

	return nil
}

func (r *OrganizationRepository) RemoveSystemNotificationTemplates(ctx context.Context, organizationId string, templates []model.SystemNotificationTemplate) (err error) {
	var organization = model.Organization{}
	res := r.db.WithContext(ctx).Preload("SystemNotificationTemplates").First(&organization, "id = ?", organizationId)
	if res.Error != nil {
		return res.Error
	}

	err = r.db.WithContext(ctx).Model(&organization).Association("SystemNotificationTemplates").Delete(templates)
	if err != nil {
		return err
	}

	return nil
}

func (r *OrganizationRepository) AddPermittedPolicyTemplatesByID(ctx context.Context, organizationId string, policyTemplates []model.PolicyTemplate) (err error) {
	var organization model.Organization
	organization.ID = organizationId

	err = r.db.WithContext(ctx).Model(&organization).
		Association("PolicyTemplates").Append(policyTemplates)

	return err
}

func (r *OrganizationRepository) UpdatePermittedPolicyTemplatesByID(ctx context.Context, organizationId string, policyTemplates []model.PolicyTemplate) (err error) {
	var organization model.Organization
	organization.ID = organizationId

	err = r.db.WithContext(ctx).Model(&organization).
		Association("PolicyTemplates").Replace(policyTemplates)

	return err
}

func (r *OrganizationRepository) DeletePermittedPolicyTemplatesByID(ctx context.Context, organizationId string, policyTemplateids []uuid.UUID) (err error) {
	return r.db.WithContext(ctx).
		Where("organization_id = ?", organizationId).
		Where("policy_template_id in ?", policyTemplateids).
		Delete(&model.PolicyTemplatePermittedOrganization{}).Error
}
