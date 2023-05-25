package http

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/openinfradev/tks-api/internal"
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

func NewAppServeAppHandler(h usecase.IAppServeAppUsecase) *AppServeAppHandler {
	return &AppServeAppHandler{
		usecase: h,
	}
}

// CreateAppServeApp godoc
// @Tags AppServeApps
// @Summary Install appServeApp
// @Description Install appServeApp
// @Accept json
// @Produce json
// @Param object body domain.CreateAppServeAppRequest true "create appserve request"
// @Success 200 {object} string
// @Router /organizations/{organizationId}/app-serve-apps [post]
// @Security     JWT
func (h *AppServeAppHandler) CreateAppServeApp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	fmt.Printf("organizationId = [%v]\n", organizationId)
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	appReq := domain.CreateAppServeAppRequest{}
	err := UnmarshalRequestInput(r, &appReq)
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	(appReq).SetDefaultValue()

	var app domain.AppServeApp
	if err = domain.Map(appReq, &app); err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err, "", ""))
		return
	}

	now := time.Now()
	app.OrganizationId = organizationId
	app.EndpointUrl = "N/A"
	app.PreviewEndpointUrl = "N/A"
	app.Status = "PREPARING"
	app.CreatedAt = now

	var task domain.AppServeAppTask
	if err = domain.Map(appReq, &task); err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err, "", ""))
		return
	}

	task.Version = "1"
	task.Status = "PREPARING"
	task.Output = ""
	task.CreatedAt = now

	app.AppServeAppTasks = append(app.AppServeAppTasks, task)

	// Validate port param for springboot app
	if app.AppType == "springboot" {
		if app.AppServeAppTasks[0].Port == "" {
			ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("error: 'port' param is mandatory"), "", ""))
			return
		}
	}

	// Validate 'strategy' param
	if app.AppServeAppTasks[0].Strategy != "rolling-update" {
		ErrorJSON(w, httpErrors.NewBadRequestError(
			fmt.Errorf("error: 'strategy' should be 'rolling-update' on first deployment"), "", ""))
		return
	}

	_, _, err = h.usecase.CreateAppServeApp(&app)
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	var out domain.CreateAppServeAppResponse
	if err = domain.Map(app, &out); err != nil {
		ErrorJSON(w, err)
		return
	}

	ResponseJSON(w, http.StatusOK, out)
}

// GetAppServeApps godoc
// @Tags AppServeApps
// @Summary Get appServeApp list
// @Description Get appServeApp list by giving params
// @Accept json
// @Produce json
// @Param organization_Id query string false "organization_Id"
// @Param showAll query string false "show_all"
// @Success 200 {object} []domain.AppServeApp
// @Router /organizations/{organizationId}/app-serve-apps [get]
// @Security     JWT
func (h *AppServeAppHandler) GetAppServeApps(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	fmt.Printf("organizationId = [%v]\n", organizationId)
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	urlParams := r.URL.Query()

	showAllParam := urlParams.Get("showAll")
	if showAllParam == "" {
		showAllParam = "false"
	}

	showAll, err := strconv.ParseBool(showAllParam)
	if err != nil {
		log.Error("Failed to convert showAll params. Err: ", err)
		ErrorJSON(w, err)
		return
	}

	apps, err := h.usecase.GetAppServeApps(organizationId, showAll)
	if err != nil {
		log.Error("Failed to get Failed to get app-serve-apps ", err)
		ErrorJSON(w, err)
		return
	}

	var out domain.GetAppServeAppsResponse
	out.AppServeApps = apps

	ResponseJSON(w, http.StatusOK, out)
}

