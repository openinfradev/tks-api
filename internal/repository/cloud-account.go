package repository

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/pkg/domain"
)

// Interfaces
type ICloudAccountRepository interface {
	Get(cloudAccountId uuid.UUID) (model.CloudAccount, error)
	GetByName(organizationId string, name string) (model.CloudAccount, error)
	GetByAwsAccountId(awsAccountId string) (model.CloudAccount, error)
	Fetch(organizationId string, pg *pagination.Pagination) ([]model.CloudAccount, error)
	Create(dto model.CloudAccount) (cloudAccountId uuid.UUID, err error)
	Update(dto model.CloudAccount) (err error)
	Delete(cloudAccountId uuid.UUID) (err error)
	InitWorkflow(cloudAccountId uuid.UUID, workflowId string, status domain.CloudAccountStatus) (err error)
}

type CloudAccountRepository struct {
	db *gorm.DB
}

func NewCloudAccountRepository(db *gorm.DB) ICloudAccountRepository {
	return &CloudAccountRepository{
		db: db,
	}
}

// Logics
func (r *CloudAccountRepository) Get(cloudAccountId uuid.UUID) (out model.CloudAccount, err error) {
	res := r.db.Preload(clause.Associations).First(&out, "id = ?", cloudAccountId)
	if res.Error != nil {
		return model.CloudAccount{}, res.Error
	}
	return
}

func (r *CloudAccountRepository) GetByName(organizationId string, name string) (out model.CloudAccount, err error) {
	res := r.db.Preload(clause.Associations).First(&out, "organization_id = ? AND name = ?", organizationId, name)
	if res.Error != nil {
		return model.CloudAccount{}, res.Error
	}
	return
}

func (r *CloudAccountRepository) GetByAwsAccountId(awsAccountId string) (out model.CloudAccount, err error) {
	res := r.db.Preload(clause.Associations).First(&out, "aws_account_id = ? AND status != ?", awsAccountId, domain.CloudAccountStatus_DELETED)
	if res.Error != nil {
		return model.CloudAccount{}, res.Error
	}
	return
}

func (r *CloudAccountRepository) Fetch(organizationId string, pg *pagination.Pagination) (out []model.CloudAccount, err error) {
	if pg == nil {
		pg = pagination.NewPagination(nil)
	}
	_, res := pg.Fetch(r.db.Model(&model.CloudAccount{}).
		Preload(clause.Associations).
		Where("organization_id = ? AND status != ?", organizationId, domain.CloudAccountStatus_DELETED), &out)
	if res.Error != nil {
		return nil, res.Error
	}
	return
}

func (r *CloudAccountRepository) Create(dto model.CloudAccount) (cloudAccountId uuid.UUID, err error) {
	cloudAccount := model.CloudAccount{
		OrganizationId: dto.OrganizationId,
		Name:           dto.Name,
		Description:    dto.Description,
		CloudService:   dto.CloudService,
		Resource:       dto.Resource,
		AwsAccountId:   dto.AwsAccountId,
		CreatedIAM:     false,
		Status:         domain.CloudAccountStatus_PENDING,
		CreatorId:      dto.CreatorId}
	res := r.db.Create(&cloudAccount)
	if res.Error != nil {
		return uuid.Nil, res.Error
	}
	return cloudAccount.ID, nil
}

func (r *CloudAccountRepository) Update(dto model.CloudAccount) (err error) {
	res := r.db.Model(&model.CloudAccount{}).
		Where("id = ?", dto.ID).
		Updates(map[string]interface{}{"Description": dto.Description, "Resource": dto.Resource, "UpdatorId": dto.UpdatorId})
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (r *CloudAccountRepository) Delete(cloudAccountId uuid.UUID) (err error) {
	res := r.db.Delete(&model.CloudAccount{}, "id = ?", cloudAccountId)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (r *CloudAccountRepository) InitWorkflow(cloudAccountId uuid.UUID, workflowId string, status domain.CloudAccountStatus) error {
	res := r.db.Model(&model.CloudAccount{}).
		Where("ID = ?", cloudAccountId).
		Updates(map[string]interface{}{"Status": status, "WorkflowId": workflowId})

	if res.Error != nil || res.RowsAffected == 0 {
		return fmt.Errorf("nothing updated in cloud-account with id %s", &cloudAccountId)
	}

	return nil
}
