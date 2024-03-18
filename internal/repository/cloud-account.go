package repository

import (
	"context"
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
	Get(ctx context.Context, cloudAccountId uuid.UUID) (model.CloudAccount, error)
	GetByName(ctx context.Context, organizationId string, name string) (model.CloudAccount, error)
	GetByAwsAccountId(ctx context.Context, awsAccountId string) (model.CloudAccount, error)
	Fetch(ctx context.Context, organizationId string, pg *pagination.Pagination) ([]model.CloudAccount, error)
	Create(ctx context.Context, dto model.CloudAccount) (cloudAccountId uuid.UUID, err error)
	Update(ctx context.Context, dto model.CloudAccount) (err error)
	Delete(ctx context.Context, cloudAccountId uuid.UUID) (err error)
	InitWorkflow(ctx context.Context, cloudAccountId uuid.UUID, workflowId string, status domain.CloudAccountStatus) (err error)
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
func (r *CloudAccountRepository) Get(ctx context.Context, cloudAccountId uuid.UUID) (out model.CloudAccount, err error) {
	res := r.db.WithContext(ctx).Preload(clause.Associations).First(&out, "id = ?", cloudAccountId)
	if res.Error != nil {
		return model.CloudAccount{}, res.Error
	}
	return
}

func (r *CloudAccountRepository) GetByName(ctx context.Context, organizationId string, name string) (out model.CloudAccount, err error) {
	res := r.db.WithContext(ctx).Preload(clause.Associations).First(&out, "organization_id = ? AND name = ?", organizationId, name)
	if res.Error != nil {
		return model.CloudAccount{}, res.Error
	}
	return
}

func (r *CloudAccountRepository) GetByAwsAccountId(ctx context.Context, awsAccountId string) (out model.CloudAccount, err error) {
	res := r.db.WithContext(ctx).Preload(clause.Associations).First(&out, "aws_account_id = ? AND status != ?", awsAccountId, domain.CloudAccountStatus_DELETED)
	if res.Error != nil {
		return model.CloudAccount{}, res.Error
	}
	return
}

func (r *CloudAccountRepository) Fetch(ctx context.Context, organizationId string, pg *pagination.Pagination) (out []model.CloudAccount, err error) {
	if pg == nil {
		pg = pagination.NewPagination(nil)
	}
	_, res := pg.Fetch(r.db.WithContext(ctx).Model(&model.CloudAccount{}).
		Preload(clause.Associations).
		Where("organization_id = ? AND status != ?", organizationId, domain.CloudAccountStatus_DELETED), &out)
	if res.Error != nil {
		return nil, res.Error
	}
	return
}

func (r *CloudAccountRepository) Create(ctx context.Context, dto model.CloudAccount) (cloudAccountId uuid.UUID, err error) {
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
	res := r.db.WithContext(ctx).Create(&cloudAccount)
	if res.Error != nil {
		return uuid.Nil, res.Error
	}
	return cloudAccount.ID, nil
}

func (r *CloudAccountRepository) Update(ctx context.Context, dto model.CloudAccount) (err error) {
	res := r.db.WithContext(ctx).Model(&model.CloudAccount{}).
		Where("id = ?", dto.ID).
		Updates(map[string]interface{}{"Description": dto.Description, "Resource": dto.Resource, "UpdatorId": dto.UpdatorId})
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (r *CloudAccountRepository) Delete(ctx context.Context, cloudAccountId uuid.UUID) (err error) {
	res := r.db.WithContext(ctx).Delete(&model.CloudAccount{}, "id = ?", cloudAccountId)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (r *CloudAccountRepository) InitWorkflow(ctx context.Context, cloudAccountId uuid.UUID, workflowId string, status domain.CloudAccountStatus) error {
	res := r.db.WithContext(ctx).Model(&model.CloudAccount{}).
		Where("ID = ?", cloudAccountId).
		Updates(map[string]interface{}{"Status": status, "WorkflowId": workflowId})

	if res.Error != nil || res.RowsAffected == 0 {
		return fmt.Errorf("nothing updated in cloud-account with id %s", &cloudAccountId)
	}

	return nil
}
