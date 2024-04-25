package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/repository"
	argowf "github.com/openinfradev/tks-api/pkg/argo-client"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/kubernetes"
	"github.com/openinfradev/tks-api/pkg/log"
)

type IAppServeAppUsecase interface {
	CreateAppServeApp(ctx context.Context, app *model.AppServeApp) (appId string, taskId string, err error)
	GetAppServeApps(ctx context.Context, organizationId string, projectId string, showAll bool, pg *pagination.Pagination) ([]model.AppServeApp, error)
	GetAppServeAppById(ctx context.Context, appId string) (*model.AppServeApp, error)
	GetAppServeAppTasks(ctx context.Context, appId string, pg *pagination.Pagination) ([]model.AppServeAppTask, error)
	GetAppServeAppTaskById(ctx context.Context, taskId string) (*model.AppServeAppTask, *model.AppServeApp, error)
	GetAppServeAppLatestTask(ctx context.Context, appId string) (*model.AppServeAppTask, *model.AppServeApp, error)
	GetNumOfAppsOnStack(ctx context.Context, organizationId string, clusterId string) (int64, error)
	IsAppServeAppExist(ctx context.Context, appId string) (bool, error)
	IsAppServeAppNameExist(ctx context.Context, orgId string, appName string) (bool, error)
	IsAppServeAppNamespaceExist(ctx context.Context, clusterId string, namespace string) (bool, error)
	UpdateAppServeAppStatus(ctx context.Context, appId string, taskId string, status string, output string) (ret string, err error)
	DeleteAppServeApp(ctx context.Context, appId string) (res string, err error)
	UpdateAppServeApp(ctx context.Context, app *model.AppServeApp, appTask *model.AppServeAppTask) (ret string, err error)
	UpdateAppServeAppEndpoint(ctx context.Context, appId string, taskId string, endpoint string, previewEndpoint string, helmRevision int32) (string, error)
	PromoteAppServeApp(ctx context.Context, appId string) (ret string, err error)
	AbortAppServeApp(ctx context.Context, appId string) (ret string, err error)
	RollbackAppServeApp(ctx context.Context, appId string, taskId string) (ret string, err error)
}

type AppServeAppUsecase struct {
	repo             repository.IAppServeAppRepository
	organizationRepo repository.IOrganizationRepository
	appGroupRepo     repository.IAppGroupRepository
	argo             argowf.ArgoClient
}

func NewAppServeAppUsecase(r repository.Repository, argoClient argowf.ArgoClient) IAppServeAppUsecase {
	return &AppServeAppUsecase{
		repo:             r.AppServeApp,
		organizationRepo: r.Organization,
		appGroupRepo:     r.AppGroup,
		argo:             argoClient,
	}
}

