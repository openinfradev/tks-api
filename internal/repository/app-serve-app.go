package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
	"gorm.io/gorm"
)

type IAppServeAppRepository interface {
	CreateAppServeApp(ctx context.Context, app *model.AppServeApp) (appId string, err error)
	GetAppServeApps(ctx context.Context, organizationId string, projectId string, showAll bool, pg *pagination.Pagination) ([]model.AppServeApp, error)
	GetAppServeAppById(ctx context.Context, appId string) (*model.AppServeApp, error)

	GetAppServeAppTasksByAppId(ctx context.Context, appId string, pg *pagination.Pagination) ([]model.AppServeAppTask, error)
	GetAppServeAppTaskById(ctx context.Context, taskId string) (*model.AppServeAppTask, error)
	GetAppServeAppLatestTask(ctx context.Context, appId string) (*model.AppServeAppTask, error)

	GetNumOfAppsOnStack(ctx context.Context, organizationId string, clusterId string) (int64, error)

	IsAppServeAppExist(ctx context.Context, appId string) (int64, error)
	IsAppServeAppNameExist(ctx context.Context, orgId string, appName string) (int64, error)
	CreateTask(ctx context.Context, task *model.AppServeAppTask, appId string) (taskId string, err error)
	UpdateStatus(ctx context.Context, appId string, taskId string, status string, output string) error
	UpdateEndpoint(ctx context.Context, appId string, taskId string, endpoint string, previewEndpoint string, helmRevision int32) error
	GetTaskCountById(ctx context.Context, appId string) (int64, error)
}

type AppServeAppRepository struct {
	db *gorm.DB
}

func NewAppServeAppRepository(db *gorm.DB) IAppServeAppRepository {
	return &AppServeAppRepository{
		db: db,
	}
}

func (r *AppServeAppRepository) CreateAppServeApp(ctx context.Context, app *model.AppServeApp) (appId string, err error) {
	app.ID = uuid.New().String()
	res := r.db.WithContext(ctx).Create(&app)
	if res.Error != nil {
		return "", res.Error
	}

	return app.ID, nil
}

// Update creates new appServeApp task for existing appServeApp.
func (r *AppServeAppRepository) CreateTask(ctx context.Context, task *model.AppServeAppTask, appId string) (string, error) {
	task.ID = uuid.New().String()
    if len(appId) > 0 {
        task.AppServeAppId = appId
    }
	res := r.db.WithContext(ctx).Create(task)
	if res.Error != nil {
		return "", res.Error
	}

	return task.ID, nil
}

func (r *AppServeAppRepository) GetAppServeApps(ctx context.Context, organizationId string, projectId string, showAll bool, pg *pagination.Pagination) (apps []model.AppServeApp, err error) {
	var clusters []model.Cluster
	if pg == nil {
		pg = pagination.NewPagination(nil)
	}

	// TODO: should return different records based on showAll param
	_, res := pg.Fetch(r.db.WithContext(ctx).Model(&model.AppServeApp{}).
		Where("app_serve_apps.project_id = ? AND status <> 'DELETE_SUCCESS'", projectId), &apps)
	if res.Error != nil {
		return nil, fmt.Errorf("error while finding appServeApps with projectId: %s", projectId)
	}

	// If no record is found, just return empty array.
	if res.RowsAffected == 0 {
		return apps, nil
	}

	// Add cluster names to apps list
	queryStr := fmt.Sprintf("organization_id = '%s' AND status <> '%d'", organizationId, domain.ClusterStatus_DELETED)
	res = r.db.WithContext(ctx).Find(&clusters, queryStr)
	if res.Error != nil {
		return nil, fmt.Errorf("error while fetching clusters with organizationId: %s", organizationId)
	}

	for idx, app := range apps {
		for _, cl := range clusters {
			if string(cl.ID) == app.TargetClusterId {
				apps[idx].TargetClusterName = cl.Name
				break
			}
		}
	}

	return
}

