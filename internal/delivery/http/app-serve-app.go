package http

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/openinfradev/tks-api/internal"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/serializer"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
)

var (
	StatusResult = map[string]string{
		"BUILDING":         "PROGRESS",
		"BUILD_SUCCESS":    "DONE",
		"BUILD_FAILED":     "FAILED",
		"DEPLOYING":        "PROGRESS",
		"DEPLOY_SUCCESS":   "DONE",
		"DEPLOY_FAILED":    "FAILED",
		"PROMOTE_WAIT":     "WAITING",
		"PROMOTING":        "PROGRESS",
		"PROMOTE_SUCCESS":  "DONE",
		"PROMOTE_FAILED":   "FAILED",
		"ABORTING":         "PROGRESS",
		"ABORT_SUCCESS":    "DONE",
		"ABORT_FAILED":     "FAILED",
		"ROLLBACKING":      "PROGRESS",
		"ROLLBACK_SUCCESS": "DONE",
		"ROLLBACK_FAILED":  "FAILED",
		"DELETING":         "PROGRESS",
		"DELETE_SUCCESS":   "DONE",
		"DELETE_FAILED":    "FAILED",
		"WAITING":          "WAITING",
	}
	StatusStages = map[string]map[string]string{
		"PREPARING":        {"build": "WAITING", "deploy": "WAITING", "promote": "WAITING"},
		"BUILDING":         {"build": "BUILDING", "deploy": "WAITING", "promote": "WAITING"},
		"BUILD_SUCCESS":    {"build": "BUILD_SUCCESS", "deploy": "WAITING", "promote": "WAITING"},
		"BUILD_FAILED":     {"build": "BUILD_FAILED", "deploy": "WAITING", "promote": "WAITING"},
		"DEPLOYING":        {"build": "BUILD_SUCCESS", "deploy": "DEPLOYING", "promote": "WAITING"},
		"DEPLOY_SUCCESS":   {"build": "BUILD_SUCCESS", "deploy": "DEPLOY_SUCCESS", "promote": "WAITING"},
		"DEPLOY_FAILED":    {"build": "BUILD_SUCCESS", "deploy": "DEPLOY_FAILED", "promote": "WAITING"},
		"PROMOTE_WAIT":     {"build": "BUILD_SUCCESS", "deploy": "DEPLOY_SUCCESS", "promote": "PROMOTE_WAIT"},
		"PROMOTING":        {"build": "BUILD_SUCCESS", "deploy": "DEPLOY_SUCCESS", "promote": "PROMOTING"},
		"PROMOTE_SUCCESS":  {"build": "BUILD_SUCCESS", "deploy": "DEPLOY_SUCCESS", "promote": "PROMOTE_SUCCESS"},
		"PROMOTE_FAILED":   {"build": "BUILD_SUCCESS", "deploy": "DEPLOY_SUCCESS", "promote": "PROMOTE_FAILED"},
		"ABORTING":         {"build": "BUILD_SUCCESS", "deploy": "DEPLOY_SUCCESS", "promote": "ABORTING"},
		"ABORT_SUCCESS":    {"build": "BUILD_SUCCESS", "deploy": "DEPLOY_SUCCESS", "promote": "ABORT_SUCCESS"},
		"ABORT_FAILED":     {"build": "BUILD_SUCCESS", "deploy": "DEPLOY_SUCCESS", "promote": "ABORT_FAILED"},
		"ROLLBACKING":      {"rollback": "ROLLBACKING"},
		"ROLLBACK_SUCCESS": {"rollback": "ROLLBACK_SUCCESS"},
		"ROLLBACK_FAILED":  {"rollback": "ROLLBACK_FAILED"},
		"DELETING":         {"delete": "DELETING"},
		"DELETE_SUCCESS":   {"delete": "DELETE_SUCCESS"},
		"DELETE_FAILED":    {"delete": "DELETE_FAILED"},
	}
)

type AppServeAppHandler struct {
	usecase    usecase.IAppServeAppUsecase
	prjUsecase usecase.IProjectUsecase
}

