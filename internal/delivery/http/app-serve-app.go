package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

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

	cont_id := urlParams.Get("contract_id")
	if cont_id == "" {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid contract_id")))
		return
	}

	show_all_str := urlParams.Get("show_all")
	if show_all_str == "" {
		show_all_str = "false"
	}

	show_all, err := strconv.ParseBool(show_all_str)
	if err != nil {
		log.Error("Failed to convert show_all params. Err: ", err)
		ErrorJSON(w, err)
		return
	}

	appServeApps, err := h.usecase.Fetch(cont_id, show_all)
	if err != nil {
		log.Error("Failed to get Failed to get app-serve-apps ", err)
		ErrorJSON(w, err)
		return
	}

	var out struct {
		AppServeApps []*domain.AppServeApp `json:"appServeApps"`
	}
	out.AppServeApps = appServeApps

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
	appServeAppId, ok := vars["appServeAppId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid appServeAppId")))
		return
	}

	res, err := h.usecase.Get(appServeAppId)
	if err != nil {
		log.Error("Failed to get Failed to get app-serve-app ", err)
		ErrorJSON(w, err)
		return
	}

	var out struct {
		AppServeAppCombined domain.AppServeAppCombined `json:"appServeApp"`
	}
	out.AppServeAppCombined = *res

	ResponseJSON(w, http.StatusOK, out)

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
	var appObj = domain.CreateAppServeAppRequest{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}
	err = json.Unmarshal(body, &appObj)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	log.Debug(fmt.Sprintf("*****\nIn handlers, appObj:\n%+v\n*****\n", appObj))

	// Validate common params
	if appObj.Name == "" || appObj.Type == "" || appObj.Version == "" ||
		appObj.AppType == "" || appObj.ContractId == "" {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf(`Error: The following params are always mandatory.
		- name
		- type
		- app_type
		- contract_id
		- version`)))
		return
	}

	// Validate port param for springboot app
	if appObj.AppType == "springboot" {
		if appObj.Port == "" {
			ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf(`Error: 'port' param is mandatory.`)))
			return
		}
	}

	// Validate 'type' param
	if !(appObj.Type == "build" || appObj.Type == "deploy" || appObj.Type == "all") {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf(`Error: 'type' should be one of these values.
- build
- deploy
- all`)))
		return
	}

	// Validate 'strategy' param
	if appObj.Strategy != "rolling-update" {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf(`Error: 'strategy' should be 'rolling-update' on first deployment.`)))
		return
	}

	// Validate 'app_type' param
	if !(appObj.AppType == "spring" || appObj.AppType == "springboot") {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf(`Error: 'type' should be one of these values.
- string
- stringboot`)))
		return
	}

	appServeAppId, err := h.usecase.Create(&appObj)
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	var out struct {
		AppServeAppId string `json:"appServeAppId"`
	}
	out.AppServeAppId = appServeAppId

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
	appServeAppId, ok := vars["appServeAppId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("invalid appServeAppId")))
		return
	}

	var app = domain.UpdateAppServeAppRequest{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}
	err = json.Unmarshal(body, &app)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	res := ""
	if app.Promote {
		res, err = h.usecase.Promote(appServeAppId, &app)
	} else if app.Abort {
		res, err = h.usecase.Abort(appServeAppId, &app)
	} else {
		res, err = h.usecase.Update(appServeAppId, &app)
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
	appServeAppId, ok := vars["appServeAppId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid appServeAppId")))
		return
	}

	res, err := h.usecase.Delete(appServeAppId)
	if err != nil {
		log.Error("Failed to delete appServeAppId err : ", err)
		ErrorJSON(w, err)
		return
	}

	ResponseJSON(w, http.StatusOK, res)
}