func (r *AppServeAppRepository) GetAppServeAppById(ctx context.Context, appId string) (*model.AppServeApp, error) {
	var app model.AppServeApp
	var cluster model.Cluster

	res := r.db.WithContext(ctx).Where("id = ?", appId).First(&app)
	if res.Error != nil {
		log.Debug(ctx, res.Error)
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, fmt.Errorf("No app with ID %s", appId)
	}

	// Add cluster name to app object
	r.db.WithContext(ctx).Select("name").Where("id = ?", app.TargetClusterId).First(&cluster)
	app.TargetClusterName = cluster.Name

	return &app, nil
}

func (r *AppServeAppRepository) GetAppServeAppTasksByAppId(ctx context.Context, appId string, pg *pagination.Pagination) (tasks []model.AppServeAppTask, err error) {
	if pg == nil {
		pg = pagination.NewPagination(nil)
	}

	_, res := pg.Fetch(r.db.WithContext(ctx).Model(&model.AppServeAppTask{}).
		Where("app_serve_app_tasks.app_serve_app_id = ?", appId), &tasks)
	if res.Error != nil {
		return nil, fmt.Errorf("Error while finding tasks with appId: %s", appId)
	}

	// If no record is found, just return empty array.
	if res.RowsAffected == 0 {
		return tasks, nil
	}

	return
}

// Return single task info along with its parent app info
func (r *AppServeAppRepository) GetAppServeAppTaskById(ctx context.Context, taskId string) (*model.AppServeAppTask, error) {
	var task model.AppServeAppTask

	// Retrieve task info
	res := r.db.WithContext(ctx).Where("id = ?", taskId).First(&task)
	if res.Error != nil {
		log.Debug(ctx, res.Error)
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, fmt.Errorf("No task with ID %s", taskId)
	}

	return &task, nil
}

func (r *AppServeAppRepository) GetAppServeAppLatestTask(ctx context.Context, appId string) (*model.AppServeAppTask, error) {
	var task model.AppServeAppTask

	res := r.db.WithContext(ctx).Order("created_at desc").Where("app_serve_app_id = ?", appId).First(&task)
	if res.Error != nil {
		log.Debug(ctx, res.Error)
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, fmt.Errorf("No task with App ID %s", appId)
	}

	return &task, nil
}

func (r *AppServeAppRepository) GetNumOfAppsOnStack(ctx context.Context, organizationId string, clusterId string) (int64, error) {
	var apps []model.AppServeApp

	queryStr := fmt.Sprintf("organization_id = '%s' AND target_cluster_id = '%s' AND status <> 'DELETE_SUCCESS'", organizationId, clusterId)
	res := r.db.WithContext(ctx).Find(&apps, queryStr)
	if res.Error != nil {
		return -1, fmt.Errorf("Error while finding appServeApps with organizationId: %s", organizationId)
	}

	return res.RowsAffected, nil
}

func (r *AppServeAppRepository) IsAppServeAppExist(ctx context.Context, appId string) (int64, error) {
	var result int64

	res := r.db.WithContext(ctx).Table("app_serve_apps").Where("id = ? AND status <> 'DELETE_SUCCESS'", appId).Count(&result)
	if res.Error != nil {
		log.Debug(ctx, res.Error)
		return 0, res.Error
	}
	return result, nil
}

func (r *AppServeAppRepository) IsAppServeAppNameExist(ctx context.Context, orgId string, appName string) (int64, error) {
	var result int64

	queryString := fmt.Sprintf("organization_id = '%v' "+
		"AND name = '%v' "+
		"AND status <> 'DELETE_SUCCESS'", orgId, appName)

	log.Info(ctx, "query = ", queryString)
	res := r.db.WithContext(ctx).Table("app_serve_apps").Where(queryString).Count(&result)
	if res.Error != nil {
		log.Debug(ctx, res.Error)
		return 0, res.Error
	}
	return result, nil
}