func NewAppServeAppHandler(u usecase.Usecase) *AppServeAppHandler {
	return &AppServeAppHandler{
		usecase:    u.AppServeApp,
		prjUsecase: u.Project,
	}
}

// CreateAppServeApp godoc
//
//	@Tags			AppServeApps
//	@Summary		Install appServeApp
//	@Description	Install appServeApp
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string							true	"Organization ID"
//	@Param			projectId		path		string							true	"Project ID"
//	@Param			object			body		domain.CreateAppServeAppRequest	true	"Request body to create app"
//	@Success		200				{object}	string
//	@Router			/organizations/{organizationId}/projects/{projectId}/app-serve-apps [post]
//	@Security		JWT
func (h *AppServeAppHandler) CreateAppServeApp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	log.Debugf(r.Context(), "organizationId = [%v]\n", organizationId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	projectId, ok := vars["projectId"]
	log.Debugf(r.Context(), "projectId = [%v]\n", projectId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid projectId"), "C_INVALID_PROJECT_ID", ""))
		return
	}

	appReq := domain.CreateAppServeAppRequest{}
	err := UnmarshalRequestInput(r, &appReq)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Error while unmarshalling request"), "C_INTERNAL_ERROR", ""))
		return
	}

	(appReq).SetDefaultValue()

	var app model.AppServeApp
	if err = serializer.Map(r.Context(), appReq, &app); err != nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
		return
	}

	log.Infof(r.Context(), "Processing CREATE request for app '%s'...", app.Name)

	// Namespace validation
	exist, err := h.prjUsecase.IsProjectNamespaceExist(r.Context(), organizationId, projectId, app.TargetClusterId, app.Namespace)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewInternalServerError(fmt.Errorf("Error while checking namespace record: %s", err), "", ""))
		return
	}

	if !exist {
		ErrorJSON(w, r, httpErrors.NewInternalServerError(fmt.Errorf("Namespace '%s' doesn't exist", app.Namespace), "", ""))
		return
	}

	now := time.Now()
	app.OrganizationId = organizationId
	app.ProjectId = projectId
	app.EndpointUrl = "N/A"
	app.PreviewEndpointUrl = "N/A"
	app.Status = "PREPARING"
	app.CreatedAt = now

	var task model.AppServeAppTask
	if err = serializer.Map(r.Context(), appReq, &task); err != nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
		return
	}

	task.Version = "1"
	task.Status = "PREPARING"
	task.Output = ""
	task.CreatedAt = now

	// Validate name param
	re, _ := regexp.Compile("^[a-z][a-z0-9-]*$")
	if !(re.MatchString(app.Name)) {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("error: name should consist of alphanumeric characters and hyphens only"), "", ""))
		return
	}

	exist, err = h.usecase.IsAppServeAppNameExist(r.Context(), organizationId, app.Name)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return
	}
	if exist {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("error: name '%s' already exists.", app.Name), "", ""))
		return
	}

	// Create namespace if it's not given by user
	if len(strings.TrimSpace(app.Namespace)) == 0 {
		// Check if the new namespace is already used in the target cluster
		ns := ""
		nsExist := true
		for nsExist {
			// Generate unique namespace based on name and random number
			src := rand.NewSource(time.Now().UnixNano())
			r1 := rand.New(src)
			ns = fmt.Sprintf("%s-%s", app.Name, strconv.Itoa(r1.Intn(10000)))

			nsExist, err = h.usecase.IsAppServeAppNamespaceExist(r.Context(), app.TargetClusterId, ns)
			if err != nil {
				ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
				return
			}
		}

		log.Infof(r.Context(), "Created new namespace: %s", ns)
		app.Namespace = ns
	} else {
		log.Infof(r.Context(), "Using existing namespace: %s", app.Namespace)
	}

	// Validate port param for springboot app
	if app.AppType == "springboot" {
		if task.Port == "" {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("error: 'port' param is mandatory"), "", ""))
			return
		}
	}

	// Validate 'strategy' param
	if task.Strategy != "rolling-update" {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(
			fmt.Errorf("error: 'strategy' should be 'rolling-update' on first deployment"), "", ""))
		return
	}

	_, _, err = h.usecase.CreateAppServeApp(r.Context(), &app, &task)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.CreateAppServeAppResponse
	if err = serializer.Map(r.Context(), app, &out); err != nil {
		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetAppServeApps godoc
//
//	@Tags			AppServeApps
//	@Summary		Get appServeApp list
//	@Description	Get appServeApp list by giving params
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string		true	"Organization ID"
//	@Param			projectId		path		string		true	"Project ID"
//	@Param			showAll			query		boolean		false	"Show all apps including deleted apps"
//	@Param			pageSize		query		string		false	"pageSize"
//	@Param			pageNumber		query		string		false	"pageNumber"
//	@Param			soertColumn		query		string		false	"sortColumn"
//	@Param			sortOrder		query		string		false	"sortOrder"
//	@Param			filters			query		[]string	false	"filters"
//	@Success		200				{object}	[]model.AppServeApp
//	@Router			/organizations/{organizationId}/projects/{projectId}/app-serve-apps [get]
//	@Security		JWT
func (h *AppServeAppHandler) GetAppServeApps(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	organizationId, ok := vars["organizationId"]
	log.Debugf(r.Context(), "organizationId = [%v]\n", organizationId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	projectId, ok := vars["projectId"]
	log.Debugf(r.Context(), "projectId = [%v]\n", projectId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid projectId"), "C_INVALID_PROJECT_ID", ""))
		return
	}

	urlParams := r.URL.Query()

	showAllParam := urlParams.Get("showAll")
	if showAllParam == "" {
		showAllParam = "false"
	}

	showAll, err := strconv.ParseBool(showAllParam)
	if err != nil {
		log.Error(r.Context(), "Failed to convert showAll params. Err: ", err)
		ErrorJSON(w, r, err)
		return
	}

	pg := pagination.NewPagination(&urlParams)
	apps, err := h.usecase.GetAppServeApps(r.Context(), organizationId, projectId, showAll, pg)
	if err != nil {
		log.Error(r.Context(), "Failed to get Failed to get app-serve-apps ", err)
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetAppServeAppsResponse
	out.AppServeApps = make([]domain.AppServeAppResponse, len(apps))
	for i, app := range apps {
		if err := serializer.Map(r.Context(), app, &out.AppServeApps[i]); err != nil {
			log.Info(r.Context(), err)
			continue
		}
	}

	if out.Pagination, err = pg.Response(r.Context()); err != nil {
		log.Info(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetAppServeAppLatestTask godoc
//
//	@Tags			AppServeApps
//	@Summary		Get latest task from appServeApp
//	@Description	Get latest task from appServeApp
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"Organization ID"
//	@Param			projectId		path		string	true	"Project ID"
//	@Param			appId			path		string	true	"App ID"
//	@Success		200				{object}	domain.GetAppServeAppTaskResponse
//	@Router			/organizations/{organizationId}/projects/{projectId}/app-serve-apps/{appId}/latest-task [get]
//	@Security		JWT
func (h *AppServeAppHandler) GetAppServeAppLatestTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	organizationId, ok := vars["organizationId"]
	fmt.Printf("organizationId = [%v]\n", organizationId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "", ""))
		return
	}

	projectId, ok := vars["projectId"]
	log.Debugf(r.Context(), "projectId = [%v]\n", projectId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid projectId: [%s]", projectId), "C_INVALID_PROJECT_ID", ""))
		return
	}

	appId, ok := vars["appId"]
	fmt.Printf("appId = [%s]\n", appId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid appId"), "", ""))
		return
	}

	// Check if projectId exists
	prj, err := h.prjUsecase.GetProject(r.Context(), organizationId, projectId)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewInternalServerError(fmt.Errorf("Error while checking project record: %s", err), "", ""))
		return
	} else if prj == nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("projectId not found: %s", projectId), "C_INVALID_PROJECT_ID", ""))
	}

	task, err := h.usecase.GetAppServeAppLatestTask(r.Context(), appId)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return
	}
	if task == nil {
		ErrorJSON(w, r, httpErrors.NewNoContentError(fmt.Errorf("No task exists"), "", ""))
		return
	}
	// Rollbacking to latest task should be blocked.
	task.AvailableRollback = false

	app, err := h.usecase.GetAppServeAppById(r.Context(), appId)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return
	}
	if app == nil {
		ErrorJSON(w, r, httpErrors.NewNoContentError(fmt.Errorf("No app exists"), "D_NO_ASA", ""))
		return
	}

	var out domain.GetAppServeAppTaskResponse
	if err := serializer.Map(r.Context(), *app, &out.AppServeApp); err != nil {
		log.Info(r.Context(), err)
	}
	if err := serializer.Map(r.Context(), *task, &out.AppServeAppTask); err != nil {
		log.Info(r.Context(), err)
	}

	out.Stages = makeStages(r.Context(), task, app)

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetNumOfAppsOnStack godoc
//
//	@Tags			AppServeApps
//	@Summary		Get number of apps on given stack
//	@Description	Get number of apps on given stack
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"Organization ID"
//	@Param			projectId		path		string	true	"Project ID"
//	@Param			stackId			query		string	true	"Stack ID"
//	@Success		200				{object}	int64
//	@Router			/organizations/{organizationId}/projects/{projectId}/app-serve-apps/count [get]
//	@Security		JWT
func (h *AppServeAppHandler) GetNumOfAppsOnStack(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	organizationId, ok := vars["organizationId"]
	log.Debugf(r.Context(), "organizationId = [%s]\n", organizationId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "", ""))
		return
	}

	urlParams := r.URL.Query()
	stackId := urlParams.Get("stackId")
	if stackId == "" {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("StackId must be provided."), "", ""))
	}
	fmt.Printf("stackId = [%s]\n", stackId)

	numApps, err := h.usecase.GetNumOfAppsOnStack(r.Context(), organizationId, stackId)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return
	}

	ResponseJSON(w, r, http.StatusOK, numApps)
}