func (u *AppServeAppUsecase) CreateAppServeApp(ctx context.Context, app *model.AppServeApp) (string, string, error) {
	if app == nil {
		return "", "", fmt.Errorf("invalid app obj")
	}

	// For type 'build' and 'all', imageUrl and executablePath
	// are constructed based on pre-defined rule
	// (Refer to 'tks-appserve-template')
	if app.Type != "deploy" {
		// Validate param
		if app.AppServeAppTasks[0].ArtifactUrl == "" {
			return "", "", fmt.Errorf("error: For 'build'/'all' type task, 'artifact_url' is mandatory param")
		}

		// Construct imageUrl
		imageUrl := viper.GetString("image-registry-url") + "/" + app.Name + "-" + app.TargetClusterId + ":" + app.AppServeAppTasks[0].Version
		app.AppServeAppTasks[0].ImageUrl = imageUrl

		if app.AppType == "springboot" {
			// Construct executable_path
			artiUrl := app.AppServeAppTasks[0].ArtifactUrl
			tempArr := strings.Split(artiUrl, "/")
			exeFilename := tempArr[len(tempArr)-1]

			executablePath := "/usr/src/myapp/" + exeFilename
			app.AppServeAppTasks[0].ExecutablePath = executablePath
		}
	} else {
		// Validate param for 'deploy' type.
		// TODO: check params for legacy spring app case
		if app.AppType == "springboot" {
			if app.AppServeAppTasks[0].ImageUrl == "" || app.AppServeAppTasks[0].ExecutablePath == "" ||
				app.AppServeAppTasks[0].Profile == "" || app.AppServeAppTasks[0].ResourceSpec == "" {
				return "",
					"",
					fmt.Errorf("Error: For 'deploy' type task, the following params must be provided." +
						"\n\t- image_url\n\t- executable_path\n\t- profile\n\t- resource_spec")
			}
		}
	}

	extEnv := app.AppServeAppTasks[0].ExtraEnv
	if extEnv != "" {
		/* Preprocess extraEnv param */
		log.Debug(ctx, "extraEnv received: ", extEnv)

		tempMap := map[string]string{}
		err := json.Unmarshal([]byte(extEnv), &tempMap)
		if err != nil {
			log.Error(ctx, err)
			return "", "", errors.Wrap(err, "Failed to process extraEnv param.")
		}
		log.Debugf(ctx, "extraEnv marshalled: %v", tempMap)

		newExtEnv := map[string]string{}
		for key, val := range tempMap {
			newkey := "\"" + key + "\""
			newval := "\"" + val + "\""
			newExtEnv[newkey] = newval
		}

		mJson, _ := json.Marshal(newExtEnv)
		extEnv = string(mJson)
		log.Debug(ctx, "After transform, extraEnv: ", extEnv)
	}

	appId, taskId, err := u.repo.CreateAppServeApp(ctx, app)
	if err != nil {
		log.Error(ctx, err)
		return "", "", errors.Wrap(err, "Failed to create app.")
	}

	fmt.Printf("appId = %s, taskId = %s", appId, taskId)

	// TODO: Validate PV params

	// Call argo workflow
	workflow := "serve-java-app"

	opts := argowf.SubmitOptions{}
	opts.Parameters = []string{
		"type=" + app.Type,
		"strategy=" + app.AppServeAppTasks[0].Strategy,
		"app_type=" + app.AppType,
		"organization_id=" + app.OrganizationId,
		"project_id=" + app.ProjectId,
		"target_cluster_id=" + app.TargetClusterId,
		"app_name=" + app.Name,
		"namespace=" + app.Namespace,
		"asa_id=" + appId,
		"asa_task_id=" + taskId,
		"artifact_url=" + app.AppServeAppTasks[0].ArtifactUrl,
		"image_url=" + app.AppServeAppTasks[0].ImageUrl,
		"port=" + app.AppServeAppTasks[0].Port,
		"profile=" + app.AppServeAppTasks[0].Profile,
		"extra_env=" + extEnv,
		"app_config=" + app.AppServeAppTasks[0].AppConfig,
		"app_secret=" + app.AppServeAppTasks[0].AppSecret,
		"resource_spec=" + app.AppServeAppTasks[0].ResourceSpec,
		"executable_path=" + app.AppServeAppTasks[0].ExecutablePath,
		"git_repo_url=" + viper.GetString("git-repository-url"),
		"harbor_pw_secret=" + viper.GetString("harbor-pw-secret"),
		"pv_enabled=" + strconv.FormatBool(app.AppServeAppTasks[0].PvEnabled),
		"pv_storage_class=" + app.AppServeAppTasks[0].PvStorageClass,
		"pv_access_mode=" + app.AppServeAppTasks[0].PvAccessMode,
		"pv_size=" + app.AppServeAppTasks[0].PvSize,
		"pv_mount_path=" + app.AppServeAppTasks[0].PvMountPath,
		"tks_api_url=" + viper.GetString("external-address"),
	}

	log.Info(ctx, "Submitting workflow: ", workflow)

	workflowId, err := u.argo.SumbitWorkflowFromWftpl(ctx, workflow, opts)
	if err != nil {
		log.Error(ctx, err)
		return "", "", errors.Wrap(err, fmt.Sprintf("failed to submit workflow. %s", workflow))
	}
	log.Info(ctx, "Successfully submitted workflow: ", workflowId)

	return appId, app.Name, nil
}

func (u *AppServeAppUsecase) GetAppServeApps(ctx context.Context, organizationId string, projectId string, showAll bool, pg *pagination.Pagination) ([]model.AppServeApp, error) {
	apps, err := u.repo.GetAppServeApps(ctx, organizationId, projectId, showAll, pg)
	if err != nil {
		log.Debugf(ctx, "Apps: [%v]", apps)
	}

	return apps, nil
}

