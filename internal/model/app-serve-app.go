package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type AppServeApp struct {
	ID                 string     `gorm:"primarykey" json:"id"`
	Name               string     `gorm:"index" json:"name"`              // application name
	Namespace          string     `json:"namespace"`                      // application namespace
	OrganizationId     string     `json:"organizationId"`                 // contractId is a contract ID which this app belongs to
	ProjectId          string     `json:"projectId"`                      // project ID which this app belongs to
	Type               string     `json:"type"`                           // type (build/deploy/all)
	AppType            string     `json:"appType"`                        // appType (spring/springboot)
	EndpointUrl        string     `json:"endpointUrl"`                    // endpoint URL of deployed app
	PreviewEndpointUrl string     `json:"previewEndpointUrl"`             // preview svc endpoint URL in B/G deployment
	TargetClusterId    string     `json:"targetClusterId"`                // target cluster to which the app is deployed
	TargetClusterName  string     `gorm:"-:all" json:"targetClusterName"` // target cluster name
	Status             string     `gorm:"index" json:"status"`            // status is status of deployed app
	GrafanaUrl         string     `json:"grafanaUrl"`                     // grafana dashboard URL for deployed app
	Description        string     `json:"description"`                    // description for application
	CreatedAt          time.Time  `gorm:"autoCreateTime:false" json:"createdAt" `
	UpdatedAt          *time.Time `gorm:"autoUpdateTime:false" json:"updatedAt"`
	DeletedAt          *time.Time `json:"deletedAt"`
}

type AppServeAppTask struct {
	ID                string     `gorm:"primarykey" json:"id"`
	AppServeAppId     string     `gorm:"not null" json:"appServeAppId"` // ID for appServeApp that this task belongs to
	Version           string     `json:"version"`                       // application version
	Status            string     `json:"status"`                        // status is app status
	Output            string     `json:"output"`                        // output for task result
	ArtifactUrl       string     `json:"artifactUrl"`                   // URL of java app artifact (Eg, Jar)
	ImageUrl          string     `json:"imageUrl"`                      // URL of built image for app
	ExecutablePath    string     `json:"executablePath"`                // Executable path of app image
	Profile           string     `json:"profile"`                       // java app profile
	AppConfig         string     `json:"appConfig"`                     // java app config
	AppSecret         string     `json:"appSecret"`                     // java app secret
	ExtraEnv          string     `json:"extraEnv"`                      // env variable list for java app
	Port              string     `json:"port"`                          // java app port
	ResourceSpec      string     `json:"resourceSpec"`                  // resource spec of app pod
	HelmRevision      int32      `gorm:"default:0" json:"helmRevision"` // revision of deployed helm release
	Strategy          string     `json:"strategy"`                      // deployment strategy (eg, rolling-update)
	RollbackVersion   string     `json:"rollbackVersion"`               // rollback target version
	PvEnabled         bool       `json:"pvEnabled"`
	PvStorageClass    string     `json:"pvStorageClass"`
	PvAccessMode      string     `json:"pvAccessMode"`
	PvSize            string     `json:"pvSize"`
	PvMountPath       string     `json:"pvMountPath"`
	AvailableRollback bool       `gorm:"-:all" json:"availableRollback"`
	CreatedAt         time.Time  `gorm:"autoCreateTime:false" json:"createdAt"` // createdAt is  a creation timestamp for the application
	UpdatedAt         *time.Time `gorm:"autoUpdateTime:false" json:"updatedAt"`
	DeletedAt         *time.Time `json:"deletedAt"`
}

func (t *AppServeAppTask) BeforeCreate(tx *gorm.DB) (err error) {
	t.ID = uuid.New().String()
	return nil
}
