package repository

import (
	"context"
	"fmt"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
)

// Interfaces
type IAppGroupRepository interface {
	Fetch(ctx context.Context, clusterId domain.ClusterId, pg *pagination.Pagination) (res []model.AppGroup, err error)
	Get(ctx context.Context, id domain.AppGroupId) (model.AppGroup, error)
	Create(ctx context.Context, dto model.AppGroup) (id domain.AppGroupId, err error)
	Update(ctx context.Context, dto model.AppGroup) (err error)
	Delete(ctx context.Context, id domain.AppGroupId) error
	GetApplications(ctx context.Context, id domain.AppGroupId, applicationType domain.ApplicationType) (applications []model.Application, err error)
	UpsertApplication(ctx context.Context, dto model.Application) error
	InitWorkflow(ctx context.Context, appGroupId domain.AppGroupId, workflowId string, status domain.AppGroupStatus) error
	InitWorkflowDescription(ctx context.Context, clusterId domain.ClusterId) error
}

type AppGroupRepository struct {
	db *gorm.DB
}

func NewAppGroupRepository(db *gorm.DB) IAppGroupRepository {
	return &AppGroupRepository{
		db: db,
	}
}

// Logics
func (r *AppGroupRepository) Fetch(ctx context.Context, clusterId domain.ClusterId, pg *pagination.Pagination) (out []model.AppGroup, err error) {
	if pg == nil {
		pg = pagination.NewPagination(nil)
	}

	_, res := pg.Fetch(r.db.WithContext(ctx).Model(&model.AppGroup{}).
		Where("cluster_id = ?", clusterId), &out)
	if res.Error != nil {
		return nil, res.Error
	}

	return out, nil
}

func (r *AppGroupRepository) Get(ctx context.Context, id domain.AppGroupId) (out model.AppGroup, err error) {
	res := r.db.WithContext(ctx).First(&out, "id = ?", id)
	if res.RowsAffected == 0 || res.Error != nil {
		return model.AppGroup{}, fmt.Errorf("Not found appGroup for %s", id)
	}
	return out, nil
}

func (r *AppGroupRepository) Create(ctx context.Context, dto model.AppGroup) (appGroupId domain.AppGroupId, err error) {
	appGroup := model.AppGroup{
		ClusterId:    dto.ClusterId,
		AppGroupType: dto.AppGroupType,
		Name:         dto.Name,
		Description:  dto.Description,
		Status:       domain.AppGroupStatus_PENDING,
		CreatorId:    dto.CreatorId,
		UpdatorId:    nil,
	}
	res := r.db.WithContext(ctx).Create(&appGroup)
	if res.Error != nil {
		log.Error(ctx, res.Error)
		return "", res.Error
	}

	return appGroup.ID, nil
}

func (r *AppGroupRepository) Update(ctx context.Context, dto model.AppGroup) (err error) {
	res := r.db.WithContext(ctx).Model(&model.AppGroup{}).
		Where("id = ?", dto.ID).
		Updates(map[string]interface{}{
			"ClusterId":    dto.ClusterId,
			"AppGroupType": dto.AppGroupType,
			"Name":         dto.Name,
			"Description":  dto.Description,
			"Status":       domain.AppGroupStatus_PENDING,
			"UpdatorId":    dto.UpdatorId})
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (r *AppGroupRepository) Delete(ctx context.Context, id domain.AppGroupId) error {
	res := r.db.WithContext(ctx).Unscoped().Delete(&model.AppGroup{}, "id = ?", id)
	if res.Error != nil {
		return fmt.Errorf("could not delete appGroup %s", id)
	}
	return nil
}

func (r *AppGroupRepository) GetApplications(ctx context.Context, id domain.AppGroupId, applicationType domain.ApplicationType) (out []model.Application, err error) {
	res := r.db.WithContext(ctx).Where("app_group_id = ? AND type = ?", id, applicationType).Find(&out)
	if res.Error != nil {
		return nil, res.Error
	}
	return out, nil
}

func (r *AppGroupRepository) UpsertApplication(ctx context.Context, dto model.Application) error {
	res := r.db.WithContext(ctx).Where(model.Application{
		AppGroupId: dto.AppGroupId,
		Type:       dto.Type,
	}).
		Assign(model.Application{
			Endpoint: dto.Endpoint,
			Metadata: datatypes.JSON([]byte(dto.Metadata))}).
		FirstOrCreate(&model.Application{})
	if res.Error != nil {
		return res.Error
	}

	return nil
}

func (r *AppGroupRepository) InitWorkflow(ctx context.Context, appGroupId domain.AppGroupId, workflowId string, status domain.AppGroupStatus) error {
	res := r.db.WithContext(ctx).Model(&model.AppGroup{}).
		Where("ID = ?", appGroupId).
		Updates(map[string]interface{}{"Status": status, "WorkflowId": workflowId, "StatusDesc": ""})

	if res.Error != nil || res.RowsAffected == 0 {
		return fmt.Errorf("nothing updated in appgroup with id %s", appGroupId)
	}

	return nil
}

func (r *AppGroupRepository) InitWorkflowDescription(ctx context.Context, clusterId domain.ClusterId) error {
	res := r.db.WithContext(ctx).Model(&model.AppGroup{}).
		Where("cluster_id = ?", clusterId).
		Updates(map[string]interface{}{"WorkflowId": "", "StatusDesc": ""})

	if res.Error != nil || res.RowsAffected == 0 {
		return fmt.Errorf("nothing updated in appgroup with id %s", clusterId)
	}

	return nil
}