// GetAppServeAppTasksByAppId godoc
//
//	@Tags			AppServeApps
//	@Summary		Get appServeAppTask list
//	@Description	Get appServeAppTask list by giving params
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string		true	"Organization ID"
//	@Param			projectId		path		string		true	"Project ID"
//	@Param			pageSize		query		string		false	"pageSize"
//	@Param			pageNumber		query		string		false	"pageNumber"
//	@Param			sortColumn		query		string		false	"sortColumn"
//	@Param			sortOrder		query		string		false	"sortOrder"
//	@Param			filters			query		[]string	false	"filters"
//	@Success		200				{object}	[]model.AppServeApp
//	@Router			/organizations/{organizationId}/projects/{projectId}/app-serve-apps/{appId}/tasks [get]
//	@Security		JWT
func (h *AppServeAppHandler) GetAppServeAppTasksByAppId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId: %s", organizationId), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	projectId, ok := vars["projectId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid projectId: %s", projectId), "C_INVALID_PROJECT_ID", ""))
		return
	}

	appId, ok := vars["appId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid appId: %s", appId), "C_INVALID_ASA_ID", ""))
		return
	}

	urlParams := r.URL.Query()
	pg := pagination.NewPagination(&urlParams)

	// Check if projectId exists
	prj, err := h.prjUsecase.GetProject(r.Context(), organizationId, projectId)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewInternalServerError(fmt.Errorf("Error while checking project record: %v", err), "", ""))
		return
	} else if prj == nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("projectId not found: %s", projectId), "C_INVALID_PROJECT_ID", ""))
	}

	tasks, err := h.usecase.GetAppServeAppTasks(r.Context(), appId, pg)
	if err != nil {
		log.Error(r.Context(), "Failed to get app-serve-app-tasks ", err)
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetAppServeAppTasksResponse
	out.AppServeAppTasks = make([]domain.AppServeAppTaskResponse, len(tasks))
	for i, task := range tasks {
		// Rollbacking to latest task should be blocked.
		if i > 0 && strings.Contains(task.Status, "SUCCESS") && task.Status != "ABORT_SUCCESS" &&
			task.Status != "ROLLBACK_SUCCESS" {
			task.AvailableRollback = true
		}

		if err := serializer.Map(r.Context(), task, &out.AppServeAppTasks[i]); err != nil {
			log.Info(r.Context(), err)
			continue
		}
	}

	if out.Pagination, err = pg.Response(r.Context()); err != nil {
		log.Info(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetAppServeAppTaskDetail godoc
//
//	@Tags			AppServeApps
//	@Summary		Get task detail from appServeApp
//	@Description	Get task detail from appServeApp
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"Organization ID"
//	@Param			projectId		path		string	true	"Project ID"
//	@Param			appId			path		string	true	"App ID"
//	@Param			taskId			path		string	true	"Task ID"
//	@Success		200				{object}	domain.GetAppServeAppTaskResponse
//	@Router			/organizations/{organizationId}/projects/{projectId}/app-serve-apps/{appId}/tasks/{taskId} [get]
//	@Security		JWT
func (h *AppServeAppHandler) GetAppServeAppTaskDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	organizationId, ok := vars["organizationId"]
	log.Debugf(r.Context(), "organizationId = [%v]\n", organizationId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId: [%s]", organizationId), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	projectId, ok := vars["projectId"]
	log.Debugf(r.Context(), "projectId = [%v]\n", projectId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid projectId: [%s]", projectId), "C_INVALID_PROJECT_ID", ""))
		return
	}

	appId, ok := vars["appId"]
	log.Debugf(r.Context(), "appId = [%s]\n", appId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid appId: [%s]", appId), "C_INVALID_ASA_ID", ""))
		return
	}

	taskId, ok := vars["taskId"]
	log.Debugf(r.Context(), "taskId = [%s]\n", taskId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid taskId: [%s]", taskId), "C_INVALID_ASA_TASK_ID", ""))
		return
	}

	// Check if projectId exists
	prj, err := h.prjUsecase.GetProject(r.Context(), organizationId, projectId)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewInternalServerError(fmt.Errorf("Error while checking project record: %s", err), "", ""))
		return
	} else if prj == nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("projectId not found: %s", projectId), "C_INVALID_PROJECT_ID", ""))
	}

	task, err := h.usecase.GetAppServeAppTaskById(r.Context(), taskId)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return
	}
	if task == nil {
		ErrorJSON(w, r, httpErrors.NewNoContentError(fmt.Errorf("No task exists"), "", ""))
		return
	}

	app, err := h.usecase.GetAppServeAppById(r.Context(), appId)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return
	}
	if app == nil {
		ErrorJSON(w, r, httpErrors.NewNoContentError(fmt.Errorf("No app exists"), "D_NO_ASA", ""))
		return
	}

	// TODO: this should be false for latest task
	if strings.Contains(task.Status, "SUCCESS") && task.Status != "ABORT_SUCCESS" &&
		task.Status != "ROLLBACK_SUCCESS" {
		task.AvailableRollback = true
	}

	var out domain.GetAppServeAppTaskResponse
	if err := serializer.Map(r.Context(), *app, &out.AppServeApp); err != nil {
		log.Info(r.Context(), err)
	}
	if err := serializer.Map(r.Context(), *task, &out.AppServeAppTask); err != nil {
		log.Info(r.Context(), err)
	}

	out.Stages = makeStages(r.Context(), task, app)

	ResponseJSON(w, r, http.StatusOK, out)
}