// GetAppServeApp godoc
// @Tags AppServeApps
// @Summary Get appServeApp
// @Description Get appServeApp by giving params
// @Accept json
// @Produce json
// @Success 200 {object} domain.GetAppServeAppResponse
// @Router /organizations/{organizationId}/app-serve-apps/{appId} [get]
// @Security     JWT
func (h *AppServeAppHandler) GetAppServeApp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	organizationId, ok := vars["organizationId"]
	fmt.Printf("organizationId = [%v]\n", organizationId)
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	appId, ok := vars["appId"]
	fmt.Printf("appId = [%s]\n", appId)
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid appId"), "C_INVALID_ASA_ID", ""))
		return
	}
	app, err := h.usecase.GetAppServeAppById(appId)
	if err != nil {
		ErrorJSON(w, httpErrors.NewInternalServerError(err, "", ""))
		return
	}
	if app == nil {
		ErrorJSON(w, httpErrors.NewNoContentError(fmt.Errorf("no appId"), "D_NO_ASA", ""))
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
	out.Stages = makeStages(app)

	ResponseJSON(w, http.StatusOK, out)
}

func makeStages(app *domain.AppServeApp) []domain.StageResponse {
	stages := make([]domain.StageResponse, 0)

	var stage domain.StageResponse
	var pipelines []string
	taskStatus := app.AppServeAppTasks[0].Status
	strategy := app.AppServeAppTasks[0].Strategy

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
		stage = makeStage(app, pl)
		stages = append(stages, stage)
	}

	return stages
}

func makeStage(app *domain.AppServeApp, pl string) domain.StageResponse {
	stage := domain.StageResponse{
		Name:   pl,
		Status: StatusStages[app.Status][pl],
		Result: StatusResult[StatusStages[app.Status][pl]],
	}

	var actions []domain.ActionResponse
	if stage.Status == "DEPLOY_SUCCESS" {
		action := domain.ActionResponse{
			Name: "ENDPOINT",
			Uri:  app.EndpointUrl,
			Type: "LINK",
		}
		actions = append(actions, action)
	} else if stage.Status == "PROMOTE_WAIT" && app.AppServeAppTasks[0].Strategy == "blue-green" {
		action := domain.ActionResponse{
			Name: "PREVIEW",
			Uri:  app.PreviewEndpointUrl,
			Type: "LINK",
		}
		actions = append(actions, action)

		action = domain.ActionResponse{
			Name: "PROMOTE",
			Uri: fmt.Sprintf(internal.API_PREFIX+internal.API_VERSION+
				"/organizations/%v/app-serve-apps/%v", app.OrganizationId, app.ID),
			Type:   "API",
			Method: "PUT",
			Body:   map[string]string{"strategy": "blue-green", "promote": "true"},
		}
		actions = append(actions, action)

		action = domain.ActionResponse{
			Name: "ABORT",
			Uri: fmt.Sprintf(internal.API_PREFIX+internal.API_VERSION+
				"/organizations/%v/app-serve-apps/%v", app.OrganizationId, app.ID),
			Type:   "API",
			Method: "PUT",
			Body:   map[string]string{"strategy": "blue-green", "abort": "true"},
		}
		actions = append(actions, action)
	} else if stage.Status == "PROMOTE_SUCCESS" {
		action := domain.ActionResponse{
			Name: "ENDPOINT",
			Uri:  app.EndpointUrl,
			Type: "LINK",
		}
		actions = append(actions, action)
	}

	stage.Actions = &actions

	return stage
}

// IsAppServeAppExist godoc
// @Tags AppServeApps
// @Summary Get appServeApp
// @Description Get appServeApp by giving params
// @Accept json
// @Produce json
// @Success 200 {object} bool
// @Router /organizations/{organizationId}/app-serve-apps/app-id/exist [get]
// @Security     JWT
func (h *AppServeAppHandler) IsAppServeAppExist(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	organizationId, ok := vars["organizationId"]
	fmt.Printf("organizationId = [%v]\n", organizationId)
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	urlParams := r.URL.Query()
	appId := urlParams.Get("appId")
	if appId == "" {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid appId"), "C_INVALID_ASA_ID", ""))
		return
	}

	exist, err := h.usecase.IsAppServeAppExist(appId)
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	var out = struct {
		Exist bool `json:"exist"`
	}{Exist: exist}

	ResponseJSON(w, http.StatusOK, out)
}

