package repository

import (
	"fmt"
	"math"
	"time"

	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
	"gorm.io/gorm"
)

type IAppServeAppRepository interface {
	CreateAppServeApp(app *domain.AppServeApp) (appId string, taskId string, err error)
	GetAppServeApps(organizationId string, showAll bool, pg *pagination.Pagination) ([]domain.AppServeApp, error)
	GetAppServeAppById(appId string) (*domain.AppServeApp, error)

	GetAppServeAppTasksByAppId(appId string, pg *pagination.Pagination) ([]domain.AppServeAppTask, error)
	GetAppServeAppTaskById(taskId string) (*domain.AppServeAppTask, *domain.AppServeApp, error)

	GetAppServeAppLatestTask(appId string) (*domain.AppServeAppTask, error)
	GetNumOfAppsOnStack(organizationId string, clusterId string) (int64, error)

	IsAppServeAppExist(appId string) (int64, error)
	IsAppServeAppNameExist(orgId string, appName string) (int64, error)
	CreateTask(task *domain.AppServeAppTask) (taskId string, err error)
	UpdateStatus(appId string, taskId string, status string, output string) error
	UpdateEndpoint(appId string, taskId string, endpoint string, previewEndpoint string, helmRevision int32) error
	GetTaskCountById(appId string) (int64, error)
}

type AppServeAppRepository struct {
	db *gorm.DB
}

func NewAppServeAppRepository(db *gorm.DB) IAppServeAppRepository {
	return &AppServeAppRepository{
		db: db,
	}
}

func (r *AppServeAppRepository) CreateAppServeApp(app *domain.AppServeApp) (appId string, taskId string, err error) {

	res := r.db.Create(&app)
	if res.Error != nil {
		return "", "", res.Error
	}

	return app.ID, app.AppServeAppTasks[0].ID, nil
}

// Update creates new appServeApp task for existing appServeApp.
func (r *AppServeAppRepository) CreateTask(
	task *domain.AppServeAppTask) (string, error) {
	res := r.db.Create(task)
	if res.Error != nil {
		return "", res.Error
	}

	return task.ID, nil
}