func makeStages(ctx context.Context, task *model.AppServeAppTask, app *model.AppServeApp) []domain.StageResponse {
	stages := make([]domain.StageResponse, 0)

	var stage domain.StageResponse
	var pipelines []string
	taskStatus := task.Status
	strategy := task.Strategy

	if taskStatus == "ROLLBACKING" ||
		taskStatus == "ROLLBACK_SUCCESS" ||
		taskStatus == "ROLLBACK_FAILED" {
		pipelines = []string{"rollback"}
	} else if taskStatus == "DELETING" ||
		taskStatus == "DELETE_SUCCESS" ||
		taskStatus == "DELETE_FAILED" {
		pipelines = []string{"delete"}
	} else if app.Type == "all" {
		if strategy == "rolling-update" {
			pipelines = []string{"build", "deploy"}
		} else if strategy == "blue-green" {
			pipelines = []string{"build", "deploy", "promote"}
		}
	} else if app.Type == "deploy" {
		if strategy == "rolling-update" {
			pipelines = []string{"deploy"}
		} else if strategy == "blue-green" {
			pipelines = []string{"deploy", "promote"}
		}
	} else {
		log.Error(ctx, "Unexpected case happened while making stages!")
	}

	fmt.Printf("Pipeline stages: %v\n", pipelines)

	for _, pl := range pipelines {
		stage = makeStage(ctx, task, app, pl)
		stages = append(stages, stage)
	}

	return stages
}