func (u *AppServeAppUsecase) GetAppServeAppById(ctx context.Context, appId string) (*model.AppServeApp, error) {
	asa, err := u.repo.GetAppServeAppById(ctx, appId)
	if err != nil {
		return nil, err
	}

	/************************
	* Construct grafana URL *
	************************/
	organization, err := u.organizationRepo.Get(ctx, asa.OrganizationId)
	if err != nil {
		return asa, httpErrors.NewInternalServerError(errors.Wrap(err, fmt.Sprintf("Failed to get organization for app %s", asa.Name)), "S_FAILED_FETCH_ORGANIZATION", "")
	}

	appGroupsInPrimaryCluster, err := u.appGroupRepo.Fetch(ctx, domain.ClusterId(organization.PrimaryClusterId), nil)
	if err != nil {
		return asa, err
	}

	for _, appGroup := range appGroupsInPrimaryCluster {
		if appGroup.AppGroupType == domain.AppGroupType_LMA {
			applications, err := u.appGroupRepo.GetApplications(ctx, appGroup.ID, domain.ApplicationType_GRAFANA)
			if err != nil {
				return asa, err
			}
			if len(applications) > 0 {
				asa.GrafanaUrl = applications[0].Endpoint + "/d/tks_appserving_dashboard/tks-appserving-dashboard?refresh=30s&var-cluster=" + asa.TargetClusterId + "&var-kubernetes_namespace_name=" + asa.Namespace + "&var-kubernetes_pod_name=All&var-kubernetes_container_name=main&var-TopK=10"
			}
		}
	}

	return asa, nil
}

func (u *AppServeAppUsecase) GetAppServeAppTasks(ctx context.Context, appId string, pg *pagination.Pagination) ([]model.AppServeAppTask, error) {
	tasks, err := u.repo.GetAppServeAppTasksByAppId(ctx, appId, pg)
	if err != nil {
		log.Debugf(ctx, "Tasks: %v", tasks)
	}

	return tasks, nil
}

func (u *AppServeAppUsecase) GetAppServeAppTaskById(ctx context.Context, taskId string) (*model.AppServeAppTask, *model.AppServeApp, error) {
	task, app, err := u.repo.GetAppServeAppTaskById(ctx, taskId)
	if err != nil {
		return nil, nil, err
	}

	return task, app, nil
}

func (u *AppServeAppUsecase) GetAppServeAppLatestTask(ctx context.Context, appId string) (*model.AppServeAppTask, *model.AppServeApp, error) {
	task, app, err := u.repo.GetAppServeAppLatestTask(ctx, appId)
	if err != nil {
		return nil, nil, err
	}

	return task, app, nil
}

func (u *AppServeAppUsecase) GetNumOfAppsOnStack(ctx context.Context, organizationId string, clusterId string) (int64, error) {
	numApps, err := u.repo.GetNumOfAppsOnStack(ctx, organizationId, clusterId)
	if err != nil {
		return -1, err
	}

	return numApps, nil
}

func (u *AppServeAppUsecase) IsAppServeAppExist(ctx context.Context, appId string) (bool, error) {
	count, err := u.repo.IsAppServeAppExist(ctx, appId)
	if err != nil {
		return false, err
	}

	if count > 0 {
		return true, nil
	}

	return false, nil
}

func (u *AppServeAppUsecase) IsAppServeAppNameExist(ctx context.Context, orgId string, appName string) (bool, error) {
	count, err := u.repo.IsAppServeAppNameExist(ctx, orgId, appName)
	if err != nil {
		return false, err
	}

	if count > 0 {
		return true, nil
	}

	return false, nil
}

func (u *AppServeAppUsecase) IsAppServeAppNamespaceExist(ctx context.Context, clusterId string, new_ns string) (bool, error) {
	clientset, err := kubernetes.GetClientFromClusterId(ctx, clusterId)
	if err != nil {
		log.Error(ctx, err)
		return false, err
	}

	namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Error(ctx, err)
		return false, err
	}
	for _, ns := range namespaces.Items {
		if new_ns == ns.ObjectMeta.Name {
			log.Debugf(ctx, "Namespace %s already exists.", new_ns)
			return true, nil
		}
	}
	log.Debugf(ctx, "Namespace %s is available", new_ns)
	return false, nil
}