func (r *AppServeAppRepository) GetAppServeApps(organizationId string, showAll bool, pg *pagination.Pagination) (apps []domain.AppServeApp, err error) {
	var clusters []Cluster
	if pg == nil {
		pg = pagination.NewDefaultPagination()
	}

	filterFunc := CombinedGormFilter("app_serve_apps", pg.GetFilters(), pg.CombinedFilter)
	db := filterFunc(r.db.Model(&domain.AppServeApp{}).
		Where("app_serve_apps.organization_id = ? AND status <> 'DELETE_SUCCESS'", organizationId))
	db.Count(&pg.TotalRows)

	pg.TotalPages = int(math.Ceil(float64(pg.TotalRows) / float64(pg.Limit)))
	orderQuery := fmt.Sprintf("%s %s", pg.SortColumn, pg.SortOrder)
	res := db.Offset(pg.GetOffset()).Limit(pg.GetLimit()).Order(orderQuery).Find(&apps)
	if res.Error != nil {
		return nil, fmt.Errorf("error while finding appServeApps with organizationId: %s", organizationId)
	}

	// If no record is found, just return empty array.
	if res.RowsAffected == 0 {
		return apps, nil
	}

	// Add cluster names to apps list
	queryStr := fmt.Sprintf("organization_id = '%s' AND status <> '%d'", organizationId, domain.ClusterStatus_DELETED)
	res = r.db.Find(&clusters, queryStr)
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

func (r *AppServeAppRepository) GetAppServeAppById(appId string) (*domain.AppServeApp, error) {
	var app domain.AppServeApp
	var cluster Cluster

	res := r.db.Where("id = ?", appId).First(&app)
	if res.Error != nil {
		log.Debug(res.Error)
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, nil, fmt.Errorf("No app with ID %s", appId)
	}

	// Populate tasks into app object
	if err := r.db.Model(&app).Order("created_at desc").Association("AppServeAppTasks").Find(&app.AppServeAppTasks); err != nil {
		log.Debug(err)
		return nil, err
	}

	// Add cluster name to app object
	r.db.Select("name").Where("id = ?", app.TargetClusterId).First(&cluster)
	app.TargetClusterName = cluster.Name

	return &app, nil
}

func (r *AppServeAppRepository) GetAppServeAppTasksByAppId(appId string, pg *pagination.Pagination) (tasks []domain.AppServeAppTask, err error) {
	if pg == nil {
		pg = pagination.NewDefaultPagination()
	}

	filterFunc := CombinedGormFilter("app_serve_app_tasks", pg.GetFilters(), pg.CombinedFilter)
	db := filterFunc(r.db.Model(&domain.AppServeAppTask{}).
		Where("app_serve_app_tasks.app_serve_app_id = ?", appId))
	db.Count(&pg.TotalRows)

	pg.TotalPages = int(math.Ceil(float64(pg.TotalRows) / float64(pg.Limit)))
	orderQuery := fmt.Sprintf("%s %s", pg.SortColumn, pg.SortOrder)
	res := db.Offset(pg.GetOffset()).Limit(pg.GetLimit()).Order(orderQuery).Find(&tasks)
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
func (r *AppServeAppRepository) GetAppServeAppTaskById(taskId string) (*domain.AppServeAppTask, *domain.AppServeApp, error) {
	var task domain.AppServeAppTask
	var app domain.AppServeApp

	// Retrieve task info
	res := r.db.Where("id = ?", taskId).First(&task)
	if res.Error != nil {
		log.Debug(res.Error)
		return nil, nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, nil, fmt.Errorf("No task with ID %s", taskId)
	}

	// Retrieve app info
	res = r.db.Where("id = ?", task.AppServeAppId).First(&app)
	if res.Error != nil {
		log.Debug(res.Error)
		return nil, nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, nil, fmt.Errorf("Couldn't find app with ID %s associated with task %s", appId, taskId)
	}

	return &task, &app, nil
}

func (r *AppServeAppRepository) GetAppServeAppLatestTask(appId string) (*domain.AppServeAppTask, error) {
	var task domain.AppServeAppTask

	// TODO: Does this work?? where's app ID here?
	res := r.db.Order("created_at desc").First(&task)
	if res.Error != nil {
		log.Debug(res.Error)
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, nil
	}

	return &task, nil
}

func (r *AppServeAppRepository) GetNumOfAppsOnStack(organizationId string, clusterId string) (int64, error) {
	var apps []domain.AppServeApp

	queryStr := fmt.Sprintf("organization_id = '%s' AND target_cluster_id = '%s' AND status <> 'DELETE_SUCCESS'", organizationId, clusterId)
	res := r.db.Find(&apps, queryStr)
	if res.Error != nil {
		return -1, fmt.Errorf("Error while finding appServeApps with organizationId: %s", organizationId)
	}

	return res.RowsAffected, nil
}

func (r *AppServeAppRepository) IsAppServeAppExist(appId string) (int64, error) {
	var result int64

	res := r.db.Table("app_serve_apps").Where("id = ? AND status <> 'DELETE_SUCCESS'", appId).Count(&result)
	if res.Error != nil {
		log.Debug(res.Error)
		return 0, res.Error
	}
	return result, nil
}

func (r *AppServeAppRepository) IsAppServeAppNameExist(orgId string, appName string) (int64, error) {
	var result int64

	queryString := fmt.Sprintf("organization_id = '%v' "+
		"AND name = '%v' "+
		"AND status <> 'DELETE_SUCCESS'", orgId, appName)

	log.Info("query = ", queryString)
	res := r.db.Table("app_serve_apps").Where(queryString).Count(&result)
	if res.Error != nil {
		log.Debug(res.Error)
		return 0, res.Error
	}
	return result, nil
}

func (r *AppServeAppRepository) UpdateStatus(appId string, taskId string, status string, output string) error {
	now := time.Now()
	app := domain.AppServeApp{
		ID:        appId,
		Status:    status,
		UpdatedAt: &now,
	}
	res := r.db.Model(&app).Select("Status", "UpdatedAt").Updates(app)
	if res.Error != nil || res.RowsAffected == 0 {
		return fmt.Errorf("UpdateStatus: nothing updated in AppServeApp with ID %s", appId)
	}

	task := domain.AppServeAppTask{
		ID:        taskId,
		Status:    status,
		Output:    output,
		UpdatedAt: &now,
	}
	res = r.db.Model(&task).Select("Status", "Output", "UpdatedAt").Updates(task)
	if res.Error != nil || res.RowsAffected == 0 {
		return fmt.Errorf("UpdateStatus: nothing updated in AppServeAppTask with ID %s", taskId)
	}

	//// Update task status
	//res := r.db.Model(&domain.AppServeAppTask{}).
	//	Where("ID = ?", taskId).
	//	Updates(domain.AppServeAppTask{Status: status, Output: output})
	//
	//if res.Error != nil || res.RowsAffected == 0 {
	//	return fmt.Errorf("UpdateStatus: nothing updated in AppServeAppTask with ID %s", taskId)
	//}
	//
	//// Update status of the app.
	//res = r.db.Model(&domain.AppServeApp{}).
	//	Where("ID = ?", appId).
	//	Update("Status", status)
	//if res.Error != nil || res.RowsAffected == 0 {
	//	return fmt.Errorf("UpdateStatus: nothing updated in AppServeApp with id %s", appId)
	//}

	return nil
}

func (r *AppServeAppRepository) UpdateEndpoint(appId string, taskId string, endpoint string, previewEndpoint string, helmRevision int32) error {
	now := time.Now()
	app := domain.AppServeApp{
		ID:                 appId,
		EndpointUrl:        endpoint,
		PreviewEndpointUrl: previewEndpoint,
		UpdatedAt:          &now,
	}

	task := domain.AppServeAppTask{
		ID:           taskId,
		HelmRevision: helmRevision,
		UpdatedAt:    &now,
	}

	var res *gorm.DB
	if endpoint != "" && previewEndpoint != "" {
		// Both endpoints are valid
		res = r.db.Model(&app).Select("EndpointUrl", "PreviewEndpointUrl", "UpdatedAt").Updates(app)
	} else if endpoint != "" {
		// endpoint-only case
		res = r.db.Model(&app).Select("EndpointUrl", "UpdatedAt").Updates(app)
	} else if previewEndpoint != "" {
		// previewEndpoint-only case
		res = r.db.Model(&app).Select("PreviewEndpointUrl", "UpdatedAt").Updates(app)
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
		res = r.db.Model(&task).Select("HelmRevision", "UpdatedAt").Updates(task)
		if res.Error != nil || res.RowsAffected == 0 {
			return fmt.Errorf("UpdateEndpoint: "+
				"helm revision was not updated for AppServeAppTask with task ID %s", taskId)
		}
	}

	return nil
}

func (r *AppServeAppRepository) GetTaskCountById(appId string) (int64, error) {
	var count int64
	if err := r.db.Model(&domain.AppServeAppTask{}).Where("AppServeAppId = ?", appId).Count(&count); err != nil {
		return 0, fmt.Errorf("could not select count AppServeAppTask with ID: %s", appId)
	}
	return count, nil
}