func makeStage(ctx context.Context, task *model.AppServeAppTask, app *model.AppServeApp, pl string) domain.StageResponse {
	taskStatus := task.Status
	strategy := task.Strategy

	stage := domain.StageResponse{
		Name:   pl,
		Status: StatusStages[taskStatus][pl],
		Result: StatusResult[StatusStages[taskStatus][pl]],
	}

	var actions []domain.ActionResponse
	if stage.Status == "DEPLOY_SUCCESS" {
		if strategy == "blue-green" {
			if taskStatus == "PROMOTE_WAIT" {
				action := domain.ActionResponse{
					Name: "OLD_EP",
					Uri:  app.EndpointUrl,
					Type: "LINK",
				}
				actions = append(actions, action)

				action = domain.ActionResponse{
					Name: "NEW_EP",
					Uri:  app.PreviewEndpointUrl,
					Type: "LINK",
				}
				actions = append(actions, action)
			}
		}

	} else if stage.Status == "PROMOTE_WAIT" && strategy == "blue-green" {
		action := domain.ActionResponse{
			Name: "ABORT",
			Uri: fmt.Sprintf(internal.API_PREFIX+internal.API_VERSION+
				"/organizations/%v/projects/%v/app-serve-apps/%v", app.OrganizationId, app.ProjectId, app.ID),
			Type:   "API",
			Method: "PUT",
			Body:   map[string]string{"strategy": "blue-green", "abort": "true"},
		}
		actions = append(actions, action)

		action = domain.ActionResponse{
			Name: "PROMOTE",
			Uri: fmt.Sprintf(internal.API_PREFIX+internal.API_VERSION+
				"/organizations/%v/projects/%v/app-serve-apps/%v", app.OrganizationId, app.ProjectId, app.ID),
			Type:   "API",
			Method: "PUT",
			Body:   map[string]string{"strategy": "blue-green", "promote": "true"},
		}
		actions = append(actions, action)
	}

	stage.Actions = &actions

	return stage
}

