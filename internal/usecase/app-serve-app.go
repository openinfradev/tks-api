package usecase

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/domain"
	"github.com/openinfradev/tks-api/internal/repository"
	argowf "github.com/openinfradev/tks-api/pkg/argo-client"
	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/spf13/viper"
)

type IAppServeAppUsecase interface {
	Fetch(projectId string, showAll bool) ([]*domain.AppServeApp, error)
	Get(id string) (*domain.AppServeAppCombined, error)
	Create(app *domain.CreateAppServeAppRequest) (ret string, err error)
	Delete(id string) (res string, err error)
}

type AppServeAppUsecase struct {
	repo repository.IAppServeAppRepository
	argo argowf.ArgoClient
}

func NewAppServeAppUsecase(r repository.IAppServeAppRepository, argoClient argowf.ArgoClient) IAppServeAppUsecase {
	return &AppServeAppUsecase{
		repo: r,
		argo: argoClient,
	}
}

func (u *AppServeAppUsecase) Fetch(projectId string, showAll bool) (out []*domain.AppServeApp, err error) {
	out, err = u.repo.Fetch(projectId, showAll)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (u *AppServeAppUsecase) Get(id string) (out *domain.AppServeAppCombined, err error) {
	parsedId, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	out, err = u.repo.Get(parsedId)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (u *AppServeAppUsecase) Create(app *domain.CreateAppServeAppRequest) (ret string, err error) {
	if app == nil {
		return "", fmt.Errorf("Invalid app obj")
	}

	// For type 'build' and 'all', imageUrl and executablePath
	// are constructed based on pre-defined rule
	// (Refer to 'tks-appserve-template')
	if app.Type != "deploy" {
		// Validate param
		if app.ArtifactUrl == "" {
			return "", fmt.Errorf("Error: For 'build'/'all' type task, 'artifact_url' is mandatory param.")
		}

		// Construct imageUrl
		imageUrl := viper.GetString("image-registry-url") + "/" + app.Name + ":" + app.Version
		app.ImageUrl = imageUrl

		if app.AppType == "springboot" {
			// Construct executable_path
			artiUrl := app.ArtifactUrl
			tempArr := strings.Split(artiUrl, "/")
			exeFilename := tempArr[len(tempArr)-1]

			executablePath := "/usr/src/myapp/" + exeFilename
			app.ExecutablePath = executablePath
		}
	} else {
		// Validate param for 'deploy' type.
		// TODO: check params for legacy spring app case
		if app.AppType == "springboot" {
			if app.ImageUrl == "" || app.ExecutablePath == "" ||
				app.Profile == "" || app.ResourceSpec == "" {
				return "", fmt.Errorf(`Error: For 'deploy' type task, the following params must be provided.
- image_url
- executable_path
- profile
- resource_spec`)
			}
		}
	}

	// TODO: Validate PV params
	//
	//

	asa := &domain.AppServeApp{
		Name:               app.Name,
		ContractId:         app.ContractId,
		Type:               app.Type,
		AppType:            app.AppType,
		TargetClusterId:    app.TargetClusterId,
		EndpointUrl:        "",
		PreviewEndpointUrl: "N/A",
		Status:             "PREPARING",
	}

	asaTask := &domain.AppServeAppTask{
		Version:        app.Version,
		Strategy:       app.Strategy,
		ArtifactUrl:    app.ArtifactUrl,
		ImageUrl:       app.ImageUrl,
		ExecutablePath: app.ExecutablePath,
		ResourceSpec:   app.ResourceSpec,
		Status:         "PREPARING",
		Profile:        app.Profile,
		AppConfig:      app.AppConfig,
		AppSecret:      app.AppSecret,
		ExtraEnv:       app.ExtraEnv,
		Port:           app.Port,
		Output:         "",
	}

	asaId, asaTaskId, err := u.repo.Create(asa.ContractId, asa, asaTask)
	if err != nil {
		return "", fmt.Errorf("Failed to create app-serve application record. Err: %s", err)
	}

	// Save returned asa ID
	app.ID = asaId.String()
	taskId := asaTaskId.String()

	workflowId, err := u.argo.SumbitWorkflowFromWftpl("serve-java-app", argowf.SubmitOptions{
		Parameters: []string{
			"type=" + app.Type,
			"strategy=" + app.Strategy,
			"app_type=" + app.AppType,
			"target_cluster_id=" + app.TargetClusterId,
			"app_name=" + app.Name,
			"asa_id=" + app.ID,
			"asa_task_id=" + taskId,
			"artifact_url=" + app.ArtifactUrl,
			"image_url=" + app.ImageUrl,
			"port=" + app.Port,
			"profile=" + app.Profile,
			"extra_env=" + app.ExtraEnv,
			"app_config=" + app.AppConfig,
			"app_secret=" + app.AppSecret,
			"resource_spec=" + app.ResourceSpec,
			"executable_path=" + app.ExecutablePath,
			"git_repo_url=" + viper.GetString("git-repository-url"),
			"harbor_pw_secret=" + viper.GetString("harbor-pw-secret"),
			"pv_enabled=" + strconv.FormatBool(app.PvEnabled),
			"pv_storage_class=" + app.PvStorageClass,
			"pv_access_mode=" + app.PvAccessMode,
			"pv_size=" + app.PvSize,
			"pv_mount_path=" + app.PvMountPath,
		},
	})
	if err != nil {
		return "", fmt.Errorf("Failed to submit workflow. Err: %s", err)
	}
	log.Info("Successfully submited workflow: ", workflowId)

	return fmt.Sprintf(`The app <%[1]s> is being deployed. 
* App ID: %[2]s\n`, app.Name, app.ID), nil
}

func (u *AppServeAppUsecase) Delete(asaId string) (res string, err error) {
	parsedId, err := uuid.Parse(asaId)
	if err != nil {
		return "", err
	}

	asaCombined, err := u.repo.Get(parsedId)
	if err != nil {
		return "", fmt.Errorf("Error while getting ASA Info from DB. Err: %s", err)
	}

	// Validate app status
	if asaCombined.AppServeApp.Status == "WAIT_FOR_PROMOTE" ||
		asaCombined.AppServeApp.Status == "BLUEGREEN_FAILED" {
		return "", fmt.Errorf("The app is in blue-green related state. Promote or abort first before deleting!")
	}

	/********************
	 * Start delete task *
	 ********************/

	asaTask := &domain.AppServeAppTask{
		AppServeAppId: asaId,
		Version:       "",
		ArtifactUrl:   "",
		ImageUrl:      "",
		Status:        "DELETING",
		Profile:       "",
		Output:        "",
	}

	taskId, err := u.repo.Update(parsedId, asaTask)
	if err != nil {
		return "", fmt.Errorf("Failed to create delete task. Err: %s", err)
	}

	workflowId, err := u.argo.SumbitWorkflowFromWftpl("delete-java-app", argowf.SubmitOptions{
		Parameters: []string{
			"type=" + asaCombined.AppServeApp.Type,
			"target_cluster_id=" + asaCombined.AppServeApp.TargetClusterId,
			"app_name=" + asaCombined.AppServeApp.Name,
			"asa_id=" + asaId,
			"asa_task_id=" + taskId.String(),
			"artifact_url=" + "NA",
			"image_url=" + "NA",
			"port=" + "NA",
			"profile=" + "NA",
			"resource_spec=" + "NA",
			"executable_path=" + "NA",
		},
	})
	if err != nil {
		return "", fmt.Errorf("Failed to submit workflow. Err: %s", err)
	}
	log.Info("Successfully submited workflow: ", workflowId)

	res_str := fmt.Sprintf("The app '%s' is being deleted. Confirm result by checking the app status after a while.", &asaCombined.AppServeApp.Name)
	return res_str, nil
}
