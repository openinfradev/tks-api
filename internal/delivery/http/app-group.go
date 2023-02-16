package http

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/internal/domain"
	"github.com/openinfradev/tks-api/internal/usecase"
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

// CreateappGroup godoc
// @Tags appGroups
// @Summary Install appGroup
// @Description Install appGroup
// @Accept json
// @Produce json
// @Param object body string true "body"
// @Success 200 {object} appGroupId
// @Router /app-groups [post]
func (h *AppGroupHandler) CreateAppGroup(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		ClusterId   string `json:"clusterId"`
		Type        string `json:"type"`
	}
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
		log.Error("Failed to create appGroup err : ", err)
		InternalServerError(w)
		return
	}

	var out struct {
		AppGroupId string `json:"appGroupId"`
	}
	out.AppGroupId = appGroupId

	ResponseJSON(w, out, http.StatusOK)
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

	ResponseJSON(w, out, http.StatusOK)

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
func (h *AppGroupHandler) GetAppGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appGroupId, ok := vars["appGroupId"]
	if !ok {
		ErrorJSON(w, "Invalid prameters", http.StatusBadRequest)
		return
	}

	appGroup, err := h.usecase.Get(appGroupId)
	if err != nil {
		InternalServerError(w)
		return
	}

	var out struct {
		AppGroup domain.AppGroup `json:"appGroup"`
	}
	out.AppGroup = appGroup

	ResponseJSON(w, out, http.StatusOK)
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
func (h *AppGroupHandler) DeleteAppGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appGroupId, ok := vars["appGroupId"]
	if !ok {
		ErrorJSON(w, "Invalid prameters", http.StatusBadRequest)
		return
	}

	err := h.usecase.Delete(appGroupId)
	if err != nil {
		log.Error("Failed to create appGroup err : ", err)
		InternalServerError(w)
		return
	}

	ResponseJSON(w, nil, http.StatusOK)
}