// IsAppServeAppExist godoc
//
//	@Tags			AppServeApps
//	@Summary		Get appServeApp
//	@Description	Get appServeApp by giving params
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"Organization ID"
//	@Param			projectId		path		string	true	"Project ID"
//	@Success		200				{object}	bool
//	@Router			/organizations/{organizationId}/projects/{projectId}/app-serve-apps/{appId}/exist [get]
//	@Security		JWT
func (h *AppServeAppHandler) IsAppServeAppExist(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	organizationId, ok := vars["organizationId"]
	fmt.Printf("organizationId = [%v]\n", organizationId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	urlParams := r.URL.Query()
	appId := urlParams.Get("appId")
	if appId == "" {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid appId"), "C_INVALID_ASA_ID", ""))
		return
	}

	exist, err := h.usecase.IsAppServeAppExist(r.Context(), appId)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out = struct {
		Exist bool `json:"exist"`
	}{Exist: exist}

	ResponseJSON(w, r, http.StatusOK, out)
}

// IsAppServeAppNameExist godoc
//
//	@Tags			AppServeApps
//	@Summary		Check duplicate appServeAppName
//	@Description	Check duplicate appServeAppName by giving params
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"Organization ID"
//	@Param			projectId		path		string	true	"Project ID"
//	@Param			name			path		string	true	"name"
//	@Success		200				{object}	bool
//	@Router			/organizations/{organizationId}/projects/{projectId}/app-serve-apps/name/{name}/existence [get]
//	@Security		JWT
func (h *AppServeAppHandler) IsAppServeAppNameExist(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	organizationId, ok := vars["organizationId"]
	fmt.Printf("organizationId = [%v]\n", organizationId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}
	appName, ok := vars["name"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid appName"), "C_INVALID_ASA_ID", ""))
		return
	}

	existed, err := h.usecase.IsAppServeAppNameExist(r.Context(), organizationId, appName)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out = struct {
		Existed bool `json:"existed"`
	}{Existed: existed}

	ResponseJSON(w, r, http.StatusOK, out)
}

// UpdateAppServeApp godoc
//
//	@Tags			AppServeApps
//	@Summary		Update appServeApp
//	@Description	Update appServeApp
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string							true	"Organization ID"
//	@Param			projectId		path		string							true	"Project ID"
//	@Param			appId			path		string							true	"App ID"
//	@Param			object			body		domain.UpdateAppServeAppRequest	true	"Request body to update app"
//	@Success		200				{object}	string
//	@Router			/organizations/{organizationId}/projects/{projectId}/app-serve-apps/{appId} [put]
//	@Security		JWT
func (h *AppServeAppHandler) UpdateAppServeApp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	fmt.Printf("organizationId = [%v]\n", organizationId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	appId, ok := vars["appId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid appId"), "C_INVALID_ASA_ID", ""))
		return
	}

	// Get latest task
	latestTask, err := h.usecase.GetAppServeAppLatestTask(r.Context(), appId)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return
	}
	if latestTask == nil {
		ErrorJSON(w, r, httpErrors.NewNoContentError(fmt.Errorf("No task exists"), "", ""))
		return
	}

	// unmarshal request that only contains task-specific params
	appReq := domain.UpdateAppServeAppRequest{}
	err = UnmarshalRequestInput(r, &appReq)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var task model.AppServeAppTask
	if err = serializer.Map(r.Context(), *latestTask, &task); err != nil {
		//ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
		return
	}

	if err = serializer.Map(r.Context(), appReq, &task); err != nil {
		//ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
		return
	}

	// Set new version
	verInt, err := strconv.Atoi(latestTask.Version)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}
	newVerStr := strconv.Itoa(verInt + 1)

	task.Version = newVerStr
	//task.AppServeAppId = app.ID
	task.Status = "PREPARING"
	task.RollbackVersion = ""
	task.Output = ""
	task.CreatedAt = time.Now()
	task.UpdatedAt = nil

	log.Debugf(r.Context(), "New task in update request: %v\n", task)

	var res string
	if appReq.Promote {
		res, err = h.usecase.PromoteAppServeApp(r.Context(), appId)
	} else if appReq.Abort {
		res, err = h.usecase.AbortAppServeApp(r.Context(), appId)
	} else {
		res, err = h.usecase.UpdateAppServeApp(r.Context(), appId, &task)
	}

	if err != nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
		return
	}

	ResponseJSON(w, r, http.StatusOK, res)
}

