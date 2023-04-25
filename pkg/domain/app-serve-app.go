package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AppServeApp struct {
	ID string `gorm:"primarykey" json:"id,omitempty"`
	// application name
	Name string `json:"name,omitempty"`
	// application namespace
	Namespace string `json:"namespace,omitempty"`
	// contract_id is a contract ID which this app belongs to
	OrganizationId string `json:"organizationId,omitempty"`
	// type (build/deploy/all)
	Type string `json:"type,omitempty"`
	// app_type (spring/springboot)
	AppType string `json:"appType,omitempty"`
	// endpoint URL of deployed app
	EndpointUrl string `json:"endpointUrl,omitempty"`
	// preview svc endpoint URL in B/G deployment
	PreviewEndpointUrl string `json:"previewEndpointUrl,omitempty"`
	// target cluster to which the app is deployed
	TargetClusterId string `json:"targetClusterId,omitempty"`
	// status is status of deployed app
	Status string `json:"status,omitempty"`
	// created_at is a creatioin timestamp for the application
	CreatedAt        time.Time         `gorm:"autoCreateTime:false" json:"createdAt" `
	UpdatedAt        *time.Time        `gorm:"autoUpdateTime:false" json:"updatedAt"`
	DeletedAt        *time.Time        `json:"deleted_at"`
	AppServeAppTasks []AppServeAppTask `gorm:"foreignKey:AppServeAppId" json:"app_serve_app_tasks"`
}

type AppServeAppTask struct {
	ID string `gorm:"primarykey" json:"id,omitempty"`
	// ID for appServeApp that this task belongs to.
	AppServeAppId string `gorm:"not null" json:"app_serve_app_id,omitempty"`
	// application version
	Version string `json:"version,omitempty"`
	// status is app status
	Status string `json:"status,omitempty"`
	// output for task result
	Output string `json:"output,omitempty"`
	// URL of java app artifact (Eg, Jar)
	ArtifactUrl string `json:"artifact_url,omitempty"`
	// URL of built image for app
	ImageUrl string `json:"image_url,omitempty"`
	// Executable path of app image
	ExecutablePath string `json:"executable_path,omitempty"`
	// java app profile
	Profile string `json:"profile,omitempty"`
	// java app config
	AppConfig string `json:"app_config,omitempty"`
	// java app secret
	AppSecret string `json:"app_secret,omitempty"`
	// env variable list for java app
	ExtraEnv string `json:"extra_env,omitempty"`
	// java app port
	Port string `json:"port,omitempty"`
	// resource spec of app pod
	ResourceSpec string `json:"resource_spec,omitempty"`
	// revision of deployed helm release
	HelmRevision int32 `json:"helm_revision,omitempty"`
	// deployment strategy (eg, rolling-update)
	Strategy       string `json:"strategy,omitempty"`
	PvEnabled      bool   `json:"pv_enabled"`
	PvStorageClass string `json:"pv_storage_class"`
	PvAccessMode   string `json:"pv_access_mode"`
	PvSize         string `json:"pv_size"`
	PvMountPath    string `json:"pv_mount_path"`
	// created_at is  a creation timestamp for the application
	CreatedAt time.Time  `gorm:"autoCreateTime:false" json:"created_at"`
	UpdatedAt *time.Time `gorm:"autoUpdateTime:false" json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

func (a *AppServeApp) BeforeCreate(tx *gorm.DB) (err error) {
	a.ID = uuid.New().String()
	return nil
}

func (t *AppServeAppTask) BeforeCreate(tx *gorm.DB) (err error) {
	t.ID = uuid.New().String()
	return nil
}

type CreateAppServeAppRequest struct {
	// App
	Name            string `json:"name" validate:"required"`
	Namespace       string `json:"namespace"`
	Type            string `json:"type" `    // build deploy all
	AppType         string `json:"app_type"` // springboot spring
	TargetClusterId string `json:"target_cluster_id" validate:"required"`

	// Task
	Version        string `json:"version"`
	Strategy       string `json:"strategy"` // rolling-update blue-green canary
	ArtifactUrl    string `json:"artifact_url"`
	ImageUrl       string `json:"image_url"`
	ExecutablePath string `json:"executable_path"`
	ResourceSpec   string `json:"resource_spec"` // tiny medium large
	Profile        string `json:"profile"`
	AppConfig      string `json:"app_config"`
	AppSecret      string `json:"app_secret"`
	ExtraEnv       string `json:"extra_env"`
	Port           string `json:"port"`
	PvEnabled      bool   `json:"pv_enabled"`
	PvStorageClass string `json:"pv_storage_class"`
	PvAccessMode   string `json:"pv_access_mode"`
	PvSize         string `json:"pv_size"`
	PvMountPath    string `json:"pv_mount_path"`
}

func (c *CreateAppServeAppRequest) SetDefaultValue() {
	if c.Namespace == "" {
		c.Namespace = c.Name
	}
	if c.Type == "" {
		c.Type = "all"
	}
	if c.AppType == "" {
		c.AppType = "springboot"
	}
	if c.Version == "" {
		c.Version = "1"
	}
	if c.Strategy == "" {
		c.Strategy = "rolling-update"
	}
	if c.ResourceSpec == "" {
		c.ResourceSpec = "medium"
	}
	if c.Profile == "" {
		c.Profile = "default"
	}
	if c.Port == "" {
		c.Port = "8080"
	}
}

type CreateAppServeAppResponse struct {
	ID   string `json:"app_id"`
	Name string `json:"app_name"`
}

type UpdateAppServeAppStatusRequest struct {
	TaskID string `json:"task_id" validate:"required"`
	Status string `json:"status" validate:"required"`
	Output string `json:"output"`
}

type UpdateAppServeAppEndpointRequest struct {
	TaskID             string `json:"task_id" validate:"required"`
	EndpointUrl        string `json:"endpoint_url"`
	PreviewEndpointUrl string `json:"preview_endpoint_url"`
	HelmRevision       int32  `json:"helm_revision"`
}

type UpdateAppServeAppRequest struct {
	// App
	Type    string `json:"type"`
	AppType string `json:"app_type"`

	// Task
	Strategy       string `json:"strategy"`
	ArtifactUrl    string `json:"artifact_url"`
	ImageUrl       string `json:"image_url"`
	ExecutablePath string `json:"executable_path"`
	ResourceSpec   string `json:"resource_spec"`
	Profile        string `json:"profile"`
	AppConfig      string `json:"app_config"`
	AppSecret      string `json:"app_secret"`
	ExtraEnv       string `json:"extra_env"`
	Port           string `json:"port"`

	// Update Strategy
	Promote bool `json:"promote"`
	Abort   bool `json:"abort"`
}

func (u *UpdateAppServeAppRequest) SetDefaultValue() {
	if u.Type == "" {
		u.Type = "all"
	}
	if u.AppType == "" {
		u.AppType = "springboot"
	}
	if u.Strategy == "" {
		u.Strategy = "rolling-update"
	}
	if u.ResourceSpec == "" {
		u.ResourceSpec = "medium"
	}
	if u.Profile == "" {
		u.Profile = "default"
	}
	if u.Port == "" {
		u.Port = "8080"
	}
}

type GetAppServeAppsResponse struct {
	AppServeApps []AppServeApp `json:"app_serve_apps"`
}

type GetAppServeAppResponse struct {
	AppServeApp AppServeApp `json:"app_serve_app"`
}
