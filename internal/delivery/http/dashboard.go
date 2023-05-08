package http

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
)

type DashboardHandler struct {
	usecase usecase.IDashboardUsecase
}

func NewDashboardHandler(h usecase.IDashboardUsecase) *DashboardHandler {
	return &DashboardHandler{
		usecase: h,
	}
}

// GetCharts godoc
// @Tags Dashboards
// @Summary Get charts data
// @Description Get charts data
// @Accept json
// @Produce json
// @Param organizationId path string true "organizationId"
// @Param chartType query string false "chartType"
// @Param duration query string true "duration"
// @Param interval query string true "interval"
// @Success 200 {object} domain.GetDashboardChartsResponse
// @Router /organizations/{organizationId}/dashboard/charts [get]
// @Security     JWT
func (h *DashboardHandler) GetCharts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId")))
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

	charts, err := h.usecase.GetCharts(organizationId, domain.ChartType_ALL, duration, interval, year, month)
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	var out domain.GetDashboardChartsResponse
	out.Charts = make([]domain.DashboardChartResponse, len(charts))
	for i, chart := range charts {
		if err := domain.Map(chart, &out.Charts[i]); err != nil {
			log.Info(err)
			continue
		}
	}

	ResponseJSON(w, http.StatusOK, out)
}

// GetCharts godoc
// @Tags Dashboards
// @Summary Get chart data
// @Description Get chart data
// @Accept json
// @Produce json
// @Param organizationId path string true "organizationId"
// @Param chartType path string true "chartType"
// @Param duration query string true "duration"
// @Param interval query string true "interval"
// @Success 200 {object} domain.GetDashboardChartResponse
// @Router /organizations/{organizationId}/dashboard/charts/{chartType} [get]
// @Security     JWT
func (h *DashboardHandler) GetChart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId")))
		return
	}

	strType, ok := vars["chartType"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid chartType")))
		return
	}
	chartType := new(domain.ChartType).FromString(strType)
	if chartType == domain.ChartType_ERROR {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid chartType")))
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

	charts, err := h.usecase.GetCharts(organizationId, chartType, duration, interval, year, month)
	if err != nil {
		ErrorJSON(w, err)
		return
	}
	if len(charts) < 1 {
		ErrorJSON(w, httpErrors.NewInternalServerError(fmt.Errorf("Not found chart")))
		return
	}

	var out domain.DashboardChartResponse
	if err := domain.Map(charts[0], &out); err != nil {
		log.Info(err)
	}

	ResponseJSON(w, http.StatusOK, out)
}

// GetStacks godoc
// @Tags Dashboards
// @Summary Get stacks
// @Description Get stacks
// @Accept json
// @Produce json
// @Param organizationId path string true "organizationId"
// @Success 200 {object} domain.GetDashboardStacksResponse
// @Router /organizations/{organizationId}/dashboard/stacks [get]
// @Security     JWT
func (h *DashboardHandler) GetStacks(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId")))
		return
	}

	stacks, err := h.usecase.Fetch(organizationId)
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	var out domain.GetDashboardStacksResponse
	out.Stacks = make([]domain.DashboardStackResponse, len(stacks))
	for i, stack := range stacks {
		if err := domain.Map(stack, &out.Stacks[i]); err != nil {
			log.Info(err)
			continue
		}
		log.Info(out.Stacks[i])
	}

	ResponseJSON(w, http.StatusOK, out)
}

// GetResources godoc
// @Tags Dashboards
// @Summary Get resources
// @Description Get resources
// @Accept json
// @Produce json
// @Param organizationId path string true "organizationId"
// @Success 200 {object} domain.GetDashboardResourcesResponse
// @Router /organizations/{organizationId}/dashboard/resources [get]
// @Security     JWT
func (h *DashboardHandler) GetResources(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	_, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId")))
		return
	}

	var out domain.GetDashboardResourcesResponse
	out.Stack = "1 개"
	out.Cpu = "80 %"
	out.Memory = "128 GB"
	out.Storage = "10 TB"

	ResponseJSON(w, http.StatusOK, out)
}