// UpdateAppServeAppStatus godoc
//
//	@Tags			AppServeApps
//	@Summary		Update app status
//	@Description	Update app status
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string									true	"Organization ID"
//	@Param			projectId		path		string									true	"Project ID"
//	@Param			appId			path		string									true	"App ID"
//	@Param			body			body		domain.UpdateAppServeAppStatusRequest	true	"Request body to update app status"
//	@Success		200				{object}	string
//	@Router			/organizations/{organizationId}/projects/{projectId}/app-serve-apps/{appId}/status [patch]
//	@Security		JWT
func (h *AppServeAppHandler) UpdateAppServeAppStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	fmt.Printf("organizationId = [%v]\n", organizationId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	appId, ok := vars["appId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid appId"), "C_INVALID_ASA_ID", ""))
		return
	}

	appStatusReq := domain.UpdateAppServeAppStatusRequest{}
	err := UnmarshalRequestInput(r, &appStatusReq)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	res, err := h.usecase.UpdateAppServeAppStatus(r.Context(), appId, appStatusReq.TaskID, appStatusReq.Status, appStatusReq.Output)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
		return
	}

	ResponseJSON(w, r, http.StatusOK, res)
}

// UpdateAppServeAppEndpoint godoc
//
//	@Tags			AppServeApps
//	@Summary		Update app endpoint
//	@Description	Update app endpoint
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string									true	"Organization ID"
//	@Param			projectId		path		string									true	"Project ID"
//	@Param			appId			path		string									true	"appId"
//	@Param			body			body		domain.UpdateAppServeAppEndpointRequest	true	"Request body to update app endpoint"
//	@Success		200				{object}	string
//	@Router			/organizations/{organizationId}/projects/{projectId}/app-serve-apps/{appId}/endpoint [patch]
//	@Security		JWT
func (h *AppServeAppHandler) UpdateAppServeAppEndpoint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	fmt.Printf("organizationId = [%v]\n", organizationId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	appId, ok := vars["appId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid appId"), "C_INVALID_ASA_ID", ""))
		return
	}

	appReq := domain.UpdateAppServeAppEndpointRequest{}
	err := UnmarshalRequestInput(r, &appReq)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	res, err := h.usecase.UpdateAppServeAppEndpoint(
		r.Context(),
		appId,
		appReq.TaskID,
		appReq.EndpointUrl,
		appReq.PreviewEndpointUrl,
		appReq.HelmRevision)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
		return
	}

	ResponseJSON(w, r, http.StatusOK, res)
}

