package repository

import (
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
	Fetch(clusterId domain.ClusterId, pg *pagination.Pagination) (res []model.AppGroup, err error)
	Get(id domain.AppGroupId) (model.AppGroup, error)
	Create(dto model.AppGroup) (id domain.AppGroupId, err error)
	Update(dto model.AppGroup) (err error)
	Delete(id domain.AppGroupId) error
	GetApplications(id domain.AppGroupId, applicationType domain.ApplicationType) (applications []model.Application, err error)
	UpsertApplication(dto model.Application) error
	InitWorkflow(appGroupId domain.AppGroupId, workflowId string, status domain.AppGroupStatus) error
	InitWorkflowDescription(clusterId domain.ClusterId) error
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
func (r *AppGroupRepository) Fetch(clusterId domain.ClusterId, pg *pagination.Pagination) (out []model.AppGroup, err error) {
	if pg == nil {
		pg = pagination.NewPagination(nil)
	}

	_, res := pg.Fetch(r.db.Model(&model.AppGroup{}).
		Where("cluster_id = ?", clusterId), &out)
	if res.Error != nil {
		return nil, res.Error
	}

	return out, nil
}

func (r *AppGroupRepository) Get(id domain.AppGroupId) (out model.AppGroup, err error) {
	res := r.db.First(&out, "id = ?", id)
	if res.RowsAffected == 0 || res.Error != nil {
		return model.AppGroup{}, fmt.Errorf("Not found appGroup for %s", id)
	}
	return out, nil
}

func (r *AppGroupRepository) Create(dto model.AppGroup) (appGroupId domain.AppGroupId, err error) {
	appGroup := model.AppGroup{
		ClusterId:    dto.ClusterId,
		AppGroupType: dto.AppGroupType,
		Name:         dto.Name,
		Description:  dto.Description,
		Status:       domain.AppGroupStatus_PENDING,
		CreatorId:    dto.CreatorId,
		UpdatorId:    nil,
	}
	res := r.db.Create(&appGroup)
	if res.Error != nil {
		log.Error(res.Error)
		return "", res.Error
	}

	return appGroup.ID, nil
}

func (r *AppGroupRepository) Update(dto model.AppGroup) (err error) {
	res := r.db.Model(&model.AppGroup{}).
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

func (r *AppGroupRepository) Delete(id domain.AppGroupId) error {
	res := r.db.Unscoped().Delete(&model.AppGroup{}, "id = ?", id)
	if res.Error != nil {
		return fmt.Errorf("could not delete appGroup %s", id)
	}
	return nil
}

func (r *AppGroupRepository) GetApplications(id domain.AppGroupId, applicationType domain.ApplicationType) (out []model.Application, err error) {
	res := r.db.Where("app_group_id = ? AND type = ?", id, applicationType).Find(&out)
	if res.Error != nil {
		return nil, res.Error
	}
	return out, nil
}

func (r *AppGroupRepository) UpsertApplication(dto model.Application) error {
	res := r.db.Where(model.Application{
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

func (r *AppGroupRepository) InitWorkflow(appGroupId domain.AppGroupId, workflowId string, status domain.AppGroupStatus) error {
	res := r.db.Model(&model.AppGroup{}).
		Where("ID = ?", appGroupId).
		Updates(map[string]interface{}{"Status": status, "WorkflowId": workflowId, "StatusDesc": ""})

	if res.Error != nil || res.RowsAffected == 0 {
		return fmt.Errorf("nothing updated in appgroup with id %s", appGroupId)
	}

	return nil
}

func (r *AppGroupRepository) InitWorkflowDescription(clusterId domain.ClusterId) error {
	res := r.db.Model(&model.AppGroup{}).
		Where("cluster_id = ?", clusterId).
		Updates(map[string]interface{}{"WorkflowId": "", "StatusDesc": ""})

	if res.Error != nil || res.RowsAffected == 0 {
		return fmt.Errorf("nothing updated in appgroup with id %s", clusterId)
	}

	return nil
}
