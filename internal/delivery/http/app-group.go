package http

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/internal/helper"
	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
	"github.com/openinfradev/tks-api/internal/pagination"
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
		ErrorJSON(w, r, err)
		return
	}

	var dto domain.AppGroup
	if err = domain.Map(input, &dto); err != nil {
		log.InfoWithContext(r.Context(), err)
	}

	appGroupId, err := h.usecase.Create(r.Context(), dto)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.CreateAppGroupResponse
	out.ID = appGroupId.String()

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetAppGroups godoc
// @Tags AppGroups
// @Summary Get appGroup list
// @Description Get appGroup list by giving params
// @Accept json
// @Produce json
// @Param clusterId query string false "clusterId"
// @Param limit query string false "pageSize"
// @Param page query string false "pageNumber"
// @Param soertColumn query string false "sortColumn"
// @Param sortOrder query string false "sortOrder"
// @Param filters query []string false "filters"
// @Success 200 {object} domain.GetAppGroupsResponse
// @Router /app-groups [get]
// @Security     JWT
func (h *AppGroupHandler) GetAppGroups(w http.ResponseWriter, r *http.Request) {
	urlParams := r.URL.Query()

	clusterId := urlParams.Get("clusterId")
	if clusterId == "" || !helper.ValidateClusterId(clusterId) {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid clusterId"), "C_INVALID_CLUSTER_ID", ""))
		return
	}
	pg := pagination.NewPagination(&urlParams)

	appGroups, err := h.usecase.Fetch(r.Context(), domain.ClusterId(clusterId), pg)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetAppGroupsResponse
	out.AppGroups = make([]domain.AppGroupResponse, len(appGroups))
	for i, appGroup := range appGroups {
		if err := domain.Map(appGroup, &out.AppGroups[i]); err != nil {
			log.InfoWithContext(r.Context(), err)
			continue
		}
	}

	if err := domain.Map(*pg, &out.Pagination); err != nil {
		log.InfoWithContext(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
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
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid appGroupId"), "C_INVALID_APPGROUP_ID", ""))
		return
	}
	appGroupId := domain.AppGroupId(strId)
	if !appGroupId.Validate() {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid appGroupId"), "C_INVALID_APPGROUP_ID", ""))
		return
	}
	appGroup, err := h.usecase.Get(r.Context(), appGroupId)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetAppGroupResponse
	if err := domain.Map(appGroup, &out.AppGroup); err != nil {
		log.InfoWithContext(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
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
	user, ok := request.UserFrom(r.Context())
	if !ok {
		ErrorJSON(w, r, httpErrors.NewUnauthorizedError(fmt.Errorf("Invalid token"), "A_INVALID_TOKEN", ""))
		return
	}

	vars := mux.Vars(r)
	strId, ok := vars["appGroupId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid appGroupId"), "C_INVALID_APPGROUP_ID", ""))
		return
	}

	appGroupId := domain.AppGroupId(strId)
	if !appGroupId.Validate() {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid appGroupId"), "C_INVALID_APPGROUP_ID", ""))
		return
	}

	err := h.usecase.Delete(r.Context(), user.GetOrganizationId(), appGroupId)
	if err != nil {
		log.ErrorWithContext(r.Context(), "Failed to delete appGroup err : ", err)
		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, nil)
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
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid appGroupId"), "C_INVALID_APPGROUP_ID", ""))
		return
	}
	appGroupId := domain.AppGroupId(strId)
	if !appGroupId.Validate() {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid appGroupId"), "C_INVALID_APPGROUP_ID", ""))
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

	applications, err := h.usecase.GetApplications(r.Context(), appGroupId, applicationType)
	if err != nil {
		log.ErrorWithContext(r.Context(), "Failed to get applications err : ", err)
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetApplicationsResponse
	out.Applications = make([]domain.ApplicationResponse, len(applications))
	for i, application := range applications {
		if err := domain.Map(application, &out.Applications[i]); err != nil {
			log.InfoWithContext(r.Context(), err)
			continue
		}
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// CreateApplication godoc
// @Tags AppGroups
// @Summary Create application
// @Description Create application
// @Accept json
// @Produce json
// @Param object body domain.CreateApplicationRequest true "body"
// @Success 200 {object} nil
// @Router /app-groups/{appGroupId}/applications [post]
// @Security     JWT
func (h *AppGroupHandler) CreateApplication(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strId, ok := vars["appGroupId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid appGroupId"), "C_INVALID_APPGROUP_ID", ""))
		return
	}
	appGroupId := domain.AppGroupId(strId)
	if !appGroupId.Validate() {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid appGroupId"), "C_INVALID_APPGROUP_ID", ""))
		return
	}

	input := domain.CreateApplicationRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var dto domain.Application
	if err := domain.Map(input, &dto); err != nil {
		log.InfoWithContext(r.Context(), err)
	}
	dto.AppGroupId = appGroupId

	err = h.usecase.UpdateApplication(r.Context(), dto)
	if err != nil {
		log.ErrorWithContext(r.Context(), "Failed to update application err : ", err)
		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, nil)
}
