package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type AppServeApp struct {
	ID                 string            `gorm:"primarykey" json:"id,omitempty"`
	Name               string            `gorm:"index" json:"name,omitempty"`              // application name
	Namespace          string            `json:"namespace,omitempty"`                      // application namespace
	OrganizationId     string            `json:"organizationId,omitempty"`                 // contractId is a contract ID which this app belongs to
	ProjectId          string            `json:"projectId,omitempty"`                      // project ID which this app belongs to
	Type               string            `json:"type,omitempty"`                           // type (build/deploy/all)
	AppType            string            `json:"appType,omitempty"`                        // appType (spring/springboot)
	EndpointUrl        string            `json:"endpointUrl,omitempty"`                    // endpoint URL of deployed app
	PreviewEndpointUrl string            `json:"previewEndpointUrl,omitempty"`             // preview svc endpoint URL in B/G deployment
	TargetClusterId    string            `json:"targetClusterId,omitempty"`                // target cluster to which the app is deployed
	TargetClusterName  string            `gorm:"-:all" json:"targetClusterName,omitempty"` // target cluster name
	Status             string            `gorm:"index" json:"status,omitempty"`            // status is status of deployed app
	GrafanaUrl         string            `json:"grafanaUrl,omitempty"`                     // grafana dashboard URL for deployed app
	CreatedAt          time.Time         `gorm:"autoCreateTime:false" json:"createdAt" `
	UpdatedAt          *time.Time        `gorm:"autoUpdateTime:false" json:"updatedAt"`
	DeletedAt          *time.Time        `json:"deletedAt"`
	AppServeAppTasks   []AppServeAppTask `gorm:"foreignKey:AppServeAppId" json:"appServeAppTasks"`
}

type AppServeAppTask struct {
	ID                string     `gorm:"primarykey" json:"id,omitempty"`
	AppServeAppId     string     `gorm:"not null" json:"appServeAppId,omitempty"` // ID for appServeApp that this task belongs to
	Version           string     `json:"version,omitempty"`                       // application version
	Status            string     `json:"status,omitempty"`                        // status is app status
	Output            string     `json:"output,omitempty"`                        // output for task result
	ArtifactUrl       string     `json:"artifactUrl,omitempty"`                   // URL of java app artifact (Eg, Jar)
	ImageUrl          string     `json:"imageUrl,omitempty"`                      // URL of built image for app
	ExecutablePath    string     `json:"executablePath,omitempty"`                // Executable path of app image
	Profile           string     `json:"profile,omitempty"`                       // java app profile
	AppConfig         string     `json:"appConfig,omitempty"`                     // java app config
	AppSecret         string     `json:"appSecret,omitempty"`                     // java app secret
	ExtraEnv          string     `json:"extraEnv,omitempty"`                      // env variable list for java app
	Port              string     `json:"port,omitempty"`                          // java app port
	ResourceSpec      string     `json:"resourceSpec,omitempty"`                  // resource spec of app pod
	HelmRevision      int32      `gorm:"default:0" json:"helmRevision,omitempty"` // revision of deployed helm release
	Strategy          string     `json:"strategy,omitempty"`                      // deployment strategy (eg, rolling-update)
	RollbackVersion   string     `json:"rollbackVersion,omitempty"`               // rollback target version
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