// DeleteAppServeApp godoc
//
//	@Tags			AppServeApps
//	@Summary		Uninstall appServeApp
//	@Description	Uninstall appServeApp
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"Organization ID"
//	@Param			projectId		path		string	true	"Project ID"
//	@Param			appId			path		string	true	"App ID"
//	@Success		200				{object}	string
//	@Router			/organizations/{organizationId}/projects/{projectId}/app-serve-apps/{appId} [delete]
//	@Security		JWT
func (h *AppServeAppHandler) DeleteAppServeApp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	fmt.Printf("organizationId = [%v]\n", organizationId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	appId, ok := vars["appId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid appId"), "C_INVALID_ASA_ID", ""))
		return
	}

	res, err := h.usecase.DeleteAppServeApp(r.Context(), appId)
	if err != nil {
		log.Error(r.Context(), "Failed to delete appId err : ", err)
		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, res)
}

// RollbackAppServeApp godoc
//
//	@Tags			AppServeApps
//	@Summary		Rollback appServeApp
//	@Description	Rollback appServeApp
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string								true	"Organization ID"
//	@Param			projectId		path		string								true	"Project ID"
//	@Param			appId			path		string								true	"App ID"
//	@Param			object			body		domain.RollbackAppServeAppRequest	true	"Request body to rollback app"
//	@Success		200				{object}	string
//	@Router			/organizations/{organizationId}/projects/{projectId}/app-serve-apps/{appId}/rollback [post]
//	@Security		JWT
func (h *AppServeAppHandler) RollbackAppServeApp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	fmt.Printf("organizationId = [%v]\n", organizationId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	appId, ok := vars["appId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid appId"), "C_INVALID_ASA_ID", ""))
		return
	}

	appReq := domain.RollbackAppServeAppRequest{}
	err := UnmarshalRequestInput(r, &appReq)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}
	if appReq.TaskId == "" {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("no taskId"), "C_INVALID_ASA_TASK_ID", ""))
		return
	}

	res, err := h.usecase.RollbackAppServeApp(r.Context(), appId, appReq.TaskId)

	if err != nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
		return
	}

	ResponseJSON(w, r, http.StatusOK, res)
}
