package http

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/serializer"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
	"net/http"
	"strings"
)

type IDashboardHandler interface {
	CreateDashboard(w http.ResponseWriter, r *http.Request)
	GetDashboard(w http.ResponseWriter, r *http.Request)
	UpdateDashboard(w http.ResponseWriter, r *http.Request)
	GetCharts(w http.ResponseWriter, r *http.Request)
	GetChart(w http.ResponseWriter, r *http.Request)
	GetStacks(w http.ResponseWriter, r *http.Request)
	GetResources(w http.ResponseWriter, r *http.Request)
}

type DashboardHandler struct {
	usecase usecase.IDashboardUsecase
}

func NewDashboardHandler(h usecase.Usecase) IDashboardHandler {
	return &DashboardHandler{
		usecase: h.Dashboard,
	}
}

// CreateDashboard godoc
//
//	@Tags			Dashboards
//	@Summary		Create new dashboard
//	@Description	Create new dashboard
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string							true	"Organization ID"
//	@Param			request			body		domain.CreateDashboardRequest	true	"Request body to create dashboard"
//	@Success		200				{object}	domain.CreateDashboardResponse
//	@Router			/organizations/{organizationId}/dashboards [post]
//	@Security		JWT
func (h *DashboardHandler) CreateDashboard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("%s: invalid organizationId", organizationId),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	var dashboardReq []domain.CreateDashboardRequest
	if err := UnmarshalRequestInput(r, &dashboardReq); err != nil {
		ErrorJSON(w, r, err)
		return
	}
	content, err := MarshalToString(r.Context(), dashboardReq)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	requestUserInfo, ok := request.UserFrom(r.Context())
	if !ok {
		log.Error(r.Context(), "Failed to retrieve user info from request")
		ErrorJSON(w, r, fmt.Errorf("failed to retrieve user info from request"))
	}
	userId := requestUserInfo.GetUserId()

	dashboard, err := h.usecase.GetDashboard(r.Context(), organizationId, userId.String())
	if err == nil && dashboard != nil {
		log.Error(r.Context(), "Dashboard already exists")
		ResponseJSON(w, r, http.StatusInternalServerError, "Dashboard already exists")
		return
	}

	dashboard = &model.Dashboard{
		OrganizationId: organizationId,
		UserId:         userId,
		Content:        content,
		IsAdmin:        false,
	}
	log.Info(r.Context(), "Processing CREATE request for dashboard")

	dashboardId, err := h.usecase.CreateDashboard(r.Context(), dashboard)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return
	}

	out := domain.CreateDashboardResponse{DashboardId: dashboardId}
	ResponseJSON(w, r, http.StatusOK, out)
}

// GetDashboard godoc
//
//	@Tags			Dashboards
//	@Summary		Get dashboard
//	@Description	Get dashboard
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"Organization ID"
//	@Success		200				{object}	domain.GetDashboardResponse
//	@Router			/organizations/{organizationId}/dashboards [get]
//	@Security		JWT
func (h *DashboardHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("%s: invalid organizationId", organizationId),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}
	requestUserInfo, ok := request.UserFrom(r.Context())
	if !ok {
		log.Error(r.Context(), "Failed to retrieve user info from request")
		ErrorJSON(w, r, fmt.Errorf("failed to retrieve user info from request"))
	}
	userId := requestUserInfo.GetUserId().String()

	dashboard, err := h.usecase.GetDashboard(r.Context(), organizationId, userId)
	if err != nil {
		log.Error(r.Context(), "Failed to retrieve dashboard", err)
		ErrorJSON(w, r, err)
		return
	}
	if dashboard == nil {
		ResponseJSON(w, r, http.StatusOK, nil)
		return
	}
	if len(dashboard.Content) == 0 {
		ResponseJSON(w, r, http.StatusOK, []domain.GetDashboardResponse{})
		return
	}

	var dashboardRes []domain.GetDashboardResponse
	if err := UnmarshalFromString(r.Context(), dashboard.Content, &dashboardRes); err != nil {
		ErrorJSON(w, r, err)
	}
	ResponseJSON(w, r, http.StatusOK, dashboardRes)
}

// UpdateDashboard godoc
//
//	@Tags			Dashboards
//	@Summary		Update dashboard
//	@Description	Update dashboard
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string							true	"Organization ID"
//	@Param			request			body		domain.UpdateDashboardRequest	true	"Request body to update dashboard"
//	@Success		200				{object}	domain.CommonDashboardResponse
//	@Router			/organizations/{organizationId}/dashboards [put]
//	@Security		JWT
func (h *DashboardHandler) UpdateDashboard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("%s: invalid organizationId", organizationId),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	var dashboardReq []domain.CreateDashboardRequest
	if err := UnmarshalRequestInput(r, &dashboardReq); err != nil {
		ErrorJSON(w, r, err)
		return
	}
	content, err := MarshalToString(r.Context(), dashboardReq)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	requestUserInfo, ok := request.UserFrom(r.Context())
	if !ok {
		log.Error(r.Context(), "Failed to retrieve user info from request")
		ErrorJSON(w, r, fmt.Errorf("failed to retrieve user info from request"))
	}
	userId := requestUserInfo.GetUserId().String()

	dashboard, err := h.usecase.GetDashboard(r.Context(), organizationId, userId)
	if err != nil || dashboard == nil {
		log.Error(r.Context(), "Failed to retrieve dashboard", err)
		ErrorJSON(w, r, err)
		return
	}

	dashboard.Content = content
	if err := h.usecase.UpdateDashboard(r.Context(), dashboard); err != nil {
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "", ""))
		return
	}
	ResponseJSON(w, r, http.StatusOK, domain.CommonDashboardResponse{Result: "OK"})
}