// IsAppServeAppNameExist godoc
// @Tags AppServeApps
// @Summary Check duplicate appServeAppName
// @Description Check duplicate appServeAppName by giving params
// @Accept json
// @Produce json
// @Param organizationId path string true "organizationId"
// @Param name path string true "name"
// @Success 200 {object} bool
// @Router /organizations/{organizationId}/app-serve-apps/name/{name}/existence [get]
// @Security     JWT
func (h *AppServeAppHandler) IsAppServeAppNameExist(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	organizationId, ok := vars["organizationId"]
	fmt.Printf("organizationId = [%v]\n", organizationId)
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}
	appName, ok := vars["name"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid appName"), "C_INVALID_ASA_ID", ""))
		return
	}

	existed, err := h.usecase.IsAppServeAppNameExist(organizationId, appName)
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	var out = struct {
		Existed bool `json:"existed"`
	}{Existed: existed}

	ResponseJSON(w, http.StatusOK, out)
}

// UpdateAppServeApp godoc
// @Tags AppServeApps
// @Summary Update appServeApp
// @Description Update appServeApp
// @Accept json
// @Produce json
// @Param object body domain.UpdateAppServeAppRequest true "update appserve request"
// @Success 200 {object} object
// @Router /organizations/{organizationId}/app-serve-apps/{appId} [put]
// @Security     JWT
func (h *AppServeAppHandler) UpdateAppServeApp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	fmt.Printf("organizationId = [%v]\n", organizationId)
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	appId, ok := vars["appId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid appId"), "C_INVALID_ASA_ID", ""))
		return
	}

	// priority
	// 1. Request,  2. default value  3. previous app and task

	// priority: 3. previous app
	app, err := h.usecase.GetAppServeAppById(appId)
	if err != nil {
		ErrorJSON(w, err)
		return
	}
	if len(app.AppServeAppTasks) < 1 {
		ErrorJSON(w, err)
	}

	// priority: 1. Request
	appReq := domain.UpdateAppServeAppRequest{}
	err = UnmarshalRequestInput(r, &appReq)
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	// priority: 2. Default Value
	appReq.SetDefaultValue()

	if err = domain.Map(appReq, app); err != nil {
		//ErrorJSON(w, httpErrors.NewBadRequestError(err, "", ""))
		return
	}

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
	//if err = domain.Map(latestTask, &task); err != nil {
	//	ErrorJSON(w, httpErrors.NewBadRequestError(err, "", ""))
	//	return
	//}

	// priority: 3. previous task
	//var latestTask = app.AppServeAppTasks[len(app.AppServeAppTasks)-1]
	var latestTask = app.AppServeAppTasks[0]
	if err = domain.Map(latestTask, &task); err != nil {
		//ErrorJSON(w, httpErrors.NewBadRequestError(err, "", ""))
		return
	}

	// priority: 1. Request
	if err = domain.Map(appReq, &task); err != nil {
		//ErrorJSON(w, httpErrors.NewBadRequestError(err, "", ""))
		return
	}

	//updateVersion, err := strconv.Atoi(latestTask.Version)
	//if err != nil {
	//	ErrorJSON(w, httpErrors.NewInternalServerError(err,""))
	//}
	//task.Version = strconv.Itoa(updateVersion + 1)
	task.Version = strconv.Itoa(len(app.AppServeAppTasks) + 1)
	//task.AppServeAppId = app.ID
	task.Status = "PREPARING"
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
		ErrorJSON(w, httpErrors.NewBadRequestError(err, "", ""))
		return
	}

	ResponseJSON(w, http.StatusOK, res)
}

