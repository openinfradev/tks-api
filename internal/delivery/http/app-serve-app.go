package http

import (
	"encoding/json"
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
// @Param object body string true "body"
// @Success 200 {object} string
// @Router /app-serve-apps [post]
// @Security     JWT
func (h *AppServeAppHandler) CreateAppServeApp(w http.ResponseWriter, r *http.Request) {
	var appReq domain.CreateAppServeAppRequest
	var app domain.AppServeApp
	if err := json.NewDecoder(r.Body).Decode(&appReq); err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	now := time.Now()
	app = domain.AppServeApp{
		Name:               appReq.Name,
		OrganizationId:     appReq.OrganizationId,
		Type:               appReq.Type,
		AppType:            appReq.AppType,
		TargetClusterId:    appReq.TargetClusterId,
		EndpointUrl:        "N/A",
		PreviewEndpointUrl: "N/A",
		Status:             "PREPARING",
		CreatedAt:          now,
	}
	task := domain.AppServeAppTask{
		Version:        appReq.Version,
		Strategy:       appReq.Strategy,
		ArtifactUrl:    appReq.ArtifactUrl,
		ImageUrl:       appReq.ImageUrl,
		ExecutablePath: appReq.ExecutablePath,
		ResourceSpec:   appReq.ResourceSpec,
		Status:         "PREPARING",
		Profile:        appReq.Profile,
		AppConfig:      appReq.AppConfig,
		AppSecret:      appReq.AppSecret,
		ExtraEnv:       appReq.ExtraEnv,
		Port:           appReq.Port,
		Output:         "",
		PvEnabled:      appReq.PvEnabled,
		PvStorageClass: appReq.PvStorageClass,
		PvAccessMode:   appReq.PvAccessMode,
		PvSize:         appReq.PvSize,
		PvMountPath:    appReq.PvMountPath,
		CreatedAt:      now,
	}
	app.AppServeAppTasks = append(app.AppServeAppTasks, task)

	// Validate common params
	if app.Name == "" || app.Type == "" || app.AppServeAppTasks[0].Version == "" ||
		app.AppType == "" || app.OrganizationId == "" {
		ErrorJSON(w, httpErrors.NewBadRequestError(
			fmt.Errorf("Error: The following params are always mandatory."+
				"\n\t- name\n\t- type\n\t- app_type\n\t- organization_id\n\t- version")))
		return
	}

	// Validate port param for springboot app
	if app.AppType == "springboot" {
		if app.AppServeAppTasks[0].Port == "" {
			ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("error: 'port' param is mandatory")))
			return
		}
	}

	// Validate 'type' param
	if !(app.Type == "build" || app.Type == "deploy" || app.Type == "all") {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Error: 'type' should be one of these values."+
			"\n\t- build\n\t- deploy\n\t- all")))
		return
	}

	// Validate 'strategy' param
	if app.AppServeAppTasks[0].Strategy != "rolling-update" {
		ErrorJSON(w, httpErrors.NewBadRequestError(
			fmt.Errorf("error: 'strategy' should be 'rolling-update' on first deployment")))
		return
	}

	// Validate 'app_type' param
	if !(app.AppType == "spring" || app.AppType == "springboot") {
		ErrorJSON(w, httpErrors.NewBadRequestError(
			fmt.Errorf("Error: 'type' should be one of these values."+
				"\n\t- string\n\t- stringboot")))
		return
	}

	appId, appName, err := h.usecase.CreateAppServeApp(&app)
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	var out struct {
		AppId   string `json:"app_serve_app_id"`
		AppName string `json:"app_serve_app_name"`
	}
	out.AppId = appId
	out.AppName = appName

	ResponseJSON(w, http.StatusOK, out)
}

// GetAppServeApps godoc
// @Tags AppServeApps
// @Summary Get appServeApp list
// @Description Get appServeApp list by giving params
// @Accept json
// @Produce json
// @Param projectId query string false "project_id"
// @Param showAll query string false "show_all"
// @Success 200 {object} []domain.AppServeApp
// @Router /app-serve-apps [get]
// @Security     JWT
func (h *AppServeAppHandler) GetAppServeApps(w http.ResponseWriter, r *http.Request) {
	urlParams := r.URL.Query()

	organizationId := urlParams.Get("organizationId")
	if organizationId == "" {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId")))
		return
	}

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

	var out struct {
		AppServeApps []domain.AppServeApp `json:"app_serve_apps"`
	}
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
// @Router /app-serve-apps/{appServeAppId} [get]
// @Security     JWT
func (h *AppServeAppHandler) GetAppServeApp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appId, ok := vars["appId"]
	fmt.Printf("appId = [%s]", appId)
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid appId")))
		return
	}
	app, _ := h.usecase.GetAppServeAppById(appId)

	var out struct {
		AppServeApp domain.AppServeApp `json:"app_serve_app"`
	}
	out.AppServeApp = *app

	ResponseJSON(w, http.StatusOK, out)
}

// UpdateAppServeApp godoc
// @Tags AppServeApps
// @Summary Update appServeApp
// @Description Update appServeApp
// @Accept json
// @Produce json
// @Param object body string true "body"
// @Success 200 {object} object
// @Router /app-serve-apps [put]
// @Security     JWT
func (h *AppServeAppHandler) UpdateAppServeApp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appId, ok := vars["appId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid appId")))
		return
	}

	var appReq domain.UpdateAppServeAppRequest
	if err := json.NewDecoder(r.Body).Decode(&appReq); err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	appTask := &domain.AppServeAppTask{
		AppServeAppId:  appId,
		Version:        appReq.Version,
		Strategy:       appReq.Strategy,
		ArtifactUrl:    appReq.ArtifactUrl,
		ImageUrl:       appReq.ImageUrl,
		ExecutablePath: appReq.ExecutablePath,
		ResourceSpec:   appReq.ResourceSpec,
		Status:         "PREPARING",
		Profile:        appReq.Profile,
		AppConfig:      appReq.AppConfig,
		AppSecret:      appReq.AppSecret,
		ExtraEnv:       appReq.ExtraEnv,
		Port:           appReq.Port,
		Output:         "",
		CreatedAt:      time.Now(),
	}

	var res string
	var err error
	if appReq.Promote {
		res, err = h.usecase.PromoteAppServeApp(appId)
	} else if appReq.Abort {
		res, err = h.usecase.AbortAppServeApp(appId)
	} else {
		// Validate 'strategy' param
		if !(appReq.Strategy == "rolling-update" || appReq.Strategy == "blue-green" || appReq.Strategy == "canary") {
			errMsg := fmt.Sprintf("Error: 'strategy' should be one of these values." +
				"\n\t- rolling-update\n\t- blue-green\n\t- canary")
			log.Error(errMsg)
			ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf(errMsg)))
			return
		}

		res, err = h.usecase.UpdateAppServeApp(appTask)
	}

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
// @Router /app-serve-apps [delete]
// @Security     JWT
func (h *AppServeAppHandler) DeleteAppServeApp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
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
