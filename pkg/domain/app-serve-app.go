package domain

import "time"

type AppServeAppResponse struct {
	ID                 string     `json:"id,omitempty"`
	Name               string     `json:"name,omitempty"`               // application name
	Namespace          string     `json:"namespace,omitempty"`          // application namespace
	OrganizationId     string     `json:"organizationId,omitempty"`     // contractId is a contract ID which this app belongs to
	ProjectId          string     `json:"projectId,omitempty"`          // project ID which this app belongs to
	Type               string     `json:"type,omitempty"`               // type (build/deploy/all)
	AppType            string     `json:"appType,omitempty"`            // appType (spring/springboot)
	EndpointUrl        string     `json:"endpointUrl,omitempty"`        // endpoint URL of deployed app
	PreviewEndpointUrl string     `json:"previewEndpointUrl,omitempty"` // preview svc endpoint URL in B/G deployment
	TargetClusterId    string     `json:"targetClusterId,omitempty"`    // target cluster to which the app is deployed
	TargetClusterName  string     `json:"targetClusterName,omitempty"`  // target cluster name
	Status             string     `json:"status,omitempty"`             // status is status of deployed app
	GrafanaUrl         string     `json:"grafanaUrl,omitempty"`         // grafana dashboard URL for deployed app
	CreatedAt          time.Time  `json:"createdAt" `
	UpdatedAt          *time.Time `json:"updatedAt"`
	DeletedAt          *time.Time `json:"deletedAt"`
}

type AppServeAppTaskResponse struct {
	ID                string     `json:"id,omitempty"`
	AppServeAppId     string     `json:"appServeAppId,omitempty"`   // ID for appServeApp that this task belongs to
	Version           string     `json:"version,omitempty"`         // application version
	Status            string     `json:"status,omitempty"`          // status is app status
	Output            string     `json:"output,omitempty"`          // output for task result
	ArtifactUrl       string     `json:"artifactUrl,omitempty"`     // URL of java app artifact (Eg, Jar)
	ImageUrl          string     `json:"imageUrl,omitempty"`        // URL of built image for app
	ExecutablePath    string     `json:"executablePath,omitempty"`  // Executable path of app image
	Profile           string     `json:"profile,omitempty"`         // java app profile
	AppConfig         string     `json:"appConfig,omitempty"`       // java app config
	AppSecret         string     `json:"appSecret,omitempty"`       // java app secret
	ExtraEnv          string     `json:"extraEnv,omitempty"`        // env variable list for java app
	Port              string     `json:"port,omitempty"`            // java app port
	ResourceSpec      string     `json:"resourceSpec,omitempty"`    // resource spec of app pod
	HelmRevision      int32      `json:"helmRevision,omitempty"`    // revision of deployed helm release
	Strategy          string     `json:"strategy,omitempty"`        // deployment strategy (eg, rolling-update)
	RollbackVersion   string     `json:"rollbackVersion,omitempty"` // rollback target version
	PvEnabled         bool       `json:"pvEnabled"`
	PvStorageClass    string     `json:"pvStorageClass"`
	PvAccessMode      string     `json:"pvAccessMode"`
	PvSize            string     `json:"pvSize"`
	PvMountPath       string     `json:"pvMountPath"`
	AvailableRollback bool       `json:"availableRollback"`
	CreatedAt         time.Time  `json:"createdAt"` // createdAt is  a creation timestamp for the application
	UpdatedAt         *time.Time `json:"updatedAt"`
	DeletedAt         *time.Time `json:"deletedAt"`
}

type CreateAppServeAppRequest struct {
	// App
	Name            string `json:"name" validate:"required,rfc1123,name"`
	Namespace       string `json:"namespace"`
	Type            string `json:"type" `   // build deploy all
	AppType         string `json:"appType"` // springboot spring
	TargetClusterId string `json:"targetClusterId" validate:"required"`

	// Task
	Version        string `json:"version"`
	Strategy       string `json:"strategy"` // rolling-update blue-green canary
	ArtifactUrl    string `json:"artifactUrl"`
	ImageUrl       string `json:"imageUrl"`
	ExecutablePath string `json:"executablePath"`
	ResourceSpec   string `json:"resourceSpec"` // tiny medium large
	Profile        string `json:"profile"`
	AppConfig      string `json:"appConfig"`
	AppSecret      string `json:"appSecret"`
	ExtraEnv       string `json:"extraEnv"`
	Port           string `json:"port"`
	PvEnabled      bool   `json:"pvEnabled"`
	PvStorageClass string `json:"pvStorageClass"`
	PvAccessMode   string `json:"pvAccessMode"`
	PvSize         string `json:"pvSize"`
	PvMountPath    string `json:"pvMountPath"`
}

func (c *CreateAppServeAppRequest) SetDefaultValue() {
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
	ID   string `json:"appId"`
	Name string `json:"appName"`
}

type UpdateAppServeAppStatusRequest struct {
	TaskID string `json:"taskId" validate:"required"`
	Status string `json:"status" validate:"required"`
	Output string `json:"output"`
}

type UpdateAppServeAppEndpointRequest struct {
	TaskID             string `json:"taskId" validate:"required"`
	EndpointUrl        string `json:"endpointUrl"`
	PreviewEndpointUrl string `json:"previewEndpointUrl"`
	HelmRevision       int32  `json:"helmRevision"`
}

type UpdateAppServeAppRequest struct {
	// Task
	Strategy       string `json:"strategy"`
	ArtifactUrl    string `json:"artifactUrl"`
	ImageUrl       string `json:"imageUrl"`
	ExecutablePath string `json:"executablePath"`
	ResourceSpec   string `json:"resourceSpec"`
	Profile        string `json:"profile"`
	AppConfig      string `json:"appConfig"`
	AppSecret      string `json:"appSecret"`
	ExtraEnv       string `json:"extraEnv"`
	Port           string `json:"port"`

	// Update Strategy
	Promote bool `json:"promote"`
	Abort   bool `json:"abort"`
}

type RollbackAppServeAppRequest struct {
	TaskId string `json:"taskId"`
}

type GetAppServeAppsResponse struct {
	AppServeApps []AppServeAppResponse `json:"appServeApps"`
	Pagination   PaginationResponse    `json:"pagination"`
}

type GetAppServeAppTasksResponse struct {
	AppServeAppTasks []AppServeAppTaskResponse `json:"appServeAppTasks"`
	Pagination       PaginationResponse        `json:"pagination"`
}

type GetAppServeAppTaskResponse struct {
	AppServeApp     AppServeAppResponse     `json:"appServeApp"`
	AppServeAppTask AppServeAppTaskResponse `json:"appServeAppTask"`
	Stages          []StageResponse         `json:"stages"`
}

type StageResponse struct {
	Name    string            `json:"name"` // BUILD (빌드), DEPLOY (배포), PROMOTE (프로모트), ROLLBACK (롤백)
	Status  string            `json:"status"`
	Result  string            `json:"result"`
	Actions *[]ActionResponse `json:"actions"`
}

type ActionResponse struct {
	Name   string            `json:"name"` // ENDPOINT (화면보기), PREVIEW (미리보기), PROMOTE (배포), ABORT (중단)
	Uri    string            `json:"uri"`
	Type   string            `json:"type"` // LINK, API
	Method string            `json:"method"`
	Body   map[string]string `json:"body"`
}
