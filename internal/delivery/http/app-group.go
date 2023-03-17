package http

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
)

type AppGroupHandler struct {
	usecase usecase.IAppGroupUsecase
}

func NewAppGroupHandler(h usecase.IAppGroupUsecase) *AppGroupHandler {
	return &AppGroupHandler{
		usecase: h,
	}
}

// CreateAppGroup godoc
// @Tags AppGroups
// @Summary Install appGroup
// @Description Install appGroup
// @Accept json
// @Produce json
// @Param body body domain.CreateAppGroupRequest true "create appgroup request"
// @Success 200 {object} string
// @Router /app-groups [post]
// @Security     JWT
func (h *AppGroupHandler) CreateAppGroup(w http.ResponseWriter, r *http.Request) {
	var input = domain.CreateAppGroupRequest{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		ErrorJSON(w, "Invalid json", http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(body, &input)
	if err != nil {
		ErrorJSON(w, "Invalid json", http.StatusBadRequest)
		return
	}

	if input.Type != "LMA" && input.Type != "SERVICE_MESH" && input.Type != "LMA_EFK" {
		ErrorJSON(w, "Invalid application type", http.StatusBadRequest)
		return
	}

	appGroupId, err := h.usecase.Create(input.ClusterId, input.Name, input.Type, "", input.Description)
	if err != nil {
		InternalServerError(w, err)
		return
	}

	var out struct {
		AppGroupId string `json:"appGroupId"`
	}
	out.AppGroupId = appGroupId

	ResponseJSON(w, out, "", http.StatusOK)
}

// GetAppGroups godoc
// @Tags AppGroups
// @Summary Get appGroup list
// @Description Get appGroup list by giving params
// @Accept json
// @Produce json
// @Param clusterId query string false "clusterId"
// @Success 200 {object} []domain.AppGroup
// @Router /app-groups [get]
// @Security     JWT
func (h *AppGroupHandler) GetAppGroups(w http.ResponseWriter, r *http.Request) {
	urlParams := r.URL.Query()

	clusterId := urlParams.Get("clusterId")
	if clusterId == "" {
		ErrorJSON(w, "Invalid prameters", http.StatusBadRequest)
		return
	}

	appGroups, err := h.usecase.Fetch(clusterId)
	if err != nil {
		ErrorJSON(w, "Failed to get appGroups", http.StatusBadRequest)
		return
	}

	var out struct {
		AppGroups []domain.AppGroup `json:"appGroups"`
	}
	out.AppGroups = appGroups

	ResponseJSON(w, out, "", http.StatusOK)

}

// GetAppGroup godoc
// @Tags AppGroups
// @Summary Get appGroup detail
// @Description Get appGroup detail by appGroupId
// @Accept json
// @Produce json
// @Param appGroupId path string true "appGroupId"
// @Success 200 {object} []domain.AppGroup
// @Router /app-groups/{appGroupId} [get]
// @Security     JWT
func (h *AppGroupHandler) GetAppGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appGroupId, ok := vars["appGroupId"]
	if !ok {
		ErrorJSON(w, "Invalid prameters", http.StatusBadRequest)
		return
	}

	appGroup, err := h.usecase.Get(appGroupId)
	if err != nil {
		InternalServerError(w, err)
		return
	}

	var out struct {
		AppGroup domain.AppGroup `json:"appGroup"`
	}
	out.AppGroup = appGroup

	ResponseJSON(w, out, "", http.StatusOK)
}

// DeleteAppGroup godoc
// @Tags AppGroups
// @Summary Uninstall appGroup
// @Description Uninstall appGroup
// @Accept json
// @Produce json
// @Param object body string true "body"
// @Success 200 {object} object
// @Router /app-groups [delete]
// @Security     JWT
func (h *AppGroupHandler) DeleteAppGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appGroupId, ok := vars["appGroupId"]
	if !ok {
		ErrorJSON(w, "Invalid prameters", http.StatusBadRequest)
		return
	}

	err := h.usecase.Delete(appGroupId)
	if err != nil {
		log.Error("Failed to delete appGroup err : ", err)
		InternalServerError(w, err)
		return
	}

	ResponseJSON(w, nil, "", http.StatusOK)
}

// GetApplications godoc
// @Tags AppGroups
// @Summary Get applications
// @Description Get applications
// @Accept json
// @Produce json
// @Param appGroupId path string true "appGroupId"
// @Success 200 {object} object
// @Router /app-groups/{appGroupId}/applications [get]
// @Security     JWT
func (h *AppGroupHandler) GetApplications(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appGroupId, ok := vars["appGroupId"]
	if !ok {
		ErrorJSON(w, "Invalid prameters", http.StatusBadRequest)
		return
	}

	applications, err := h.usecase.GetApplications(appGroupId)
	if err != nil {
		log.Error("Failed to get application err : ", err)
		InternalServerError(w, err)
		return
	}

	var out struct {
		Applications []domain.Application `json:"applications"`
	}
	out.Applications = applications

	ResponseJSON(w, out, "", http.StatusOK)
}

// UpdateApplication godoc
// @Tags AppGroups
// @Summary Update application
// @Description Update application
// @Accept json
// @Produce json
// @Param object body domain.UpdateApplicationRequest true "body"
// @Success 200 {object} object
// @Router /app-groups/{appGroupId}/applications [post]
// @Security     JWT
func (h *AppGroupHandler) UpdateApplication(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appGroupId, ok := vars["appGroupId"]
	if !ok {
		ErrorJSON(w, "Invalid prameters", http.StatusBadRequest)
		return
	}

	var input = domain.UpdateApplicationRequest{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		ErrorJSON(w, "Invalid json", http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(body, &input)
	if err != nil {
		ErrorJSON(w, "Invalid json", http.StatusBadRequest)
		return
	}

	err = h.usecase.UpdateApplication(appGroupId, input)
	if err != nil {
		log.Error("Failed to update application err : ", err)
		InternalServerError(w, err)
		return
	}

	ResponseJSON(w, nil, "", http.StatusOK)
}
