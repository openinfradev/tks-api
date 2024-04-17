package http

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
	"github.com/openinfradev/tks-api/internal/model"
	policytemplate "github.com/openinfradev/tks-api/internal/policy-template"
	"github.com/openinfradev/tks-api/internal/serializer"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
	"net/http"
	"strings"
	"time"
)

type IDashboardHandler interface {
	CreateDashboard(w http.ResponseWriter, r *http.Request)
	GetDashboard(w http.ResponseWriter, r *http.Request)
	UpdateDashboard(w http.ResponseWriter, r *http.Request)
	GetCharts(w http.ResponseWriter, r *http.Request)
	GetChart(w http.ResponseWriter, r *http.Request)
	GetStacks(w http.ResponseWriter, r *http.Request)
	GetResources(w http.ResponseWriter, r *http.Request)
	GetPolicyStatus(w http.ResponseWriter, r *http.Request)
	GetPolicyUpdate(w http.ResponseWriter, r *http.Request)
	GetPolicyEnforcement(w http.ResponseWriter, r *http.Request)
	GetPolicyViolation(w http.ResponseWriter, r *http.Request)
	GetPolicyViolationLog(w http.ResponseWriter, r *http.Request)
}

type DashboardHandler struct {
	usecase             usecase.IDashboardUsecase
	organizationUsecase usecase.IOrganizationUsecase
}

