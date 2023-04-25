package http

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
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
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId")))
		return
	}

	appReq := domain.CreateAppServeAppRequest{}
	err := UnmarshalRequestInput(r, &appReq)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	(appReq).SetDefaultValue()

	var app domain.AppServeApp
	if err = domain.Map(appReq, &app); err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
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
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
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
			ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("error: 'port' param is mandatory")))
			return
		}
	}

	// Validate 'strategy' param
	if app.AppServeAppTasks[0].Strategy != "rolling-update" {
		ErrorJSON(w, httpErrors.NewBadRequestError(
			fmt.Errorf("error: 'strategy' should be 'rolling-update' on first deployment")))
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
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId")))
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
// @Success 200 {object} domain.AppServeApp
// @Router /organizations/{organizationId}/app-serve-apps/{appServeAppId} [get]
// @Security     JWT
func (h *AppServeAppHandler) GetAppServeApp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	organizationId, ok := vars["organizationId"]
	fmt.Printf("organizationId = [%v]\n", organizationId)
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId")))
		return
	}

	appId, ok := vars["appId"]
	fmt.Printf("appId = [%s]\n", appId)
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid appId")))
		return
	}
	app, _ := h.usecase.GetAppServeAppById(appId)

	var out domain.GetAppServeAppResponse
	out.AppServeApp = *app

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
// @Router /organizations/{organizationId}/app-serve-apps [put]
// @Security     JWT
func (h *AppServeAppHandler) UpdateAppServeApp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	fmt.Printf("organizationId = [%v]\n", organizationId)
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId")))
		return
	}

	appId, ok := vars["appId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid appId")))
		return
	}

	// priority
	// 1. Request,  2. default value  3. previous app and task

	// priority: 3. previous app
	app, err := h.usecase.GetAppServeAppById(appId)
	if err != nil {
		ErrorJSON(w, httpErrors.NewInternalServerError(err))
		return
	}
	if len(app.AppServeAppTasks) < 1 {
		ErrorJSON(w, httpErrors.NewInternalServerError(err))
	}

	// priority: 1. Request
	appReq := domain.UpdateAppServeAppRequest{}
	err = UnmarshalRequestInput(r, &appReq)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	// priority: 2. Default Value
	appReq.SetDefaultValue()

	if err = domain.Map(appReq, app); err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	var task domain.AppServeAppTask
	var latestTask domain.AppServeAppTask
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
	//	ErrorJSON(w, httpErrors.NewBadRequestError(err))
	//	return
	//}

	// priority: 3. previous task
	latestTask = app.AppServeAppTasks[len(app.AppServeAppTasks)-1]
	if err = domain.Map(latestTask, &task); err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	// priority: 1. Request
	if err = domain.Map(appReq, &task); err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	//updateVersion, err := strconv.Atoi(latestTask.Version)
	//if err != nil {
	//	ErrorJSON(w, httpErrors.NewInternalServerError(err))
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
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
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
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId")))
		return
	}

	appId, ok := vars["appId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid appId")))
		return
	}

	appStatusReq := domain.UpdateAppServeAppStatusRequest{}
	err := UnmarshalRequestInput(r, &appStatusReq)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	res, err := h.usecase.UpdateAppServeAppStatus(appId, appStatusReq.TaskID, appStatusReq.Status, appStatusReq.Output)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
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
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId")))
		return
	}

	appId, ok := vars["appId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid appId")))
		return
	}

	appReq := domain.UpdateAppServeAppEndpointRequest{}
	err := UnmarshalRequestInput(r, &appReq)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	res, err := h.usecase.UpdateAppServeAppEndpoint(
		appId,
		appReq.TaskID,
		appReq.EndpointUrl,
		appReq.PreviewEndpointUrl,
		appReq.HelmRevision)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
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
// @Router /organizations/{organizationId}/app-serve-apps [delete]
// @Security     JWT
func (h *AppServeAppHandler) DeleteAppServeApp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	fmt.Printf("organizationId = [%v]\n", organizationId)
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId")))
		return
	}

	appId, ok := vars["appId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid appId")))
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