func (u *AppServeAppUsecase) UpdateAppServeAppStatus(ctx context.Context, appId string, taskId string, status string,
	output string) (string, error) {

	log.Info(ctx, "Starting status update process..")

	err := u.repo.UpdateStatus(ctx, appId, taskId, status, output)
	if err != nil {
		log.Info(ctx, "appId = ", appId)
		log.Info(ctx, "taskId = ", taskId)
		return "", fmt.Errorf("failed to update app status. Err: %s", err)
	}
	return fmt.Sprintf("The appId '%s' status is being updated.", appId), nil
}

func (u *AppServeAppUsecase) UpdateAppServeAppEndpoint(
	ctx context.Context,
	appId string,
	taskId string,
	endpoint string,
	previewEndpoint string,
	helmRevision int32) (string, error) {

	log.Info(ctx, "Starting endpoint update process..")

	err := u.repo.UpdateEndpoint(ctx, appId, taskId, endpoint, previewEndpoint, helmRevision)
	if err != nil {
		log.Info(ctx, "appId = ", appId)
		log.Info(ctx, "taskId = ", taskId)
		return "", fmt.Errorf("failed to update endpoint. Err: %s", err)
	}
	return fmt.Sprintf("The appId '%s' endpoint is being updated.", appId), nil
}

func (u *AppServeAppUsecase) DeleteAppServeApp(ctx context.Context, appId string) (res string, err error) {
	app, err := u.repo.GetAppServeAppById(ctx, appId)
	if err != nil {
		return "", fmt.Errorf("error while getting ASA Info from DB. Err: %s", err)
	}

	if app == nil {
		return "", httpErrors.NewNoContentError(fmt.Errorf("the appId doesn't exist"), "", "")
	}
	// Validate app status
	// TODO: Add common helper function for this kind of status validation
	if app.Status == "BUILDING" || app.Status == "DEPLOYING" ||
		app.Status == "PROMOTING" || app.Status == "ABORTING" {
		return "작업 진행 중에는 앱을 삭제할 수 없습니다", fmt.Errorf("Can't delete app while the task is in progress.")
	}

	/********************
	 * Start delete task *
	 ********************/

	appTask := &model.AppServeAppTask{
		AppServeAppId: app.ID,
		Version:       strconv.Itoa(len(app.AppServeAppTasks) + 1),
		ArtifactUrl:   "",
		ImageUrl:      app.AppServeAppTasks[0].ImageUrl,
		Status:        "DELETING",
		Profile:       app.AppServeAppTasks[0].Profile,
		Output:        "",
		CreatedAt:     time.Now(),
	}

	taskId, err := u.repo.CreateTask(ctx, appTask)
	if err != nil {
		log.Error(ctx, "taskId = ", taskId)
		log.Error(ctx, "Failed to create delete task. Err:", err)
		return "", errors.Wrap(err, "Failed to create delete task.")
	}

	log.Info(ctx, "Updating app status to 'DELETING'..")

	err = u.repo.UpdateStatus(ctx, appId, taskId, "DELETING", "")
	if err != nil {
		log.Debug(ctx, "appId = ", appId)
		log.Debug(ctx, "taskId = ", taskId)
		return "", fmt.Errorf("failed to update app status on DeleteAppServeApp. Err: %s", err)
	}

	workflow := "delete-java-app"
	log.Info(ctx, "Submitting workflow: ", workflow)

	workflowId, err := u.argo.SumbitWorkflowFromWftpl(ctx, workflow, argowf.SubmitOptions{
		Parameters: []string{
			"target_cluster_id=" + app.TargetClusterId,
			"app_name=" + app.Name,
			"namespace=" + app.Namespace,
			"asa_id=" + app.ID,
			"asa_task_id=" + taskId,
			"organization_id=" + app.OrganizationId,
			"project_id=" + app.ProjectId,
			"tks_api_url=" + viper.GetString("external-address"),
		},
	})
	if err != nil {
		log.Error(ctx, "Failed to submit workflow. Err:", err)
		return "", errors.Wrap(err, "Failed to submit workflow.")
	}
	log.Info(ctx, "Successfully submitted workflow: ", workflowId)

	return fmt.Sprintf("The app %s is being deleted. "+
		"Confirm result by checking the app status after a while.", app.Name), nil
}