// GetCharts godoc
//
//	@Tags			Dashboards
//	@Summary		Get charts data
//	@Description	Get charts data
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"organizationId"
//	@Param			chartType		query		string	false	"chartType"
//	@Param			duration		query		string	true	"duration"
//	@Param			interval		query		string	true	"interval"
//	@Success		200				{object}	domain.GetDashboardChartsResponse
//	@Router			/organizations/{organizationId}/dashboard/charts [get]
//	@Security		JWT
func (h *DashboardHandler) GetCharts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	query := r.URL.Query()
	duration := query.Get("duration")
	if duration == "" {
		duration = "1d" // default
	}

	interval := query.Get("interval")
	if interval == "" {
		interval = "1d" // default
	}
	year := query.Get("year")
	if year == "" {
		year = "2023" // default
	}

	month := query.Get("month")
	if month == "" {
		month = "5" // default
	}

	charts, err := h.usecase.GetCharts(r.Context(), organizationId, domain.ChartType_ALL, duration, interval, year, month)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetDashboardChartsResponse
	out.Charts = make([]domain.DashboardChartResponse, len(charts))
	for i, chart := range charts {
		if err := serializer.Map(r.Context(), chart, &out.Charts[i]); err != nil {
			log.Info(r.Context(), err)
			continue
		}
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetChart godoc
//
//	@Tags			Dashboards
//	@Summary		Get chart data
//	@Description	Get chart data
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"organizationId"
//	@Param			chartType		path		string	true	"chartType"
//	@Param			duration		query		string	true	"duration"
//	@Param			interval		query		string	true	"interval"
//	@Success		200				{object}	domain.GetDashboardChartResponse
//	@Router			/organizations/{organizationId}/dashboard/charts/{chartType} [get]
//	@Security		JWT
func (h *DashboardHandler) GetChart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	strType, ok := vars["chartType"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid chartType"), "D_INVALID_CHART_TYPE", ""))
		return
	}
	chartType := new(domain.ChartType).FromString(strType)
	if chartType == domain.ChartType_ERROR {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid chartType"), "D_INVALID_CHART_TYPE", ""))
		return
	}

	query := r.URL.Query()
	duration := query.Get("duration")
	if duration == "" {
		duration = "1d" // default
	}

	interval := query.Get("interval")
	if interval == "" {
		interval = "1d" // default
	}

	year := query.Get("year")
	if year == "" {
		year = "2023" // default
	}

	month := query.Get("month")
	if month == "" {
		month = "4" // default
	}

	charts, err := h.usecase.GetCharts(r.Context(), organizationId, chartType, duration, interval, year, month)
	if err != nil {
		if strings.Contains(err.Error(), "Invalid primary clusterId") {
			ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "D_INVALID_PRIMARY_STACK", ""))
			return
		}

		ErrorJSON(w, r, err)
		return
	}
	if len(charts) < 1 {
		ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "D_NOT_FOUND_CHART", ""))
		return
	}

	var out domain.DashboardChartResponse
	if err := serializer.Map(r.Context(), charts[0], &out); err != nil {
		log.Info(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetStacks godoc
//
//	@Tags			Dashboards
//	@Summary		Get stacks
//	@Description	Get stacks
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"organizationId"
//	@Success		200				{object}	domain.GetDashboardStacksResponse
//	@Router			/organizations/{organizationId}/dashboard/stacks [get]
//	@Security		JWT
func (h *DashboardHandler) GetStacks(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	stacks, err := h.usecase.GetStacks(r.Context(), organizationId)
	if err != nil {
		if strings.Contains(err.Error(), "Invalid primary clusterId") {
			ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "D_INVALID_PRIMARY_STACK", ""))
			return
		}

		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetDashboardStacksResponse
	out.Stacks = make([]domain.DashboardStackResponse, len(stacks))
	for i, stack := range stacks {
		if err := serializer.Map(r.Context(), stack, &out.Stacks[i]); err != nil {
			log.Info(r.Context(), err)
			continue
		}
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetResources godoc
//
//	@Tags			Dashboards
//	@Summary		Get resources
//	@Description	Get resources
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"organizationId"
//	@Success		200				{object}	domain.GetDashboardResourcesResponse
//	@Router			/organizations/{organizationId}/dashboard/resources [get]
//	@Security		JWT
func (h *DashboardHandler) GetResources(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	resources, err := h.usecase.GetResources(r.Context(), organizationId)
	if err != nil {
		if strings.Contains(err.Error(), "Invalid primary clusterId") {
			ErrorJSON(w, r, httpErrors.NewInternalServerError(err, "D_INVALID_PRIMARY_STACK", ""))
			return
		}
		ErrorJSON(w, r, err)
		return
	}
	var out domain.GetDashboardResourcesResponse
	if err := serializer.Map(r.Context(), resources, &out.Resources); err != nil {
		log.Info(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
}
