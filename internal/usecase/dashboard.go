package usecase

import (
	"fmt"
	"time"

	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/pkg/errors"
)

type IDashboardUsecase interface {
	GetCharts(organizationId string, chartType domain.ChartType, duration string, interval string) (res []domain.DashboardChart, err error)
}

type DashboardUsecase struct {
	organizationRepo repository.IOrganizationRepository
}

func NewDashboardUsecase(r repository.Repository) IDashboardUsecase {
	return &DashboardUsecase{
		organizationRepo: r.Organization,
	}
}

func (u *DashboardUsecase) GetCharts(organizationId string, chartType domain.ChartType, duration string, interval string) (out []domain.DashboardChart, err error) {
	_, err = u.organizationRepo.Get(organizationId)
	if err != nil {
		return nil, errors.Wrap(err, "invalid organization")
	}

	for _, strType := range chartType.All() {
		if chartType != domain.ChartType_ALL && chartType.String() != strType {
			continue
		}

		chart, err := u.getPrometheus(organizationId, strType, duration, interval)
		if err != nil {
			log.Error(err)
			continue
		}

		out = append(out, chart)
	}

	// [TODO] For sample data
	if chartType == domain.ChartType_TRAFFIC {

	}

	return
}

func (u *DashboardUsecase) getPrometheus(organizationId string, chartType string, duration string, interval string) (res domain.DashboardChart, err error) {
	// [TODO] get prometheus
	if chartType == domain.ChartType_TRAFFIC.String() {
		chartData := domain.ChartData{}
		chartData.XAxis.Type = "category"
		chartData.XAxis.Data = []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
		chartData.YAxis.Type = "value"

		chartData.Series = append(chartData.Series, domain.Unit{
			Name: "Cluster in",
			Data: []string{"820", "932", "901", "934", "1290", "1330", "1320"},
		})
		chartData.Series = append(chartData.Series, domain.Unit{
			Name: "Cluster out",
			Data: []string{"730", "860", "793", "821", "1271", "648", "927"},
		})
		return domain.DashboardChart{
			ChartType:      domain.ChartType_TRAFFIC,
			OrganizationId: organizationId,
			Name:           "Traffic",
			Description:    "Traffic 통계 데이터",
			Duration:       duration,
			Interval:       interval,
			ChartData:      chartData,
			UpdatedAt:      time.Now(),
		}, nil
	}

	return domain.DashboardChart{}, fmt.Errorf("No data")
}