func (u *AppServeAppUsecase) UpdateAppServeApp(ctx context.Context, app *model.AppServeApp, appTask *model.AppServeAppTask) (ret string, err error) {
	if appTask == nil {
		return "", errors.New("invalid parameters. appTask is nil")
	}

	app_, err := u.repo.GetAppServeAppById(ctx, app.ID)
	if err != nil {
		return "", fmt.Errorf("error while getting ASA Info from DB. Err: %s", err)
	}

	// Block update if the app's current status is one of those.
	if app_.Status == "PROMOTE_WAIT" || app_.Status == "PROMOTING" || app_.Status == "ABORTING" {
		return "승인대기 또는 프로모트 작업 중에는 업그레이드를 수행할 수 없습니다", fmt.Errorf("Update not possible. The app is waiting for promote or in the middle of promote process.")
	}

	log.Info(ctx, "Starting normal update process..")

	// TODO: for more strict validation, check if immutable fields are provided by user
	// and those values are changed or not. (name, type, app_type, target_cluster)

	// Validate 'strategy' param
	if !(appTask.Strategy == "rolling-update" || appTask.Strategy == "blue-green" || appTask.Strategy == "canary") {
		return "", fmt.Errorf("Error: 'strategy' should be one of these values." +
			"\n\t- rolling-update\n\t- blue-green\n\t- canary")
	}

	if app.Type != "deploy" {
		// Construct imageUrl
		imageUrl := viper.GetString("image-registry-url") + "/" + app.Name + "-" + app.TargetClusterId + ":" + appTask.Version
		appTask.ImageUrl = imageUrl

		// Construct executable_path
		if app.AppType == "springboot" {
			artiUrl := appTask.ArtifactUrl
			tempArr := strings.Split(artiUrl, "/")
			exeFilename := tempArr[len(tempArr)-1]

			executablePath := "/usr/src/myapp/" + exeFilename
			appTask.ExecutablePath = executablePath
		}
	}

	extEnv := appTask.ExtraEnv
	if extEnv != "" {
		/* Preprocess extraEnv param */
		log.Debug(ctx, "extraEnv received: ", extEnv)

		tempMap := map[string]string{}
		err = json.Unmarshal([]byte(extEnv), &tempMap)
		if err != nil {
			log.Error(ctx, err)
			return "", errors.Wrap(err, "Failed to process extraEnv param.")
		}
		log.Debugf(ctx, "extraEnv marshalled: %v", tempMap)

		newExtEnv := map[string]string{}
		for key, val := range tempMap {
			newkey := "\"" + key + "\""
			newval := "\"" + val + "\""
			newExtEnv[newkey] = newval
		}

		mJson, _ := json.Marshal(newExtEnv)
		extEnv = string(mJson)
		log.Debug(ctx, "After transform, extraEnv: ", extEnv)
	}

	taskId, err := u.repo.CreateTask(ctx, appTask)
	if err != nil {
		log.Info(ctx, "taskId = ", taskId)
		return "", fmt.Errorf("failed to update app-serve application. Err: %s", err)
	}

	// Sync new task status to the parent app
	log.Info(ctx, "Updating app status to 'PREPARING'..")

	err = u.repo.UpdateStatus(ctx, app.ID, taskId, "PREPARING", "")
	if err != nil {
		log.Debug(ctx, "appId = ", app.ID)
		log.Debug(ctx, "taskId = ", taskId)
		return "", fmt.Errorf("failed to update app status on UpdateAppServeApp. Err: %s", err)
	}

	// Call argo workflow
	workflow := "serve-java-app"

	log.Info(ctx, "Submitting workflow: ", workflow)

	workflowId, err := u.argo.SumbitWorkflowFromWftpl(ctx, workflow, argowf.SubmitOptions{
		Parameters: []string{
			"type=" + app.Type,
			"strategy=" + appTask.Strategy,
			"app_type=" + app.AppType,
			"organization_id=" + app.OrganizationId,
			"project_id=" + app.ProjectId,
			"target_cluster_id=" + app.TargetClusterId,
			"app_name=" + app.Name,
			"namespace=" + app.Namespace,
			"asa_id=" + app.ID,
			"asa_task_id=" + taskId,
			"artifact_url=" + appTask.ArtifactUrl,
			"image_url=" + appTask.ImageUrl,
			"port=" + appTask.Port,
			"profile=" + appTask.Profile,
			"extra_env=" + extEnv,
			"app_config=" + appTask.AppConfig,
			"app_secret=" + appTask.AppSecret,
			"resource_spec=" + appTask.ResourceSpec,
			"executable_path=" + appTask.ExecutablePath,
			"git_repo_url=" + viper.GetString("git-repository-url"),
			"harbor_pw_secret=" + viper.GetString("harbor-pw-secret"),
			"pv_enabled=" + strconv.FormatBool(appTask.PvEnabled),
			"pv_storage_class=" + appTask.PvStorageClass,
			"pv_access_mode=" + appTask.PvAccessMode,
			"pv_size=" + appTask.PvSize,
			"pv_mount_path=" + appTask.PvMountPath,
			"tks_api_url=" + viper.GetString("external-address"),
		},
	})
	if err != nil {
		log.Error(ctx, "Failed to submit workflow. Err:", err)
		return "", fmt.Errorf("failed to submit workflow. Err: %s", err)
	}
	log.Info(ctx, "Successfully submitted workflow: ", workflowId)

	var message string
	if appTask.Strategy == "rolling-update" {
		message = fmt.Sprintf("The app '%s' is successfully updated", app.Name)
	} else {
		message = fmt.Sprintf("The app '%s' is being updated. "+
			"Confirm result by checking the app status after a while.", app.Name)
	}
	return message, nil
}

