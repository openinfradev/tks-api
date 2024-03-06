package http

import (
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/openinfradev/tks-api/internal"
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
	usecase usecase.IAppServeAppUsecase
}

func NewAppServeAppHandler(h usecase.Usecase) *AppServeAppHandler {
	return &AppServeAppHandler{
		usecase: h.AppServeApp,
	}
}

// CreateAppServeApp godoc
//	@Tags			AppServeApps
//	@Summary		Install appServeApp
//	@Description	Install appServeApp
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string							true	"Organization ID"
//	@Param			projectId		path		string							true	"Project ID"
//	@Param			object			body		domain.CreateAppServeAppRequest	true	"Request body to create app"
//	@Success		200				{object}	string
//	@Router			/api/1.0/organizations/{organizationId}/projects/{projectId}/app-serve-apps [post]
//	@Security		JWT
func (h *AppServeAppHandler) CreateAppServeApp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	log.Debugf("organizationId = [%v]\n", organizationId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	projectId, ok := vars["projectId"]
	log.Debugf("projectId = [%v]\n", projectId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid projectId"), "C_INVALID_PROJECT_ID", ""))
		return
	}

	appReq := domain.CreateAppServeAppRequest{}
	err := UnmarshalRequestInput(r, &appReq)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	(appReq).SetDefaultValue()

	var app domain.AppServeApp
	if err = serializer.Map(appReq, &app); err != nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
		return
	}

	log.Infof("Processing CREATE request for app '%s'...", app.Name)

	now := time.Now()
	app.OrganizationId = organizationId
	app.ProjectId = projectId
	app.EndpointUrl = "N/A"
	app.PreviewEndpointUrl = "N/A"
	app.Status = "PREPARING"
	app.CreatedAt = now

	var task domain.AppServeAppTask
	if err = serializer.Map(appReq, &task); err != nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
		return
	}

	task.Version = "1"
	task.Status = "PREPARING"
	task.Output = ""
	task.CreatedAt = now

	app.AppServeAppTasks = append(app.AppServeAppTasks, task)

	// Validate name param
	re, _ := regexp.Compile("^[a-z][a-z0-9-]*$")
	if !(re.MatchString(app.Name)) {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("error: name should consist of alphanumeric characters and hyphens only"), "", ""))
		return
	}

	exist, err := h.usecase.IsAppServeAppNameExist(organizationId, app.Name)
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

			nsExist, err = h.usecase.IsAppServeAppNamespaceExist(app.TargetClusterId, ns)
			if err != nil {
				ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
				return
			}
		}

		log.Infof("Created new namespace: %s", ns)
		app.Namespace = ns
	} else {
		log.Infof("Using existing namespace: %s", app.Namespace)
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

	_, _, err = h.usecase.CreateAppServeApp(&app)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.CreateAppServeAppResponse
	if err = serializer.Map(app, &out); err != nil {
		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetAppServeApps godoc
//	@Tags			AppServeApps
//	@Summary		Get appServeApp list
//	@Description	Get appServeApp list by giving params
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string		true	"Organization ID"
//	@Param			projectId		path		string		true	"Project ID"
//	@Param			showAll			query		boolean		false	"Show all apps including deleted apps"
//	@Param			limit			query		string		false	"pageSize"
//	@Param			page			query		string		false	"pageNumber"
//	@Param			soertColumn		query		string		false	"sortColumn"
//	@Param			sortOrder		query		string		false	"sortOrder"
//	@Param			filters			query		[]string	false	"filters"
//	@Success		200				{object}	[]domain.AppServeApp
//	@Router			/api/1.0/organizations/{organizationId}/projects/{projectId}/app-serve-apps [get]
//	@Security		JWT
func (h *AppServeAppHandler) GetAppServeApps(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	fmt.Printf("organizationId = [%v]\n", organizationId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	urlParams := r.URL.Query()

	showAllParam := urlParams.Get("showAll")
	if showAllParam == "" {
		showAllParam = "false"
	}

	showAll, err := strconv.ParseBool(showAllParam)
	if err != nil {
		log.ErrorWithContext(r.Context(), "Failed to convert showAll params. Err: ", err)
		ErrorJSON(w, r, err)
		return
	}
	pg := pagination.NewPagination(&urlParams)
	apps, err := h.usecase.GetAppServeApps(organizationId, showAll, pg)
	if err != nil {
		log.ErrorWithContext(r.Context(), "Failed to get Failed to get app-serve-apps ", err)
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetAppServeAppsResponse
	out.AppServeApps = apps

	if out.Pagination, err = pg.Response(); err != nil {
		log.InfoWithContext(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetAppServeApp godoc
//	@Tags			AppServeApps
//	@Summary		Get appServeApp
//	@Description	Get appServeApp by giving params
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"Organization ID"
//	@Param			projectId		path		string	true	"Project ID"
//	@Param			appId			path		string	true	"App ID"
//	@Success		200				{object}	domain.GetAppServeAppResponse
//	@Router			/api/1.0/organizations/{organizationId}/projects/{projectId}/app-serve-apps/{appId} [get]
//	@Security		JWT
func (h *AppServeAppHandler) GetAppServeApp(w http.ResponseWriter, r *http.Request) {
	//////////////////////////////////////////////////////////////////////////////////////////
	// TODO: this API will'be deprecated soon once the new task-related API's are verified.
	// Until then, this is available (except for stage info) just for backward compatibility.
	//////////////////////////////////////////////////////////////////////////////////////////

	vars := mux.Vars(r)

	organizationId, ok := vars["organizationId"]
	fmt.Printf("organizationId = [%v]\n", organizationId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	appId, ok := vars["appId"]
	fmt.Printf("appId = [%s]\n", appId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid appId"), "C_INVALID_ASA_ID", ""))
		return
	}
	app, err := h.usecase.GetAppServeAppById(appId)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return
	}
	if app == nil {
		ErrorJSON(w, r, httpErrors.NewNoContentError(fmt.Errorf("no appId"), "D_NO_ASA", ""))
		return
	}

	newTasks := make([]domain.AppServeAppTask, 0)

	for idx, t := range app.AppServeAppTasks {
		// Rollbacking to latest task should be blocked.
		if idx > 0 && strings.Contains(t.Status, "SUCCESS") && t.Status != "ABORT_SUCCESS" &&
			t.Status != "ROLLBACK_SUCCESS" {
			t.AvailableRollback = true
		}
		newTasks = append(newTasks, t)
	}
	app.AppServeAppTasks = newTasks

	var out domain.GetAppServeAppResponse
	out.AppServeApp = *app
	// NOTE: makeStages function's been changed to use task instead of app
	//out.Stages = makeStages(app)

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetAppServeAppLatestTask godoc
//	@Tags			AppServeApps
//	@Summary		Get latest task from appServeApp
//	@Description	Get latest task from appServeApp
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"Organization ID"
//	@Param			projectId		path		string	true	"Project ID"
//	@Param			appId			path		string	true	"App ID"
//	@Success		200				{object}	domain.GetAppServeAppTaskResponse
//	@Router			/api/1.0/organizations/{organizationId}/projects/{projectId}/app-serve-apps/{appId}/latest-task [get]
//	@Security		JWT
func (h *AppServeAppHandler) GetAppServeAppLatestTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	organizationId, ok := vars["organizationId"]
	fmt.Printf("organizationId = [%v]\n", organizationId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "", ""))
		return
	}

	appId, ok := vars["appId"]
	fmt.Printf("appId = [%s]\n", appId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid appId"), "", ""))
		return
	}
	task, err := h.usecase.GetAppServeAppLatestTask(appId)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return
	}
	if task == nil {
		ErrorJSON(w, r, httpErrors.NewNoContentError(fmt.Errorf("no task exists"), "", ""))
		return
	}

	var out domain.GetAppServeAppTaskResponse
	out.AppServeAppTask = *task

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetNumOfAppsOnStack godoc
//	@Tags			AppServeApps
//	@Summary		Get number of apps on given stack
//	@Description	Get number of apps on given stack
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"Organization ID"
//	@Param			projectId		path		string	true	"Project ID"
//	@Param			stackId			query		string	true	"Stack ID"
//	@Success		200				{object}	int64
//	@Router			/api/1.0/organizations/{organizationId}/projects/{projectId}/app-serve-apps/count [get]
//	@Security		JWT
func (h *AppServeAppHandler) GetNumOfAppsOnStack(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	organizationId, ok := vars["organizationId"]
	log.Debugf("organizationId = [%s]\n", organizationId)
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

	numApps, err := h.usecase.GetNumOfAppsOnStack(organizationId, stackId)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return
	}

	ResponseJSON(w, r, http.StatusOK, numApps)
}

// GetAppServeAppTasksByAppId godoc
//	@Tags			AppServeApps
//	@Summary		Get appServeAppTask list
//	@Description	Get appServeAppTask list by giving params
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string		true	"Organization ID"
//	@Param			projectId		path		string		true	"Project ID"
//	@Param			limit			query		string		false	"pageSize"
//	@Param			page			query		string		false	"pageNumber"
//	@Param			sortColumn		query		string		false	"sortColumn"
//	@Param			sortOrder		query		string		false	"sortOrder"
//	@Param			filters			query		[]string	false	"filters"
//	@Success		200				{object}	[]domain.AppServeApp
//	@Router			/api/1.0/organizations/{organizationId}/projects/{projectId}/app-serve-apps/{appId}/tasks [get]
//	@Security		JWT
func (h *AppServeAppHandler) GetAppServeAppTasksByAppId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId: %s", organizationId), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	appId, ok := vars["appId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid appId: %s", appId), "C_INVALID_ASA_ID", ""))
		return
	}

	urlParams := r.URL.Query()
	pg := pagination.NewPagination(&urlParams)

	tasks, err := h.usecase.GetAppServeAppTasks(appId, pg)
	if err != nil {
		log.ErrorWithContext(r.Context(), "Failed to get app-serve-app-tasks ", err)
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetAppServeAppTasksResponse
	out.AppServeAppTasks = tasks

	if out.Pagination, err = pg.Response(); err != nil {
		log.InfoWithContext(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetAppServeAppTaskDetail godoc
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
//	@Router			/api/1.0/organizations/{organizationId}/projects/{projectId}/app-serve-apps/{appId}/tasks/{taskId} [get]
//	@Security		JWT
func (h *AppServeAppHandler) GetAppServeAppTaskDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	organizationId, ok := vars["organizationId"]
	log.Debugf("organizationId = [%v]\n", organizationId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId: [%s]", organizationId), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	projectId, ok := vars["projectId"]
	log.Debugf("projectId = [%v]\n", projectId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid projectId: [%s]", projectId), "C_INVALID_PROJECT_ID", ""))
		return
	}

	appId, ok := vars["appId"]
	log.Debugf("appId = [%s]\n", appId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid appId: [%s]", appId), "C_INVALID_ASA_ID", ""))
		return
	}

	taskId, ok := vars["taskId"]
	log.Debugf("taskId = [%s]\n", taskId)
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid taskId: [%s]", taskId), "C_INVALID_ASA_TASK_ID", ""))
		return
	}

	task, app, err := h.usecase.GetAppServeAppTaskById(taskId)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return
	}
	if task == nil {
		ErrorJSON(w, r, httpErrors.NewNoContentError(fmt.Errorf("No task exists"), "", ""))
		return
	}

	if strings.Contains(task.Status, "SUCCESS") && task.Status != "ABORT_SUCCESS" &&
		task.Status != "ROLLBACK_SUCCESS" {
		task.AvailableRollback = true
	}

	var out domain.GetAppServeAppTaskResponse
	out.AppServeApp = *app
	out.AppServeAppTask = *task
	out.Stages = makeStages(task, app)

	ResponseJSON(w, r, http.StatusOK, out)
}

func makeStages(task *domain.AppServeAppTask, app *domain.AppServeApp) []domain.StageResponse {
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
		log.Error("Unexpected case happened while making stages!")
	}

	fmt.Printf("Pipeline stages: %v\n", pipelines)

	for _, pl := range pipelines {
		stage = makeStage(task, app, pl)
		stages = append(stages, stage)
	}

	return stages
}

func makeStage(task *domain.AppServeAppTask, app *domain.AppServeApp, pl string) domain.StageResponse {
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
		} else {
			log.Error("Not supported strategy!")
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
//	@Tags			AppServeApps
//	@Summary		Get appServeApp
//	@Description	Get appServeApp by giving params
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"Organization ID"
//	@Param			projectId		path		string	true	"Project ID"
//	@Success		200				{object}	bool
//	@Router			/api/1.0/organizations/{organizationId}/projects/{projectId}/app-serve-apps/{appId}/exist [get]
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

	exist, err := h.usecase.IsAppServeAppExist(appId)
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
//	@Tags			AppServeApps
//	@Summary		Check duplicate appServeAppName
//	@Description	Check duplicate appServeAppName by giving params
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"Organization ID"
//	@Param			projectId		path		string	true	"Project ID"
//	@Param			name			path		string	true	"name"
//	@Success		200				{object}	bool
//	@Router			/api/1.0/organizations/{organizationId}/projects/{projectId}/app-serve-apps/name/{name}/existence [get]
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

	existed, err := h.usecase.IsAppServeAppNameExist(organizationId, appName)
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
//	@Router			/api/1.0/organizations/{organizationId}/projects/{projectId}/app-serve-apps/{appId} [put]
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

	app, err := h.usecase.GetAppServeAppById(appId)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}
	if len(app.AppServeAppTasks) < 1 {
		ErrorJSON(w, r, err)
	}

	// unmarshal request that only contains task-specific params
	appReq := domain.UpdateAppServeAppRequest{}
	err = UnmarshalRequestInput(r, &appReq)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	// Instead of setting default value, some fields should be retrieved
	// from existing app config.
	//appReq.SetDefaultValue()

	var task domain.AppServeAppTask
	//tasks := app.AppServeAppTasks
	//sort.Slice(tasks, func(i, j int) bool {
	//	return tasks[i].CreatedAt.String() > tasks[j].CreatedAt.String()
	//})
	//for _, t := range tasks {
	//	if t.Status == "DEPLOY_SUCCESS" {
	//		latestTask = t
	//		break
	//	}
	//}
	//if err = serializer.Map(latestTask, &task); err != nil {
	//	ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
	//	return
	//}

	var latestTask = app.AppServeAppTasks[0]
	if err = serializer.Map(latestTask, &task); err != nil {
		//ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
		return
	}

	if err = serializer.Map(appReq, &task); err != nil {
		//ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
		return
	}

	//updateVersion, err := strconv.Atoi(latestTask.Version)
	//if err != nil {
	//	ErrorJSON(w, r, httpErrors.NewInternalServerError(err,""))
	//}
	//task.Version = strconv.Itoa(updateVersion + 1)
	task.Version = strconv.Itoa(len(app.AppServeAppTasks) + 1)
	//task.AppServeAppId = app.ID
	task.Status = "PREPARING"
	task.RollbackVersion = ""
	task.Output = ""
	task.CreatedAt = time.Now()
	task.UpdatedAt = nil

	fmt.Println("===========================")
	fmt.Printf("%v\n", task)
	fmt.Println("===========================")

	var res string
	if appReq.Promote {
		res, err = h.usecase.PromoteAppServeApp(appId)
	} else if appReq.Abort {
		res, err = h.usecase.AbortAppServeApp(appId)
	} else {
		res, err = h.usecase.UpdateAppServeApp(app, &task)
	}

	if err != nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
		return
	}

	ResponseJSON(w, r, http.StatusOK, res)
}

// UpdateAppServeAppStatus godoc
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
//	@Router			/api/1.0/organizations/{organizationId}/projects/{projectId}/app-serve-apps/{appId}/status [patch]
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

	res, err := h.usecase.UpdateAppServeAppStatus(appId, appStatusReq.TaskID, appStatusReq.Status, appStatusReq.Output)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
		return
	}

	ResponseJSON(w, r, http.StatusOK, res)
}

// UpdateAppServeAppEndpoint godoc
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
//	@Router			/api/1.0/organizations/{organizationId}/projects/{projectId}/app-serve-apps/{appId}/endpoint [patch]
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
//	@Tags			AppServeApps
//	@Summary		Uninstall appServeApp
//	@Description	Uninstall appServeApp
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"Organization ID"
//	@Param			projectId		path		string	true	"Project ID"
//	@Param			appId			path		string	true	"App ID"
//	@Success		200				{object}	string
//	@Router			/api/1.0/organizations/{organizationId}/projects/{projectId}/app-serve-apps/{appId} [delete]
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

	res, err := h.usecase.DeleteAppServeApp(appId)
	if err != nil {
		log.ErrorWithContext(r.Context(), "Failed to delete appId err : ", err)
		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, res)
}

// RollbackAppServeApp godoc
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
//	@Router			/api/1.0/organizations/{organizationId}/projects/{projectId}/app-serve-apps/{appId}/rollback [post]
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

	res, err := h.usecase.RollbackAppServeApp(appId, appReq.TaskId)

	if err != nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
		return
	}

	ResponseJSON(w, r, http.StatusOK, res)
}