func NewDashboardHandler(h usecase.Usecase) IDashboardHandler {
	return &DashboardHandler{
		usecase:             h.Dashboard,
		organizationUsecase: h.Organization,
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
//	@Param			organizationId	path	string	true	"Organization ID"
//	@Success		200				{array}	domain.GetDashboardResponse
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
//	@Tags			Dashboard Widgets
//	@Summary		Get charts data
//	@Description	Get charts data
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"organizationId"
//	@Param			chartType		query		string	false	"chartType"
//	@Param			duration		query		string	true	"duration"
//	@Param			interval		query		string	true	"interval"
//	@Success		200				{object}	domain.GetDashboardChartsResponse
//	@Router			/organizations/{organizationId}/dashboards/charts [get]
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
//	@Tags			Dashboard Widgets
//	@Summary		Get chart data
//	@Description	Get chart data
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"organizationId"
//	@Param			chartType		path		string	true	"chartType"
//	@Param			duration		query		string	true	"duration"
//	@Param			interval		query		string	true	"interval"
//	@Success		200				{object}	domain.GetDashboardChartResponse
//	@Router			/organizations/{organizationId}/dashboards/charts/{chartType} [get]
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
//	@Tags			Dashboard Widgets
//	@Summary		Get stacks
//	@Description	Get stacks
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"organizationId"
//	@Success		200				{object}	domain.GetDashboardStacksResponse
//	@Router			/organizations/{organizationId}/dashboards/stacks [get]
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
//	@Tags			Dashboard Widgets
//	@Summary		Get resources
//	@Description	Get resources
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"organizationId"
//	@Success		200				{object}	domain.GetDashboardResourcesResponse
//	@Router			/organizations/{organizationId}/dashboards/resources [get]
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

// GetPolicyStatus godoc
//
//	@Tags			Dashboard Widgets
//	@Summary		Get policy status
//	@Description	Get policy status
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"Organization ID"
//	@Success		200				{object}	domain.GetDashboardPolicyStatusResponse
//	@Router			/organizations/{organizationId}/dashboards/policy-status [get]
//	@Security		JWT
func (h *DashboardHandler) GetPolicyStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("%s: invalid organizationId", organizationId),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	organization, err := h.organizationUsecase.Get(r.Context(), organizationId)
	if err != nil {
		log.Error(r.Context(), "Failed to retrieve organization")
		ErrorJSON(w, r, fmt.Errorf("failed to retrieve organization"))
		return
	}

	tksClusters, err := policytemplate.GetTksClusterCRs(r.Context(), organization.PrimaryClusterId)
	if err != nil {
		log.Error(r.Context(), "Failed to retrieve tkscluster list", err)
		ErrorJSON(w, r, err)
		return
	}

	var policyStatus domain.DashboardPolicyStatus
	for _, c := range tksClusters {
		switch status := c.Status.TKSProxy.Status; status {
		case "ready":
			policyStatus.Normal++
		case "warn":
			policyStatus.Warning++
		case "error":
			policyStatus.Error++
		default:
			continue
		}
	}

	var out domain.GetDashboardPolicyStatusResponse
	out.PolicyStatus = policyStatus
	ResponseJSON(w, r, http.StatusOK, out)
}

// GetPolicyUpdate godoc
//
//	@Tags			Dashboard Widgets
//	@Summary		Get the number of policytemplates that need to be updated
//	@Description	Get the number of policytemplates that need to be updated
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"Organization ID"
//	@Success		200				{object}	domain.GetDashboardPolicyUpdateResponse
//	@Router			/organizations/{organizationId}/dashboards/policy-update [get]
//	@Security		JWT
func (h *DashboardHandler) GetPolicyUpdate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("%s: invalid organizationId", organizationId),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	organization, err := h.organizationUsecase.Get(r.Context(), organizationId)
	if err != nil {
		log.Error(r.Context(), "Failed to retrieve organization")
		ErrorJSON(w, r, fmt.Errorf("failed to retrieve organization"))
		return
	}

	policyTemplates, err := policytemplate.GetTksPolicyTemplateCRs(r.Context(), organization.PrimaryClusterId)
	if err != nil {
		log.Error(r.Context(), "Failed to retrieve policytemplate list", err)
		ErrorJSON(w, r, err)
		return
	}
	policies, err := policytemplate.GetTksPolicyCRs(r.Context(), organization.PrimaryClusterId)
	if err != nil {
		log.Error(r.Context(), "Failed to retrieve policy list", err)
		ErrorJSON(w, r, err)
		return
	}

	dpu, err := h.usecase.GetPolicyUpdate(r.Context(), policyTemplates, policies)
	if err != nil {
		log.Error(r.Context(), "Failed to make policy update status", err)
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetDashboardPolicyUpdateResponse
	out.PolicyUpdate = dpu
	ResponseJSON(w, r, http.StatusOK, out)
}

// GetPolicyEnforcement godoc
//
//	@Tags			Dashboard Widgets
//	@Summary		Get the number of policy enforcement
//	@Description	Get the number of policy enforcement
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"Organization ID"
//	@Success		200				{object}	domain.GetDashboardPolicyEnforcementResponse
//	@Router			/organizations/{organizationId}/dashboards/policy-enforcement [get]
//	@Security		JWT
func (h *DashboardHandler) GetPolicyEnforcement(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("%s: invalid organizationId", organizationId),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	organization, err := h.organizationUsecase.Get(r.Context(), organizationId)
	if err != nil {
		log.Error(r.Context(), "Failed to retrieve organization")
		ErrorJSON(w, r, fmt.Errorf("failed to retrieve organization"))
		return
	}

	bcd, err := h.usecase.GetPolicyEnforcement(r.Context(), organizationId, organization.PrimaryClusterId)
	if err != nil {
		log.Error(r.Context(), "Failed to make policy bar chart data", err)
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetDashboardPolicyEnforcementResponse
	out.ChartType = "PolicyEnforcement"
	out.OrganizationId = organizationId
	out.Name = "정책 적용 현황"
	out.Description = "정책 적용 현황 통계 데이터"
	out.ChartData = *bcd
	out.UpdatedAt = time.Now()
	ResponseJSON(w, r, http.StatusOK, out)
}

// GetPolicyViolation godoc
//
//	@Tags			Dashboard Widgets
//	@Summary		Get the number of policy violation
//	@Description	Get the number of policy violation
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"Organization ID"
//	@Param			duration		query		string	true	"duration"
//	@Param			interval		query		string	true	"interval"
//	@Success		200				{object}	domain.GetDashboardPolicyViolationResponse
//	@Router			/organizations/{organizationId}/dashboards/policy-violation [get]
//	@Security		JWT
func (h *DashboardHandler) GetPolicyViolation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("%s: invalid organizationId", organizationId),
			"C_INVALID_ORGANIZATION_ID", ""))
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

	bcd, err := h.usecase.GetPolicyViolation(r.Context(), organizationId, duration, interval)
	if err != nil {
		log.Error(r.Context(), "Failed to make policy bar chart data", err)
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetDashboardPolicyViolationResponse
	out.ChartType = "PolicyViolation"
	out.OrganizationId = organizationId
	out.Name = "정책 위반 현황"
	out.Description = "정책 위반 현황 통계 데이터"
	out.ChartData = *bcd
	out.UpdatedAt = time.Now()
	ResponseJSON(w, r, http.StatusOK, out)
}

// GetPolicyViolationLog godoc
//
//	@Tags			Dashboard Widgets
//	@Summary		Get policy violation log
//	@Description	Get policy violation log
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"Organization ID"
//	@Success		200				{object}	domain.GetDashboardPolicyViolationLogResponse
//	@Router			/organizations/{organizationId}/dashboards/policy-violation-log [get]
//	@Security		JWT
func (h *DashboardHandler) GetPolicyViolationLog(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("%s: invalid organizationId", organizationId),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	out, err := h.usecase.GetPolicyViolationLog(r.Context(), organizationId)
	if err != nil {
		log.Error(r.Context(), "Failed to make policy violation log", err)
		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, out)
}
