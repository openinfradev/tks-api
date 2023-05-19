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
		"BUILDING":                  "BUILDING",
		"BUILD_SUCCESS":             "DONE",
		"BUILD_FAILED":              "FAILED",
		"DEPLOYING":                 "DEPLOYING",
		"DEPLOY_SUCCESS":            "DONE",
		"DEPLOY_FAILED":             "FAILED",
		"BLUEGREEN_DEPLOYING":       "DEPLOYING",
		"BLUEGREEN_WAIT":            "WAIT",
		"BLUEGREEN_DEPLOY_FAILED":   "FAILED",
		"BLUEGREEN_PROMOTING":       "PROMOTING",
		"BLUEGREEN_PROMOTE_SUCCESS": "DONE",
		"BLUEGREEN_PROMOTE_FAILED":  "FAILED",
		"BLUEGREEN_ABORTING":        "ABORTING",
		"BLUEGREEN_ABORT_SUCCESS":   "DONE",
		"BLUEGREEN_ABORT_FAILED":    "FAILED",
		"CANARY_DEPLOYING":          "DEPLOYING",
		"CANARY_WAIT":               "WAIT",
		"CANARY_DEPLOY_FAILED":      "FAILED",
		"CANARY_PROMOTING":          "PROMOTING",
		"CANARY_PROMOTE_SUCCESS":    "DONE",
		"CANARY_PROMOTE_FAILED":     "FAILED",
		"CANARY_ABORTING":           "ABORTING",
		"CANARY_ABORT_SUCCESS":      "DONE",
		"CANARY_ABORT_FAILED":       "FAILED",
		"ROLLBACKING":               "ROLLBACKING",
		"ROLLBACK_SUCCESS":          "DONE",
		"ROLLBACK_FAILED":           "FAILED",
		"DELETING":                  "DELETING",
		"DELETE_FAILED":             "FAILED",
	}
	StatusName = map[string]string{
		"BUILDING":                  "BUILD",
		"BUILD_SUCCESS":             "BUILD",
		"BUILD_FAILED":              "BUILD",
		"DEPLOYING":                 "DEPLOY",
		"DEPLOY_SUCCESS":            "DEPLOY",
		"DEPLOY_FAILED":             "DEPLOY",
		"BLUEGREEN_DEPLOYING":       "PROMOTE",
		"BLUEGREEN_WAIT":            "PROMOTE",
		"BLUEGREEN_DEPLOY_FAILED":   "PROMOTE",
		"BLUEGREEN_PROMOTING":       "PROMOTE",
		"BLUEGREEN_PROMOTE_SUCCESS": "PROMOTE",
		"BLUEGREEN_PROMOTE_FAILED":  "PROMOTE",
		"BLUEGREEN_ABORTING":        "PROMOTE",
		"BLUEGREEN_ABORT_SUCCESS":   "PROMOTE",
		"BLUEGREEN_ABORT_FAILED":    "PROMOTE",
		"CANARY_DEPLOYING":          "PROMOTE",
		"CANARY_WAIT":               "PROMOTE",
		"CANARY_DEPLOY_FAILED":      "PROMOTE",
		"CANARY_PROMOTING":          "PROMOTE",
		"CANARY_PROMOTE_SUCCESS":    "PROMOTE",
		"CANARY_PROMOTE_FAILED":     "PROMOTE",
		"CANARY_ABORTING":           "PROMOTE",
		"CANARY_ABORT_SUCCESS":      "PROMOTE",
		"CANARY_ABORT_FAILED":       "PROMOTE",
		"ROLLBACKING":               "ROLLBACK",
		"ROLLBACK_SUCCESS":          "ROLLBACK",
		"ROLLBACK_FAILED":           "ROLLBACK",
		"DELETING":                  "DELETE",
		"DELETE_FAILED":             "DELETE",
	}
	StatusStages = map[string][]string{
		"PREPARING":                 {},
		"BUILDING":                  {"BUILDING"},
		"BUILD_SUCCESS":             {"BUILD_SUCCESS"},
		"BUILD_FAILED":              {"BUILD_FAILED"},
		"DEPLOYING":                 {"BUILD_SUCCESS", "DEPLOYING"},
		"DEPLOY_SUCCESS":            {"BUILD_SUCCESS", "DEPLOY_SUCCESS"},
		"DEPLOY_FAILED":             {"BUILD_SUCCESS", "DEPLOY_FAILED"},
		"BLUEGREEN_DEPLOYING":       {"BUILD_SUCCESS", "DEPLOYING"},
		"BLUEGREEN_DEPLOY_FAILED":   {"BUILD_SUCCESS", "DEPLOY_FAILED"},
		"BLUEGREEN_WAIT":            {"BUILD_SUCCESS", "DEPLOY_SUCCESS", "BLUEGREEN_WAIT"},
		"BLUEGREEN_PROMOTING":       {"BUILD_SUCCESS", "DEPLOY_SUCCESS", "BLUEGREEN_PROMOTING"},
		"BLUEGREEN_PROMOTE_SUCCESS": {"BUILD_SUCCESS", "DEPLOY_SUCCESS", "BLUEGREEN_PROMOTE_SUCCESS"},
		"BLUEGREEN_PROMOTE_FAILED":  {"BUILD_SUCCESS", "DEPLOY_SUCCESS", "BLUEGREEN_PROMOTE_FAILED"},
		"BLUEGREEN_ABORTING":        {"BUILD_SUCCESS", "DEPLOY_SUCCESS", "BLUEGREEN_ABORTING"},
		"BLUEGREEN_ABORT_SUCCESS":   {"BUILD_SUCCESS", "DEPLOY_SUCCESS", "BLUEGREEN_ABORT_SUCCESS"},
		"BLUEGREEN_ABORT_FAILED":    {"BUILD_SUCCESS", "DEPLOY_SUCCESS", "BLUEGREEN_ABORT_FAILED"},
		"CANARY_DEPLOYING":          {"BUILD_SUCCESS", "DEPLOYING"},
		"CANARY_DEPLOY_FAILED":      {"BUILD_SUCCESS", "DEPLOY_FAILED"},
		"CANARY_WAIT":               {"BUILD_SUCCESS", "DEPLOY_SUCCESS", "CANARY_WAIT"},
		"CANARY_PROMOTING":          {"BUILD_SUCCESS", "DEPLOY_SUCCESS", "CANARY_PROMOTING"},
		"CANARY_PROMOTE_SUCCESS":    {"BUILD_SUCCESS", "DEPLOY_SUCCESS", "CANARY_PROMOTE_SUCCESS"},
		"CANARY_PROMOTE_FAILED":     {"BUILD_SUCCESS", "DEPLOY_SUCCESS", "CANARY_PROMOTE_FAILED"},
		"CANARY_ABORTING":           {"BUILD_SUCCESS", "DEPLOY_SUCCESS", "CANARY_ABORTING"},
		"CANARY_ABORT_SUCCESS":      {"BUILD_SUCCESS", "DEPLOY_SUCCESS", "CANARY_ABORT_SUCCESS"},
		"CANARY_ABORT_FAILED":       {"BUILD_SUCCESS", "DEPLOY_SUCCESS", "CANARY_ABORT_FAILED"},
		"ROLLBACKING":               {"ROLLBACKING"},
		"ROLLBACK_SUCCESS":          {"ROLLBACK_SUCCESS"},
		"ROLLBACK_FAILED":           {"ROLLBACK_FAILED"},
		"DELETING":                  {"DELETING"},
		"DELETE_FAILED":             {"DELETE_FAILED"},
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
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "", ""))
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
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "", ""))
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
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "", ""))
		return
	}

	appId, ok := vars["appId"]
	fmt.Printf("appId = [%s]\n", appId)
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid appId"), "", ""))
		return
	}
	app, err := h.usecase.GetAppServeAppById(appId)
	if err != nil {
		ErrorJSON(w, httpErrors.NewInternalServerError(err, "", ""))
		return
	}
	if app == nil {
		ErrorJSON(w, httpErrors.NewNoContentError(fmt.Errorf("no appId"), "", ""))
		return
	}

	// For very first task, rollback should be disabled.
  if len(app.AppServeAppTasks) > 1 {
		newTasks := make([]domain.AppServeAppTask, 0)
		for _, t := range app.AppServeAppTasks {
			if strings.Contains(t.Status, "SUCCESS") && t.Status != "BLUEGREEN_ABORT_SUCCESS" &&
				t.Status != "ROLLBACK_SUCCESS" {
				t.AvailableRollback = true
			}
			newTasks = append(newTasks, t)
		}
		app.AppServeAppTasks = newTasks
	}

	var out domain.GetAppServeAppResponse
	out.AppServeApp = *app
	out.Stages = makeStages(app)

	ResponseJSON(w, http.StatusOK, out)
}

