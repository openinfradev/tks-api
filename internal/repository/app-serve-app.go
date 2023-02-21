package repository

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/openinfradev/tks-api/internal/domain"
)

// Interfaces
type IAppServeAppRepository interface {
	Fetch(contractId string, showAll bool) ([]*domain.AppServeApp, error)
	Get(id uuid.UUID) (*domain.AppServeAppCombined, error)
	Create(contractId string, app *domain.AppServeApp, task *domain.AppServeAppTask) (asaId uuid.UUID, asaTaskId uuid.UUID, err error)
	Update(appServeAppId uuid.UUID, task *domain.AppServeAppTask) (uuid.UUID, error)
}

type AppServeAppRepository struct {
	db *gorm.DB
}

func NewAppServeAppRepository(db *gorm.DB) IAppServeAppRepository {
	return &AppServeAppRepository{
		db: db,
	}
}

// Models
type AppServeApp struct {
	gorm.Model
	Id                 uuid.UUID `gorm:"primarykey;type:uuid"`
	Name               string
	ContractId         string
	Type               string
	AppType            string
	EndpointUrl        string
	PreviewEndpointUrl string
	TargetClusterId    string
	Status             string
}

func (c *AppServeApp) BeforeCreate(tx *gorm.DB) (err error) {
	c.Id = uuid.New()
	return nil
}

type AppServeAppTask struct {
	gorm.Model
	Id             uuid.UUID `gorm:"primarykey;type:uuid"`
	AppServeAppId  uuid.UUID
	Version        string
	Strategy       string
	Status         string
	Output         string
	ArtifactUrl    string
	ImageUrl       string
	ExecutablePath string
	ResourceSpec   string
	Profile        string
	AppConfig      string
	AppSecret      string
	ExtraEnv       string
	Port           string
	HelmRevision   int32
}

func (c *AppServeAppTask) BeforeCreate(tx *gorm.DB) (err error) {
	c.Id = uuid.New()
	return nil
}

// Logics
func (r *AppServeAppRepository) Create(contractId string, app *domain.AppServeApp, task *domain.AppServeAppTask) (asaId uuid.UUID, asaTaskId uuid.UUID, err error) {
	// TODO: should I set initial status field here?
	asaModel := AppServeApp{
		Name:               app.Name,
		ContractId:         contractId,
		Type:               app.Type,
		AppType:            app.AppType,
		EndpointUrl:        "N/A",
		PreviewEndpointUrl: "N/A",
		TargetClusterId:    app.TargetClusterId,
	}

	res := r.db.Create(&asaModel)
	if res.Error != nil {
		return uuid.Nil, uuid.Nil, res.Error
	}

	asaTaskModel := AppServeAppTask{
		Version:        task.Version,
		Strategy:       task.Strategy,
		Status:         task.Status,
		ArtifactUrl:    task.ArtifactUrl,
		ImageUrl:       task.ImageUrl,
		ExecutablePath: task.ExecutablePath,
		ResourceSpec:   task.ResourceSpec,
		Profile:        task.Profile,
		AppConfig:      task.AppConfig,
		AppSecret:      task.AppSecret,
		ExtraEnv:       task.ExtraEnv,
		Port:           task.Port,
		AppServeAppId:  asaModel.Id,
	}

	res = r.db.Create(&asaTaskModel)
	if res.Error != nil {
		return uuid.Nil, uuid.Nil, res.Error
	}

	return asaModel.Id, asaTaskModel.Id, nil
}

// Update creates new appServeApp Task for existing appServeApp.
func (r *AppServeAppRepository) Update(appServeAppId uuid.UUID, task *domain.AppServeAppTask) (uuid.UUID, error) {
	asaTaskModel := AppServeAppTask{
		Version:        task.Version,
		Strategy:       task.Strategy,
		Status:         task.Status,
		ArtifactUrl:    task.ArtifactUrl,
		ImageUrl:       task.ImageUrl,
		ExecutablePath: task.ExecutablePath,
		ResourceSpec:   task.ResourceSpec,
		Profile:        task.Profile,
		AppConfig:      task.AppConfig,
		AppSecret:      task.AppSecret,
		ExtraEnv:       task.ExtraEnv,
		Port:           task.Port,
		AppServeAppId:  appServeAppId,
	}

	res := r.db.Create(&asaTaskModel)
	if res.Error != nil {
		return uuid.Nil, res.Error
	}

	return asaTaskModel.Id, nil
}

func (r *AppServeAppRepository) Fetch(contractId string, showAll bool) ([]*domain.AppServeApp, error) {
	var appServeApps []AppServeApp
	pbAppServeApps := []*domain.AppServeApp{}

	queryStr := fmt.Sprintf("contract_id = '%s' AND status <> 'DELETE_SUCCESS'", contractId)
	if showAll {
		queryStr = fmt.Sprintf("contract_id = '%s'", contractId)
	}
	res := r.db.Order("created_at desc").Find(&appServeApps, queryStr)
	if res.Error != nil {
		return nil, fmt.Errorf("Error while finding appServeApps with contractID: %s", contractId)
	}

	// If no record is found, just return empty array.
	if res.RowsAffected == 0 {
		return pbAppServeApps, nil
	}

	for _, asa := range appServeApps {
		pbAppServeApps = append(pbAppServeApps, r.ConvertToPbAppServeApp(asa))
	}
	return pbAppServeApps, nil
}