func (u *AppServeAppUsecase) PromoteAppServeApp(ctx context.Context, appId string) (ret string, err error) {
	app, err := u.repo.GetAppServeAppById(ctx, appId)
	if err != nil {
		return "", fmt.Errorf("error while getting ASA Info from DB. Err: %s", err)
	}

	if app.Status != "PROMOTE_WAIT" && app.Status != "PROMOTE_FAILED" {
		return "", fmt.Errorf("The app is not waiting for promote. Exiting..")
	}

	// Get the latest task ID so that the task status can be modified inside workflow once the promotion is done.
	latestTaskId := app.AppServeAppTasks[0].ID
	strategy := app.AppServeAppTasks[0].Strategy
	log.Info(ctx, "latestTaskId = ", latestTaskId)
	log.Info(ctx, "strategy = ", strategy)

	log.Info(ctx, "Updating app status to 'PROMOTING'..")

	err = u.repo.UpdateStatus(ctx, appId, latestTaskId, "PROMOTING", "")
	if err != nil {
		log.Debug(ctx, "appId = ", appId)
		log.Debug(ctx, "taskId = ", latestTaskId)
		return "", fmt.Errorf("failed to update app status on PromoteAppServeApp. Err: %s", err)
	}

	// Call argo workflow
	workflow := "promote-java-app"

	log.Info(ctx, "Submitting workflow: ", workflow)

	workflowId, err := u.argo.SumbitWorkflowFromWftpl(ctx, workflow, argowf.SubmitOptions{
		Parameters: []string{
			"organization_id=" + app.OrganizationId,
			"project_id=" + app.ProjectId,
			"target_cluster_id=" + app.TargetClusterId,
			"app_name=" + app.Name,
			"namespace=" + app.Namespace,
			"asa_id=" + app.ID,
			"asa_task_id=" + latestTaskId,
			"strategy=" + strategy,
			"tks_api_url=" + viper.GetString("external-address"),
		},
	})
	if err != nil {
		log.Error(ctx, "failed to submit workflow. Err:", err)
		return "", fmt.Errorf("failed to submit workflow. Err: %s", err)
	}
	log.Info(ctx, "Successfully submitted workflow: ", workflowId)

	return fmt.Sprintf("The app '%s' is being promoted. "+
		"Confirm result by checking the app status after a while.", app.Name), nil
}