// Name             - Status (Result)
// -------------------------------------------------------------------------------------
// PREPARE (준비)              - PREPARING (DONE)
// BUILD (빌드)                - BUILDING (BUILDING),              BUILD_SUCCESS (DONE),             BUILD_FAILED (FAILED)
// DEPLOY (배포)               - DEPLOYING (DEPLOYING),            DEPLOY_SUCCESS (DONE),            DEPLOY_FAILED (FAILED)
// BLUEGREEN_PROMOTE (프로모트) - BLUEGREEN_DEPLOYING (DEPLOYING),  BLUEGREEN_WAIT (WAIT),            BLUEGREEN_DEPLOY_FAILED (FAILED)
// BLUEGREEN_PROMOTE (프로모트) - BLUEGREEN_PROMOTING (PROMOTING),  BLUEGREEN_PROMOTE_SUCCESS (DONE), BLUEGREEN_PROMOTE_FAILED (FAILED)
// BLUEGREEN_PROMOTE (프로모트) - BLUEGREEN_ABORTING (ABORTING),    BLUEGREEN_ABORT_SUCCESS (DONE),   BLUEGREEN_ABORT_FAILED (FAILED)
// CANARY_PROMOTE (카나리아)    - CANARY_DEPLOYING (DEPLOYING),     CANARY_WAIT (WAIT),               CANARY_DEPLOY_FAILED (FAILED)
// CANARY_PROMOTE (카나리아)    - CANARY_PROMOTING (PROMOTING),     CANARY_PROMOTE_SUCCESS (DONE),    CANARY_PROMOTE_FAILED (FAILED)
// CANARY_PROMOTE (카나리아)    - CANARY_ABORTING (ABORTING),       CANARY_ABORT_SUCCESS (DONE),      CANARY_ABORT_FAILED (FAILED)
// ROLLBACK (롤백)             - ROLLBACKING (ROLLBACKING),        ROLLBACK_SUCCESS (DONE),          ROLLBACK_FAILED (FAILED)
func makeStages(app *domain.AppServeApp) []domain.StageResponse {
	stages := make([]domain.StageResponse, 0)

	var stage domain.StageResponse
	for _, s := range StatusStages[app.Status] {
		stage = makeStage(app, s)
		stages = append(stages, stage)
	}

	return stages
}