func (r *AppServeAppRepository) UpdateStatus(ctx context.Context, appId string, taskId string, status string, output string) error {
	now := time.Now()
	app := model.AppServeApp{
		ID:        appId,
		Status:    status,
		UpdatedAt: &now,
	}
	res := r.db.WithContext(ctx).Model(&app).Select("Status", "UpdatedAt").Updates(app)
	if res.Error != nil || res.RowsAffected == 0 {
		return fmt.Errorf("UpdateStatus: nothing updated in AppServeApp with ID %s", appId)
	}

	task := model.AppServeAppTask{
		ID:        taskId,
		Status:    status,
		Output:    output,
		UpdatedAt: &now,
	}
	res = r.db.WithContext(ctx).Model(&task).Select("Status", "Output", "UpdatedAt").Updates(task)
	if res.Error != nil || res.RowsAffected == 0 {
		return fmt.Errorf("UpdateStatus: nothing updated in AppServeAppTask with ID %s", taskId)
	}

	//// Update task status
	//res := r.db.Model(&model.AppServeAppTask{}).
	//	Where("ID = ?", taskId).
	//	Updates(model.AppServeAppTask{Status: status, Output: output})
	//
	//if res.Error != nil || res.RowsAffected == 0 {
	//	return fmt.Errorf("UpdateStatus: nothing updated in AppServeAppTask with ID %s", taskId)
	//}
	//
	//// Update status of the app.
	//res = r.db.Model(&model.AppServeApp{}).
	//	Where("ID = ?", appId).
	//	Update("Status", status)
	//if res.Error != nil || res.RowsAffected == 0 {
	//	return fmt.Errorf("UpdateStatus: nothing updated in AppServeApp with id %s", appId)
	//}

	return nil
}

func (r *AppServeAppRepository) UpdateEndpoint(ctx context.Context, appId string, taskId string, endpoint string, previewEndpoint string, helmRevision int32) error {
	now := time.Now()
	app := model.AppServeApp{
		ID:                 appId,
		EndpointUrl:        endpoint,
		PreviewEndpointUrl: previewEndpoint,
		UpdatedAt:          &now,
	}

	task := model.AppServeAppTask{
		ID:           taskId,
		HelmRevision: helmRevision,
		UpdatedAt:    &now,
	}

	var res *gorm.DB
	if endpoint != "" && previewEndpoint != "" {
		// Both endpoints are valid
		res = r.db.WithContext(ctx).Model(&app).Select("EndpointUrl", "PreviewEndpointUrl", "UpdatedAt").Updates(app)
	} else if endpoint != "" {
		// endpoint-only case
		res = r.db.WithContext(ctx).Model(&app).Select("EndpointUrl", "UpdatedAt").Updates(app)
	} else if previewEndpoint != "" {
		// previewEndpoint-only case
		res = r.db.WithContext(ctx).Model(&app).Select("PreviewEndpointUrl", "UpdatedAt").Updates(app)
	} else {
		return fmt.Errorf("updateEndpoint: No endpoint provided. " +
			"At least one of [endpoint, preview_endpoint] should be provided")
	}
	if res.Error != nil || res.RowsAffected == 0 {
		return fmt.Errorf("UpdateEndpoint: nothing updated in AppServeApp with id %s", appId)
	}

	// Update helm revision
	// Ignore if the value is less than 0
	if helmRevision > 0 {
		res = r.db.WithContext(ctx).Model(&task).Select("HelmRevision", "UpdatedAt").Updates(task)
		if res.Error != nil || res.RowsAffected == 0 {
			return fmt.Errorf("UpdateEndpoint: "+
				"helm revision was not updated for AppServeAppTask with task ID %s", taskId)
		}
	}

	return nil
}

func (r *AppServeAppRepository) GetTaskCountById(ctx context.Context, appId string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.AppServeAppTask{}).Where("AppServeAppId = ?", appId).Count(&count); err != nil {
		return 0, fmt.Errorf("could not select count AppServeAppTask with ID: %s", appId)
	}
	return count, nil
}
