package http

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/internal/helper"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
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
// @Success 200 {object} domain.CreateAppGroupResponse
// @Router /app-groups [post]
// @Security     JWT
func (h *AppGroupHandler) CreateAppGroup(w http.ResponseWriter, r *http.Request) {
	input := domain.CreateAppGroupRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	var dto domain.AppGroup
	if err = domain.Map(input, &dto); err != nil {
		log.Info(err)
	}

	appGroupId, err := h.usecase.Create(r.Context(), dto)
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	var out domain.CreateAppGroupResponse
	out.ID = appGroupId.String()

	ResponseJSON(w, http.StatusOK, out)
}

// GetAppGroups godoc
// @Tags AppGroups
// @Summary Get appGroup list
// @Description Get appGroup list by giving params
// @Accept json
// @Produce json
// @Param clusterId query string false "clusterId"
// @Success 200 {object} domain.GetAppGroupsResponse
// @Router /app-groups [get]
// @Security     JWT
func (h *AppGroupHandler) GetAppGroups(w http.ResponseWriter, r *http.Request) {
	urlParams := r.URL.Query()

	clusterId := urlParams.Get("clusterId")
	if clusterId == "" || !helper.ValidateClusterId(clusterId) {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid clusterId")))
		return
	}

	appGroups, err := h.usecase.Fetch(domain.ClusterId(clusterId))
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	var out domain.GetAppGroupsResponse
	out.AppGroups = make([]domain.AppGroupResponse, len(appGroups))
	for i, appGroup := range appGroups {
		if err := domain.Map(appGroup, &out.AppGroups[i]); err != nil {
			log.Info(err)
			continue
		}
	}

	ResponseJSON(w, http.StatusOK, out)
}

// GetAppGroup godoc
// @Tags AppGroups
// @Summary Get appGroup detail
// @Description Get appGroup detail by appGroupId
// @Accept json
// @Produce json
// @Param appGroupId path string true "appGroupId"
// @Success 200 {object} domain.GetAppGroupResponse
// @Router /app-groups/{appGroupId} [get]
// @Security     JWT
func (h *AppGroupHandler) GetAppGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strId, ok := vars["appGroupId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid appGroupId")))
		return
	}
	appGroupId := domain.AppGroupId(strId)
	if !appGroupId.Validate() {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid appGroupId")))
		return
	}
	appGroup, err := h.usecase.Get(appGroupId)
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	var out domain.GetAppGroupResponse
	if err := domain.Map(appGroup, &out.AppGroup); err != nil {
		log.Info(err)
	}

	ResponseJSON(w, http.StatusOK, out)
}

// DeleteAppGroup godoc
// @Tags AppGroups
// @Summary Uninstall appGroup
// @Description Uninstall appGroup
// @Accept json
// @Produce json
// @Param object body string true "body"
// @Success 200 {object} nil
// @Router /app-groups [delete]
// @Security     JWT
func (h *AppGroupHandler) DeleteAppGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strId, ok := vars["appGroupId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid appGroupId")))
		return
	}
	appGroupId := domain.AppGroupId(strId)
	if !appGroupId.Validate() {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid appGroupId")))
		return
	}

	err := h.usecase.Delete(appGroupId)
	if err != nil {
		log.Error("Failed to delete appGroup err : ", err)
		ErrorJSON(w, err)
		return
	}

	ResponseJSON(w, http.StatusOK, nil)
}

// GetApplications godoc
// @Tags AppGroups
// @Summary Get applications
// @Description Get applications
// @Accept json
// @Produce json
// @Param appGroupId path string true "appGroupId"
// @Param applicationType query string true "applicationType"
// @Success 200 {object} domain.GetApplicationsResponse
// @Router /app-groups/{appGroupId}/applications [get]
// @Security     JWT
func (h *AppGroupHandler) GetApplications(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strId, ok := vars["appGroupId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid appGroupId")))
		return
	}
	appGroupId := domain.AppGroupId(strId)
	if !appGroupId.Validate() {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid appGroupId")))
		return
	}

	urlParams := r.URL.Query()
	strApplicationType := urlParams.Get("applicationType")
	applicationType := domain.ApplicationType_PROMETHEUS // by default
	if strApplicationType == "" {
		applicationType = domain.ApplicationType_PROMETHEUS
	} else {
		applicationType.FromString(strApplicationType)
	}

	applications, err := h.usecase.GetApplications(appGroupId, applicationType)
	if err != nil {
		log.Error("Failed to get applications err : ", err)
		ErrorJSON(w, err)
		return
	}

	var out domain.GetApplicationsResponse
	out.Applications = make([]domain.ApplicationResponse, len(applications))
	for i, application := range applications {
		if err := domain.Map(application, &out.Applications[i]); err != nil {
			log.Info(err)
			continue
		}
	}

	ResponseJSON(w, http.StatusOK, out)
}

// UpdateApplication godoc
// @Tags AppGroups
// @Summary Update application
// @Description Update application
// @Accept json
// @Produce json
// @Param object body domain.UpdateApplicationRequest true "body"
// @Success 200 {object} nil
// @Router /app-groups/{appGroupId}/applications [post]
// @Security     JWT
func (h *AppGroupHandler) UpdateApplication(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strId, ok := vars["appGroupId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid appGroupId")))
		return
	}
	appGroupId := domain.AppGroupId(strId)
	if !appGroupId.Validate() {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid appGroupId")))
		return
	}

	input := domain.UpdateApplicationRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, httpErrors.NewBadRequestError(err))
		return
	}

	var dto domain.Application
	if err := domain.Map(input, &dto); err != nil {
		log.Info(err)
	}

	err = h.usecase.UpdateApplication(dto)
	if err != nil {
		log.Error("Failed to update application err : ", err)
		ErrorJSON(w, err)
		return
	}

	ResponseJSON(w, http.StatusOK, nil)
}