func makeStage(app *domain.AppServeApp, status string) domain.StageResponse {
	stage := domain.StageResponse{
		Name:   StatusName[status],
		Status: status,
		Result: StatusResult[status],
	}

	var actions []domain.ActionResponse
	if status == "DEPLOY_SUCCESS" {
		action := domain.ActionResponse{
			Name: "ENDPOINT",
			Uri:  app.EndpointUrl,
			Type: "LINK",
		}
		actions = append(actions, action)
	} else if status == "BLUEGREEN_WAIT" {
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
	} else if status == "BLUEGREEN_PROMOTE_SUCCESS" {
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
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "", ""))
		return
	}

	urlParams := r.URL.Query()
	appId := urlParams.Get("appId")
	if appId == "" {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid appId"), "", ""))
		return
	}

	exist, err := h.usecase.IsAppServeAppExist(appId)
	if err != nil {
		ErrorJSON(w, httpErrors.NewInternalServerError(err, "", ""))
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
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "", ""))
		return
	}
	appName, ok := vars["name"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid appName"), "", ""))
		return
	}

	existed, err := h.usecase.IsAppServeAppNameExist(organizationId, appName)
	if err != nil {
		ErrorJSON(w, httpErrors.NewInternalServerError(err, "", ""))
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
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "", ""))
		return
	}

	appId, ok := vars["appId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid appId"), "", ""))
		return
	}

	// priority
	// 1. Request,  2. default value  3. previous app and task

	// priority: 3. previous app
	app, err := h.usecase.GetAppServeAppById(appId)
	if err != nil {
		ErrorJSON(w, httpErrors.NewInternalServerError(err, "", ""))
		return
	}
	if len(app.AppServeAppTasks) < 1 {
		ErrorJSON(w, httpErrors.NewInternalServerError(err, "", ""))
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
		ErrorJSON(w, httpErrors.NewBadRequestError(err, "", ""))
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
	var latestTask = app.AppServeAppTasks[len(app.AppServeAppTasks)-1]
	if err = domain.Map(latestTask, &task); err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err, "", ""))
		return
	}

	// priority: 1. Request
	if err = domain.Map(appReq, &task); err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err, "", ""))
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
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "", ""))
		return
	}

	appId, ok := vars["appId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid appId"), "", ""))
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
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "", ""))
		return
	}

	appId, ok := vars["appId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid appId"), "", ""))
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
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "", ""))
		return
	}

	appId, ok := vars["appId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid appId"), "", ""))
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
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "", ""))
		return
	}

	appId, ok := vars["appId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid appId"), "", ""))
		return
	}

	appReq := domain.RollbackAppServeAppRequest{}
	err := UnmarshalRequestInput(r, &appReq)
	if err != nil {
		ErrorJSON(w, err)
		return
	}
	if appReq.TaskId == "" {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("no taskId"), "", ""))
		return
	}

	res, err := h.usecase.RollbackAppServeApp(appId, appReq.TaskId)

	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err, "", ""))
		return
	}

	ResponseJSON(w, http.StatusOK, res)
}
