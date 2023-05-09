package usecase

import (
	"fmt"
	"strconv"
	"time"

	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/openinfradev/tks-api/pkg/thanos-client"
	"github.com/pkg/errors"
	"github.com/thoas/go-funk"
)

type IDashboardUsecase interface {
	GetCharts(organizationId string, chartType domain.ChartType, duration string, interval string, year string, month string) (res []domain.DashboardChart, err error)
	GetStacks(organizationId string) (out []domain.DashboardStack, err error)
	GetResources(organizationId string) (out domain.DashboardResource, err error)
}

type DashboardUsecase struct {
	organizationRepo repository.IOrganizationRepository
	clusterRepo      repository.IClusterRepository
	appGroupRepo     repository.IAppGroupRepository
	thanosClient     thanos.ThanosClient
}

func NewDashboardUsecase(r repository.Repository, thanos thanos.ThanosClient) IDashboardUsecase {
	return &DashboardUsecase{
		organizationRepo: r.Organization,
		clusterRepo:      r.Cluster,
		appGroupRepo:     r.AppGroup,
		thanosClient:     thanos,
	}
}

func (u *DashboardUsecase) GetCharts(organizationId string, chartType domain.ChartType, duration string, interval string, year string, month string) (out []domain.DashboardChart, err error) {
	_, err = u.organizationRepo.Get(organizationId)
	if err != nil {
		return nil, errors.Wrap(err, "invalid organization")
	}

	for _, strType := range chartType.All() {
		if chartType != domain.ChartType_ALL && chartType.String() != strType {
			continue
		}

		chart, err := u.getPrometheus(organizationId, strType, duration, interval, year, month)
		if err != nil {
			log.Error(err)
			continue
		}

		out = append(out, chart)
	}

	return
}

func (u *DashboardUsecase) GetStacks(organizationId string) (out []domain.DashboardStack, err error) {
	clusters, err := u.clusterRepo.FetchByOrganizationId(organizationId)
	if err != nil {
		return out, err
	}

	for _, cluster := range clusters {
		appGroups, err := u.appGroupRepo.Fetch(cluster.ID)
		if err != nil {
			return nil, err
		}
		stack := reflectClusterToStack(cluster, appGroups)
		dashboardStack := domain.DashboardStack{}
		if err := domain.Map(stack, &dashboardStack); err != nil {
			log.Info(err)
		}

		// [TODO]
		dashboardStack.Cpu = "30 %"
		dashboardStack.Memory = "128 GB"
		dashboardStack.Storage = "20 TB"

		out = append(out, dashboardStack)
	}

	return
}

func (u *DashboardUsecase) GetResources(organizationId string) (out domain.DashboardResource, err error) {

	result, err := u.thanosClient.Get("sum by (taco_cluster) (machine_cpu_cores)")
	if err != nil {
		return out, err
	}

	// Stack
	clusters, err := u.clusterRepo.FetchByOrganizationId(organizationId)
	if err != nil {
		return out, err
	}

	filteredClusters := funk.Find(clusters, func(x domain.Cluster) bool {
		return x.Status != domain.ClusterStatus_DELETED
	})
	out.Stack = fmt.Sprintf("%s 개", len(filteredClusters.([]domain.Cluster)))

	/*
		{"data":{"result":[{"metric":{"taco_cluster":"cmsai5k5l"},"value":[1683608185.65,"32"]},{"metric":{"taco_cluster":"crjfh12oc"},"value":[1683608185.65,"12"]}],"vector":""},"status":"success"}
	*/

	// CPU
	cpu := 0
	for _, val := range result.Data.Result {
		clusterId := val.Metric.TacoCluster
		log.Info(clusterId)

		cpuVal, err := strconv.Atoi(val.Value[1].(string))
		if err != nil {
			continue
		}
		cpu = cpu + cpuVal
	}
	out.Cpu = fmt.Sprintf("%d 개", cpu)

	// Memory
	// machine_memory_bytes

	// Storage

	return
}

func (u *DashboardUsecase) getPrometheus(organizationId string, chartType string, duration string, interval string, year string, month string) (res domain.DashboardChart, err error) {
	// [TODO] get prometheus
	switch chartType {
	case domain.ChartType_TRAFFIC.String(),
		domain.ChartType_CPU.String(),
		domain.ChartType_CPU.String(),
		domain.ChartType_POD.String(),
		domain.ChartType_MEMORY.String():

		chartData := domain.ChartData{}
		chartData.XAxis.Data = []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}

		chartData.Series = append(chartData.Series, domain.Unit{
			Name: "Cluster in",
			Data: []string{"820", "932", "901", "934", "1290", "1330", "1320"},
		})
		chartData.Series = append(chartData.Series, domain.Unit{
			Name: "Cluster out",
			Data: []string{"730", "860", "793", "821", "1271", "648", "927"},
		})

		return domain.DashboardChart{
			ChartType:      new(domain.ChartType).FromString(chartType),
			OrganizationId: organizationId,
			Name:           chartType,
			Description:    chartType + " 통계 데이터",
			Duration:       duration,
			Interval:       interval,
			ChartData:      chartData,
			UpdatedAt:      time.Now(),
		}, nil

	case domain.ChartType_POD_CALENDAR.String():
		chartData := domain.ChartData{}
		chartData.Series = append(chartData.Series, domain.Unit{
			Name: "date",
			Data: []string{"2021-04-01", "2021-04-02", "2021-04-03"},
		})
		chartData.Series = append(chartData.Series, domain.Unit{
			Name: "podRestartCount",
			Data: []string{"1", "4", "0"},
		})
		chartData.Series = append(chartData.Series, domain.Unit{
			Name: "totalPodCount",
			Data: []string{"100", "120", "100"},
		})
		return domain.DashboardChart{
			ChartType:      domain.ChartType_POD_CALENDAR,
			OrganizationId: organizationId,
			Name:           "POD 기동 현황",
			Description:    "Pod 재기동 수 / 총 Pod 수",
			Year:           year,
			Month:          month,
			ChartData:      chartData,
			UpdatedAt:      time.Now(),
		}, nil

	}

	return domain.DashboardChart{}, fmt.Errorf("No data")
}
