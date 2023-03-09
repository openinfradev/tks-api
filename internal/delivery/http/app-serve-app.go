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
		ErrorJSON(w, "Invalid contract_id", http.StatusBadRequest)
		return
	}

	show_all_str := urlParams.Get("show_all")
	if show_all_str == "" {
		show_all_str = "false"
	}

	show_all, err := strconv.ParseBool(show_all_str)
	if err != nil {
		log.Error("Failed to convert show_all params. Err: ", err)
		InternalServerError(w, err)
		return
	}

	appServeApps, err := h.usecase.Fetch(cont_id, show_all)
	if err != nil {
		log.Error("Failed to get Failed to get app-serve-apps ", err)
		InternalServerError(w, err)
		return
	}

	var out struct {
		AppServeApps []*domain.AppServeApp `json:"appServeApps"`
	}
	out.AppServeApps = appServeApps

	ResponseJSON(w, out, "", http.StatusOK)

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
		ErrorJSON(w, "Invalid prameters", http.StatusBadRequest)
		return
	}

	res, err := h.usecase.Get(appServeAppId)
	if err != nil {
		log.Error("Failed to get Failed to get app-serve-app ", err)
		InternalServerError(w, err)
		return
	}

	var out struct {
		AppServeAppCombined domain.AppServeAppCombined `json:"appServeApp"`
	}
	out.AppServeAppCombined = *res

	ResponseJSON(w, out, "", http.StatusOK)

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
		ErrorJSON(w, fmt.Sprintf("Invalid json, err : %s", err), http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(body, &appObj)
	if err != nil {
		ErrorJSON(w, fmt.Sprintf("Invalid json, err : %s", err), http.StatusBadRequest)
		return
	}

	log.Debug(fmt.Sprintf("*****\nIn handlers, appObj:\n%+v\n*****\n", appObj))

	// Validate common params
	if appObj.Name == "" || appObj.Type == "" || appObj.Version == "" ||
		appObj.AppType == "" || appObj.ContractId == "" {
		errMsg := fmt.Sprintf(`Error: The following params are always mandatory.
- name
- type
- app_type
- contract_id
- version`)
		ErrorJSON(w, errMsg, http.StatusBadRequest)
		return
	}

	// Validate port param for springboot app
	if appObj.AppType == "springboot" {
		if appObj.Port == "" {
			errMsg := fmt.Sprintf("Error: 'port' param is mandatory.")
			ErrorJSON(w, errMsg, http.StatusBadRequest)
			return
		}
	}

	// Validate 'type' param
	if !(appObj.Type == "build" || appObj.Type == "deploy" || appObj.Type == "all") {
		errMsg := fmt.Sprintf(`Error: 'type' should be one of these values.
- build
- deploy
- all`)
		ErrorJSON(w, errMsg, http.StatusBadRequest)
		return
	}

	// Validate 'strategy' param
	if appObj.Strategy != "rolling-update" {
		errMsg := fmt.Sprintf("Error: 'strategy' should be 'rolling-update' on first deployment.")
		ErrorJSON(w, errMsg, http.StatusBadRequest)
		return
	}

	// Validate 'app_type' param
	if !(appObj.AppType == "spring" || appObj.AppType == "springboot") {
		errMsg := fmt.Sprintf(`Error: 'type' should be one of these values.
- string
- stringboot`)
		ErrorJSON(w, errMsg, http.StatusBadRequest)
		return
	}

	appServeAppId, err := h.usecase.Create(&appObj)
	if err != nil {
		InternalServerError(w, err)
		return
	}

	var out struct {
		AppServeAppId string `json:"appServeAppId"`
	}
	out.AppServeAppId = appServeAppId

	ResponseJSON(w, out, "", http.StatusOK)
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
		ErrorJSON(w, "Invalid prameters", http.StatusBadRequest)
		return
	}

	var app = domain.UpdateAppServeAppRequest{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		ErrorJSON(w, fmt.Sprintf("Invalid json, err : %s", err), http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(body, &app)
	if err != nil {
		ErrorJSON(w, fmt.Sprintf("Invalid json, err : %s", err), http.StatusBadRequest)
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
		ErrorJSON(w, err.Error(), http.StatusBadRequest)
		return
	}

	ResponseJSON(w, res, "", http.StatusOK)
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
		ErrorJSON(w, "Invalid prameters", http.StatusBadRequest)
		return
	}

	res, err := h.usecase.Delete(appServeAppId)
	if err != nil {
		log.Error("Failed to delete appServeAppId err : ", err)
		InternalServerError(w, err)
		return
	}

	ResponseJSON(w, res, "", http.StatusOK)
}
