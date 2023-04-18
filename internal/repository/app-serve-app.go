package repository

import (
	"fmt"
	"gorm.io/gorm"
	"time"

	"github.com/openinfradev/tks-api/pkg/domain"
)

type IAppServeAppRepository interface {
	CreateAppServeApp(app *domain.AppServeApp) (appId string, taskId string, err error)
	GetAppServeApps(organizationId string, showAll bool) ([]domain.AppServeApp, error)
	GetAppServeAppById(appId string) (*domain.AppServeApp, error)
	CreateTask(task *domain.AppServeAppTask) (taskId string, err error)
	UpdateStatus(appId string, taskId string, status string, output string) error
	UpdateEndpoint(appId string, taskId string, endpoint string, previewEndpoint string, helmRevision int32) error
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

// Update creates new appServeApp Task for existing appServeApp.
func (r *AppServeAppRepository) CreateTask(
	task *domain.AppServeAppTask) (string, error) {
	res := r.db.Create(task)
	if res.Error != nil {
		return "", res.Error
	}

	return task.ID, nil
}

func (r *AppServeAppRepository) GetAppServeApps(organizationId string, showAll bool) ([]domain.AppServeApp, error) {
	var apps []domain.AppServeApp

	queryStr := fmt.Sprintf("organization_id = '%s' AND status <> 'DELETE_SUCCESS'", organizationId)
	if showAll {
		queryStr = fmt.Sprintf("organization_id = '%s'", organizationId)
	}
	res := r.db.Order("created_at desc").Find(&apps, queryStr)
	if res.Error != nil {
		return nil, fmt.Errorf("error while finding appServeApps with organizationId: %s", organizationId)
	}

	// If no record is found, just return empty array.
	if res.RowsAffected == 0 {
		return apps, nil
	}

	return apps, nil
}

func (r *AppServeAppRepository) GetAppServeAppById(appId string) (*domain.AppServeApp, error) {
	var app domain.AppServeApp

	if err := r.db.Where("id = ?", appId).First(&app).Error; err != nil {
		return nil, fmt.Errorf("could not find AppServeApp with ID: %s", appId)
	}

	err := r.db.Model(&app).Order("created_at desc").Association("AppServeAppTasks").Find(&app.AppServeAppTasks)
	if err != nil {
		return nil, err
	}

	//res := r.db.Order("app_serve_app_tasks.created_at asc").
	//	Joins("Join app_serve_app_tasks On app_serve_app_tasks.app_serve_app_id = app_serve_apps.id").
	//	First(&repoApp, "app_serve_apps.id = ?", appId)
	//if res.RowsAffected == 0 || res.Error != nil {
	//	return nil, fmt.Errorf("could not find AppServeApp with ID: %s", appId)
	//}

	return &app, nil
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
