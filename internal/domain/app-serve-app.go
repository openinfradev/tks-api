package domain

import "time"

type AppServeApp = struct {
	Id string `json:"id,omitempty"`
	// application name
	Name string `json:"name,omitempty"`
	// contract_id is a contract ID which this app belongs to
	ContractId string `json:"contract_id,omitempty"`
	// type (build/deploy/all)
	Type string `json:"type,omitempty"`
	// app_type (spring/springboot)
	AppType string `json:"app_type,omitempty"`
	// endpoint URL of deployed app
	EndpointUrl string `json:"endpoint_url,omitempty"`
	// preview svc endpoint URL in B/G deployment
	PreviewEndpointUrl string `json:"preview_endpoint_url,omitempty"`
	// target cluster to which the app is deployed
	TargetClusterId string `json:"target_cluster_id,omitempty"`
	// status is status of deployed app
	Status string `json:"status,omitempty"`
	// created_at is a creatioin timestamp for the application
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type AppServeAppTask = struct {
	// task ID
	Id string `json:"id,omitempty"`
	// ID for appServeApp that this task belongs to.
	AppServeAppId string `json:"app_serve_app_id,omitempty"`
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
	Strategy string `json:"strategy,omitempty"`
	// created_at is  a creatioin timestamp for the application
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type AppServeAppCombined = struct {
	// app_serve_app represent basic ASA info.
	AppServeApp *AppServeApp `json:"app_serve_app,omitempty"`
	// tasks is a list of tasks that belongs to this ASA.
	Tasks []*AppServeAppTask `json:"tasks,omitempty"`
}

type CreateAppServeAppRequest = struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Type            string `json:"type"`
	Strategy        string `json:"strategy"`
	AppType         string `json:"app_type"`
	Version         string `json:"version"`
	ArtifactUrl     string `json:"artifact_url"`
	ImageUrl        string `json:"image_url"`
	Port            string `json:"port"`
	ContractId      string `json:"contract_id"`
	Profile         string `json:"profile"`
	ExtraEnv        string `json:"extra_env"`
	AppConfig       string `json:"app_config"`
	AppSecret       string `json:"app_secret"`
	TargetClusterId string `json:"target_cluster_id"`
	ExecutablePath  string `json:"executable_path"`
	ResourceSpec    string `json:"resource_spec"`
	PvEnabled       bool   `json:"pv_enabled"`
	PvStorageClass  string `json:"pv_storage_class"`
	PvAccessMode    string `json:"pv_access_mode"`
	PvSize          string `json:"pv_size"`
	PvMountPath     string `json:"pv_mount_path"`
}

type UpdateAppServeAppRequest = struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Type            string `json:"type"`
	Strategy        string `json:"strategy"`
	AppType         string `json:"app_type"`
	Version         string `json:"version"`
	ArtifactUrl     string `json:"artifact_url"`
	ImageUrl        string `json:"image_url"`
	Port            string `json:"port"`
	ContractId      string `json:"contract_id"`
	Profile         string `json:"profile"`
	ExtraEnv        string `json:"extra_env"`
	AppConfig       string `json:"app_config"`
	AppSecret       string `json:"app_secret"`
	TargetClusterId string `json:"target_cluster_id"`
	ExecutablePath  string `json:"executable_path"`
	ResourceSpec    string `json:"resource_spec"`
	PvEnabled       bool   `json:"pv_enabled"`
	PvStorageClass  string `json:"pv_storage_class"`
	PvAccessMode    string `json:"pv_access_mode"`
	PvSize          string `json:"pv_size"`
	PvMountPath     string `json:"pv_mount_path"`
	Promote         bool   `json:"promote"`
	Abort           bool   `json:"abort"`
}