func (r *AppServeAppRepository) Get(id uuid.UUID) (*domain.AppServeAppCombined, error) {
	var appServeApp AppServeApp
	var appServeAppTasks []AppServeAppTask
	pbAppServeAppCombined := &domain.AppServeAppCombined{}

	res := r.db.First(&appServeApp, "id = ?", id)
	if res.RowsAffected == 0 || res.Error != nil {
		return nil, fmt.Errorf("Could not find AppServeApp with ID: %s", id)
	}
	pbAppServeAppCombined.AppServeApp = r.ConvertToPbAppServeApp(appServeApp)

	res = r.db.Order("created_at desc").Find(&appServeAppTasks, "app_serve_app_id = ?", id)
	if res.Error != nil {
		return nil, fmt.Errorf("Error while finding appServeAppTasks with appServeApp ID %s. Err: %s", id, res.Error)
	}

	for _, task := range appServeAppTasks {
		pbAppServeAppCombined.Tasks = append(pbAppServeAppCombined.Tasks, r.ConvertToPbAppServeAppTask(task))
	}

	return pbAppServeAppCombined, nil
}

func (r *AppServeAppRepository) UpdateStatus(taskId uuid.UUID, status string, output string) error {
	// Update task status
	res := r.db.Model(&AppServeAppTask{}).Where("ID = ?", taskId).Updates(AppServeAppTask{Status: status, Output: output})

	if res.Error != nil || res.RowsAffected == 0 {
		return fmt.Errorf("UpdateStatus: nothing updated in AppServeAppTask with ID %s", taskId)
	}

	// Get Asa ID which this task belongs to.
	var appServeAppTask AppServeAppTask
	res = r.db.First(&appServeAppTask, "id = ?", taskId)
	if res.RowsAffected == 0 || res.Error != nil {
		return fmt.Errorf("Could not find AppServeAppTask with ID: %s", taskId)
	}
	asaId := appServeAppTask.AppServeAppId

	// Update status of the Asa.
	res = r.db.Model(&AppServeApp{}).Where("ID = ?", asaId).Update("Status", status)
	if res.Error != nil || res.RowsAffected == 0 {
		return fmt.Errorf("UpdateStatus: nothing updated in AppServeApp with id %s", asaId)
	}

	return nil
}

func (r *AppServeAppRepository) UpdateEndpoint(id uuid.UUID, taskId uuid.UUID, endpoint string, previewEndpoint string, helmRevision int32) error {
	if endpoint != "" && previewEndpoint != "" {
		// Both endpoints are valid
		res := r.db.Model(&AppServeApp{}).Where("ID = ?", id).Updates(AppServeApp{EndpointUrl: endpoint, PreviewEndpointUrl: previewEndpoint})
		if res.Error != nil || res.RowsAffected == 0 {
			return fmt.Errorf("UpdateEndpoint: nothing updated in AppServeApp with id %s", id)
		}
	} else if endpoint != "" {
		// endpoint-only case
		res := r.db.Model(&AppServeApp{}).Where("ID = ?", id).Update("EndpointUrl", endpoint)
		if res.Error != nil || res.RowsAffected == 0 {
			return fmt.Errorf("UpdateEndpoint: nothing updated in AppServeApp with id %s", id)
		}
	} else if previewEndpoint != "" {
		// previewEndpoint-only case
		res := r.db.Model(&AppServeApp{}).Where("ID = ?", id).Update("PreviewEndpointUrl", previewEndpoint)
		if res.Error != nil || res.RowsAffected == 0 {
			return fmt.Errorf("UpdateEndpoint: nothing updated in AppServeApp with id %s", id)
		}
	} else {
		return fmt.Errorf("UpdateEndpoint: No endpoint provided. At least one of [endpoint, preview_endpoint] should be provided.")
	}

	// Update helm revision
	// Ignore if the value is less than 0
	if helmRevision > 0 {
		res := r.db.Model(&AppServeAppTask{}).Where("ID = ?", taskId).Update("HelmRevision", helmRevision)
		if res.Error != nil || res.RowsAffected == 0 {
			return fmt.Errorf("UpdateEndpoint: helm revision was not updated for AppServeAppTask with task ID %s", taskId)
		}
	}

	return nil
}

func (r *AppServeAppRepository) ConvertToPbAppServeApp(asa AppServeApp) *domain.AppServeApp {
	return &domain.AppServeApp{
		Id:                 asa.Id.String(),
		Name:               asa.Name,
		ContractId:         asa.ContractId,
		Type:               asa.Type,
		AppType:            asa.AppType,
		Status:             asa.Status,
		EndpointUrl:        asa.EndpointUrl,
		PreviewEndpointUrl: asa.PreviewEndpointUrl,
		TargetClusterId:    asa.TargetClusterId,
		CreatedAt:          asa.CreatedAt,
		UpdatedAt:          asa.UpdatedAt,
	}
}

func (r *AppServeAppRepository) ConvertToPbAppServeAppTask(task AppServeAppTask) *domain.AppServeAppTask {
	return &domain.AppServeAppTask{
		Id:             task.Id.String(),
		Version:        task.Version,
		Strategy:       task.Strategy,
		Status:         task.Status,
		Output:         task.Output,
		ImageUrl:       task.ImageUrl,
		ArtifactUrl:    task.ArtifactUrl,
		ResourceSpec:   task.ResourceSpec,
		ExecutablePath: task.ExecutablePath,
		Profile:        task.Profile,
		AppConfig:      task.AppConfig,
		AppSecret:      task.AppSecret,
		ExtraEnv:       task.ExtraEnv,
		Port:           task.Port,
		HelmRevision:   task.HelmRevision,
		CreatedAt:      task.CreatedAt,
		UpdatedAt:      task.UpdatedAt,
	}
}