func (u *AppServeAppUsecase) AbortAppServeApp(ctx context.Context, appId string) (ret string, err error) {
	app, err := u.repo.GetAppServeAppById(ctx, appId)
	if err != nil {
		return "", fmt.Errorf("error while getting ASA Info from DB. Err: %s", err)
	}

	if app.Status != "PROMOTE_WAIT" && app.Status != "ABORT_FAILED" {
		return "", fmt.Errorf("The app is not waiting for promote. Exiting..")
	}

	// Get the latest task ID so that the task status can be modified inside workflow once the abort process is done.
	latestTaskId := app.AppServeAppTasks[0].ID
	log.Info(ctx, "latestTaskId = ", latestTaskId)
	log.Info(ctx, "Updating app status to 'ABORTING'..")

	err = u.repo.UpdateStatus(ctx, appId, latestTaskId, "ABORTING", "")
	if err != nil {
		log.Debug(ctx, "appId = ", appId)
		log.Debug(ctx, "taskId = ", latestTaskId)
		return "", fmt.Errorf("failed to update app status on AbortAppServeApp. Err: %s", err)
	}

	// Call argo workflow
	workflow := "abort-java-app"

	log.Info(ctx, "Submitting workflow: ", workflow)

	// Call argo workflow
	workflowId, err := u.argo.SumbitWorkflowFromWftpl(ctx, workflow, argowf.SubmitOptions{
		Parameters: []string{
			"organization_id=" + app.OrganizationId,
			"project_id=" + app.ProjectId,
			"target_cluster_id=" + app.TargetClusterId,
			"app_name=" + app.Name,
			"namespace=" + app.Namespace,
			"asa_id=" + app.ID,
			"asa_task_id=" + latestTaskId,
			"tks_api_url=" + viper.GetString("external-address"),
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to submit workflow. Err: %s", err)
	}
	log.Info(ctx, "Successfully submitted workflow: ", workflowId)

	return fmt.Sprintf("The app '%s' is being promoted. "+
		"Confirm result by checking the app status after a while.", app.Name), nil
}

func (u *AppServeAppUsecase) RollbackAppServeApp(ctx context.Context, appId string, taskId string) (ret string, err error) {
	log.Info(ctx, "Starting rollback process..")

	app, err := u.repo.GetAppServeAppById(ctx, appId)
	if err != nil {
		return "", err
	}

	// Find target(dest) task
	var task model.AppServeAppTask
	for _, t := range app.AppServeAppTasks {
		if t.ID == taskId {
			task = t
			break
		}
	}

	// Save target version
	targetVer := task.Version
	targetRev := task.HelmRevision

	// Insert new values to the target task object
	task.ID = ""
	task.Output = ""
	task.Status = "ROLLBACKING"
	task.Version = strconv.Itoa(len(app.AppServeAppTasks) + 1)
	task.CreatedAt = time.Now()
	task.UpdatedAt = nil
	task.HelmRevision = 0
	task.RollbackVersion = targetVer

	// Creates new task record from the target task
	newTaskId, err := u.repo.CreateTask(ctx, &task)
	if err != nil {
		log.Info(ctx, "taskId = ", newTaskId)
		return "", fmt.Errorf("failed to rollback app-serve application. Err: %s", err)
	}

	log.Info(ctx, "Updating app status to 'ROLLBACKING'..")

	err = u.repo.UpdateStatus(ctx, appId, newTaskId, "ROLLBACKING", "")
	if err != nil {
		log.Debug(ctx, "appId = ", appId)
		log.Debug(ctx, "taskId = ", newTaskId)
		return "", fmt.Errorf("failed to update app status on RollbackAppServeApp. Err: %s", err)
	}

	// Call argo workflow
	workflow := "rollback-java-app"

	log.Info(ctx, "Submitting workflow: ", workflow)

	workflowId, err := u.argo.SumbitWorkflowFromWftpl(ctx, workflow, argowf.SubmitOptions{
		Parameters: []string{
			"organization_id=" + app.OrganizationId,
			"target_cluster_id=" + app.TargetClusterId,
			"app_name=" + app.Name,
			"namespace=" + app.Namespace,
			"asa_id=" + app.ID,
			"asa_task_id=" + newTaskId,
			"helm_revision=" + strconv.Itoa(int(targetRev)),
			"tks_api_url=" + viper.GetString("external-address"),
		},
	})
	if err != nil {
		log.Error(ctx, "Failed to submit workflow. Err:", err)
		return "", fmt.Errorf("failed to submit workflow. Err: %s", err)
	}
	log.Info(ctx, "Successfully submitted workflow: ", workflowId)

	return fmt.Sprintf("Rollback app Request '%v' is successfully submitted", taskId), nil
}
