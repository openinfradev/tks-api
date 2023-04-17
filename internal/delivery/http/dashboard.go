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
// @Summary Get chart data
// @Description Get chart data
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
	strType := query.Get("chartType")
	chartType := new(domain.ChartType).FromString(strType)
	if strType == "" {
		chartType = domain.ChartType(domain.ChartType_ALL)
	}

	log.Info("a", chartType)
	duration := query.Get("duration")
	if duration == "" {
		duration = "1d" // default
	}

	interval := query.Get("interval")
	if interval == "" {
		interval = "1d" // default
	}

	charts, err := h.usecase.GetCharts(organizationId, chartType, duration, interval)
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
