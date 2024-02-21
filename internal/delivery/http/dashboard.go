package http

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/internal/serializer"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
)

type DashboardHandler struct {
	usecase usecase.IDashboardUsecase
}

func NewDashboardHandler(h usecase.Usecase) *DashboardHandler {
	return &DashboardHandler{
		usecase: h.Dashboard,
	}
}

// GetCharts godoc
// @Tags        Dashboards
// @Summary     Get charts data
// @Description Get charts data
// @Accept      json
// @Produce     json
// @Param       organizationId path     string true  "organizationId"
// @Param       chartType      query    string false "chartType"
// @Param       duration       query    string true  "duration"
// @Param       interval       query    string true  "interval"
// @Success     200            {object} domain.GetDashboardChartsResponse
// @Router      /organizations/{organizationId}/dashboard/charts [get]
// @Security    JWT
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
		if err := serializer.Map(chart, &out.Charts[i]); err != nil {
			log.InfoWithContext(r.Context(), err)
			continue
		}
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetCharts godoc
// @Tags        Dashboards
// @Summary     Get chart data
// @Description Get chart data
// @Accept      json
// @Produce     json
// @Param       organizationId path     string true "organizationId"
// @Param       chartType      path     string true "chartType"
// @Param       duration       query    string true "duration"
// @Param       interval       query    string true "interval"
// @Success     200            {object} domain.GetDashboardChartResponse
// @Router      /organizations/{organizationId}/dashboard/charts/{chartType} [get]
// @Security    JWT
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
	if err := serializer.Map(charts[0], &out); err != nil {
		log.InfoWithContext(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetStacks godoc
// @Tags        Dashboards
// @Summary     Get stacks
// @Description Get stacks
// @Accept      json
// @Produce     json
// @Param       organizationId path     string true "organizationId"
// @Success     200            {object} domain.GetDashboardStacksResponse
// @Router      /organizations/{organizationId}/dashboard/stacks [get]
// @Security    JWT
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
		if err := serializer.Map(stack, &out.Stacks[i]); err != nil {
			log.InfoWithContext(r.Context(), err)
			continue
		}
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetResources godoc
// @Tags        Dashboards
// @Summary     Get resources
// @Description Get resources
// @Accept      json
// @Produce     json
// @Param       organizationId path     string true "organizationId"
// @Success     200            {object} domain.GetDashboardResourcesResponse
// @Router      /organizations/{organizationId}/dashboard/resources [get]
// @Security    JWT
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
	if err := serializer.Map(resources, &out.Resources); err != nil {
		log.InfoWithContext(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
}