// UpdateAppServeAppStatus godoc
// @Tags AppServeApps
// @Summary Update app status
// @Description Update app status
// @Accept json
// @Produce json
// @Param appId path string true "appId"
// @Param body body domain.UpdateAppServeAppStatusRequest true "update app status request"
// @Success 200 {object} object
// @Router /organizations/{organizationId}/app-serve-apps/{appId}/status [patch]
// @Security     JWT
func (h *AppServeAppHandler) UpdateAppServeAppStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	fmt.Printf("organizationId = [%v]\n", organizationId)
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	appId, ok := vars["appId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid appId"), "C_INVALID_ASA_ID", ""))
		return
	}

	appStatusReq := domain.UpdateAppServeAppStatusRequest{}
	err := UnmarshalRequestInput(r, &appStatusReq)
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	res, err := h.usecase.UpdateAppServeAppStatus(appId, appStatusReq.TaskID, appStatusReq.Status, appStatusReq.Output)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err, "", ""))
		return
	}

	ResponseJSON(w, http.StatusOK, res)
}

// UpdateAppServeAppEndpoint godoc
// @Tags AppServeApps
// @Summary Update app endpoint
// @Description Update app endpoint
// @Accept json
// @Produce json
// @Param appId path string true "appId"
// @Param body body domain.UpdateAppServeAppEndpointRequest true "update app endpoint request"
// @Success 200 {object} object
// @Router /organizations/{organizationId}/app-serve-apps/{appId}/endpoint [patch]
// @Security     JWT
func (h *AppServeAppHandler) UpdateAppServeAppEndpoint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	fmt.Printf("organizationId = [%v]\n", organizationId)
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	appId, ok := vars["appId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid appId"), "C_INVALID_ASA_ID", ""))
		return
	}

	appReq := domain.UpdateAppServeAppEndpointRequest{}
	err := UnmarshalRequestInput(r, &appReq)
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	res, err := h.usecase.UpdateAppServeAppEndpoint(
		appId,
		appReq.TaskID,
		appReq.EndpointUrl,
		appReq.PreviewEndpointUrl,
		appReq.HelmRevision)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err, "", ""))
		return
	}

	ResponseJSON(w, http.StatusOK, res)
}

// DeleteAppServeApp godoc
// @Tags AppServeApps
// @Summary Uninstall appServeApp
// @Description Uninstall appServeApp
// @Accept json
// @Produce json
// @Param object body string true "body"
// @Success 200 {object} object
// @Router /organizations/{organizationId}/app-serve-apps/{appId} [delete]
// @Security     JWT
func (h *AppServeAppHandler) DeleteAppServeApp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	fmt.Printf("organizationId = [%v]\n", organizationId)
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	appId, ok := vars["appId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid appId"), "C_INVALID_ASA_ID", ""))
		return
	}

	res, err := h.usecase.DeleteAppServeApp(appId)
	if err != nil {
		log.Error("Failed to delete appId err : ", err)
		ErrorJSON(w, err)
		return
	}

	ResponseJSON(w, http.StatusOK, res)
}

// RollbackAppServeApp godoc
// @Tags AppServeApps
// @Summary Rollback appServeApp
// @Description Rollback appServeApp
// @Accept json
// @Produce json
// @Param object body domain.RollbackAppServeAppRequest true "rollback appserve request"
// @Success 200 {object} object
// @Router /organizations/{organizationId}/app-serve-apps/{appId}/rollback [post]
// @Security     JWT
func (h *AppServeAppHandler) RollbackAppServeApp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	fmt.Printf("organizationId = [%v]\n", organizationId)
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	appId, ok := vars["appId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid appId"), "C_INVALID_ASA_ID", ""))
		return
	}

	appReq := domain.RollbackAppServeAppRequest{}
	err := UnmarshalRequestInput(r, &appReq)
	if err != nil {
		ErrorJSON(w, err)
		return
	}
	if appReq.TaskId == "" {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("no taskId"), "C_INVALID_ASA_TASK_ID", ""))
		return
	}

	res, err := h.usecase.RollbackAppServeApp(appId, appReq.TaskId)

	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err, "", ""))
		return
	}

	ResponseJSON(w, http.StatusOK, res)
}
