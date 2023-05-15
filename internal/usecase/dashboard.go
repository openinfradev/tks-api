package usecase

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/openinfradev/tks-api/internal/kubernetes"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
	thanos "github.com/openinfradev/tks-api/pkg/thanos-client"
	gcache "github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/thoas/go-funk"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	cache            *gcache.Cache
}

func NewDashboardUsecase(r repository.Repository, cache *gcache.Cache) IDashboardUsecase {
	return &DashboardUsecase{
		organizationRepo: r.Organization,
		clusterRepo:      r.Cluster,
		appGroupRepo:     r.AppGroup,
		cache:            cache,
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
	thanosUrl, err := u.getThanosUrl(organizationId)
	if err != nil {
		return out, err
	}
	arr := strings.Split(thanosUrl, ":")
	address := arr[0] + ":" + arr[1]
	log.Info(address)
	port, _ := strconv.Atoi(arr[2])
	thanosClient, err := thanos.New(address, port, false, "")
	if err != nil {
		return out, errors.Wrap(err, "failed to create thanos client")
	}

	// Stack
	clusters, err := u.clusterRepo.FetchByOrganizationId(organizationId)
	if err != nil {
		return out, err
	}

	log.Info(len(clusters))

	filteredClusters := funk.Filter(clusters, func(x domain.Cluster) bool {
		return x.Status != domain.ClusterStatus_DELETED
	})
	if filteredClusters != nil {
		out.Stack = fmt.Sprintf("%d 개", len(filteredClusters.([]domain.Cluster)))
	} else {
		out.Stack = "0 개"
	}

	// CPU
	/*
		{"data":{"result":[{"metric":{"taco_cluster":"cmsai5k5l"},"value":[1683608185.65,"32"]},{"metric":{"taco_cluster":"crjfh12oc"},"value":[1683608185.65,"12"]}],"vector":""},"status":"success"}
	*/
	result, err := thanosClient.Get("sum by (taco_cluster) (machine_cpu_cores)")
	if err != nil {
		return out, err
	}
	cpu := 0
	for _, val := range result.Data.Result {
		cpuVal, err := strconv.Atoi(val.Value[1].(string))
		if err != nil {
			continue
		}
		if cpuVal > 0 {
			cpu = cpu + cpuVal
		}
	}
	out.Cpu = fmt.Sprintf("%d 개", cpu)

	// Memory
	result, err = thanosClient.Get("sum by (taco_cluster) (machine_memory_bytes)")
	if err != nil {
		return out, err
	}
	memory := 0
	for _, val := range result.Data.Result {
		memoryVal, err := strconv.Atoi(val.Value[1].(string))
		if err != nil {
			continue
		}
		if memoryVal > 0 {
			memoryVal = memoryVal / 1024 / 1024 / 1024
			memory = memory + memoryVal
		}
	}
	out.Memory = fmt.Sprintf("%d GB", memory)

	// Storage
	result, err = thanosClient.Get("sum by (taco_cluster) (kubelet_volume_stats_capacity_bytes)")
	if err != nil {
		return out, err
	}
	storage := 0
	for _, val := range result.Data.Result {
		storageVal, err := strconv.Atoi(val.Value[1].(string))
		if err != nil {
			continue
		}
		if storageVal > 0 {
			storageVal = storageVal / 1024 / 1024 / 1024
			storage = storage + storageVal
		}
	}
	out.Storage = fmt.Sprintf("%d GB", storage)

	return
}

func (u *DashboardUsecase) getPrometheus(organizationId string, chartType string, duration string, interval string, year string, month string) (res domain.DashboardChart, err error) {
	thanosUrl, err := u.getThanosUrl(organizationId)
	if err != nil {
		return res, err
	}
	arr := strings.Split(thanosUrl, ":")
	address := arr[0] + ":" + arr[1]
	log.Info(address)
	port, _ := strconv.Atoi(arr[2])
	thanosClient, err := thanos.New(address, port, false, "")
	if err != nil {
		return res, errors.Wrap(err, "failed to create thanos client")
	}

	now := time.Now()
	chartData := domain.ChartData{}

	durationSec := 60 * 60 * 24
	switch duration {
	case "1h":
		durationSec = 60 * 60
	case "1d":
		durationSec = 60 * 60 * 24
	case "7d":
		durationSec = 60 * 60 * 24 * 7
	case "30d":
		durationSec = 60 * 60 * 24 * 30
	}

	intervalSec := 60 * 60 // default 1h
	switch interval {
	case "1h":
		intervalSec = 60 * 60
	case "1d":
		intervalSec = 60 * 60 * 24
	case "7h":
		intervalSec = 60 * 60 * 24 * 7
	}

	switch chartType {
	case domain.ChartType_CPU.String():
		query := "sum (avg(1-rate(node_cpu_seconds_total{mode=\"idle\"}[1h])) by (taco_cluster))"
		result, err := thanosClient.FetchRange(query, int(now.Unix())-durationSec, int(now.Unix()), intervalSec)
		if err != nil {
			return res, err
		}
		xAxisData := []string{}
		yAxisData := []string{}
		for _, val := range result.Data.Result {
			for _, vals := range val.Values {
				x := int(math.Round(vals.([]interface{})[0].(float64)))
				y, err := strconv.ParseFloat(vals.([]interface{})[1].(string), 32)
				if err != nil {
					y = 0
				}
				y = y * 100

				xAxisData = append(xAxisData, strconv.Itoa(x))
				yAxisData = append(yAxisData, fmt.Sprintf("%f", y))
			}
		}
		chartData.XAxis.Data = xAxisData
		chartData.Series = append(chartData.Series, domain.Unit{
			Name: "CPU 사용량",
			Data: yAxisData,
		})

	case domain.ChartType_MEMORY.String():
		query := "sum (sum(node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes) by (taco_cluster) / sum(node_memory_MemTotal_bytes) by (taco_cluster))"
		result, err := thanosClient.FetchRange(query, int(now.Unix())-durationSec, int(now.Unix()), intervalSec)
		if err != nil {
			return res, err
		}
		xAxisData := []string{}
		yAxisData := []string{}
		for _, val := range result.Data.Result {
			for _, vals := range val.Values {
				x := int(math.Round(vals.([]interface{})[0].(float64)))
				y, err := strconv.ParseFloat(vals.([]interface{})[1].(string), 32)
				if err != nil {
					y = 0
				}
				y = y * 100
				xAxisData = append(xAxisData, strconv.Itoa(x))
				yAxisData = append(yAxisData, fmt.Sprintf("%f", y))
			}
		}
		chartData.XAxis.Data = xAxisData
		chartData.Series = append(chartData.Series, domain.Unit{
			Name: "Memory 사용량",
			Data: yAxisData,
		})
	case domain.ChartType_POD.String():
		query := "sum(increase(kube_pod_container_status_restarts_total{namespace!=\"kube-system\"}[1h]))"
		result, err := thanosClient.FetchRange(query, int(now.Unix())-durationSec, int(now.Unix()), intervalSec)
		if err != nil {
			return res, err
		}
		xAxisData := []string{}
		yAxisData := []string{}
		for _, val := range result.Data.Result {
			for _, vals := range val.Values {
				x := int(math.Round(vals.([]interface{})[0].(float64)))
				y, err := strconv.ParseFloat(vals.([]interface{})[1].(string), 32)
				if err != nil {
					y = 0
				}
				xAxisData = append(xAxisData, strconv.Itoa(x))
				yAxisData = append(yAxisData, fmt.Sprintf("%f", y))
			}
		}
		chartData.XAxis.Data = xAxisData
		chartData.Series = append(chartData.Series, domain.Unit{
			Name: "POD 재기동",
			Data: yAxisData,
		})
	case domain.ChartType_TRAFFIC.String():
		query := "sum(rate(container_network_receive_bytes_total[1h]))"
		result, err := thanosClient.FetchRange(query, int(now.Unix())-durationSec, int(now.Unix()), intervalSec)
		if err != nil {
			return res, err
		}
		xAxisData := []string{}
		yAxisData := []string{}
		for _, val := range result.Data.Result {
			for _, vals := range val.Values {
				x := int(math.Round(vals.([]interface{})[0].(float64)))
				y, err := strconv.ParseFloat(vals.([]interface{})[1].(string), 32)
				if err != nil {
					y = 0
				}
				xAxisData = append(xAxisData, strconv.Itoa(x))
				yAxisData = append(yAxisData, fmt.Sprintf("%f", y))
			}
		}
		chartData.XAxis.Data = xAxisData
		chartData.Series = append(chartData.Series, domain.Unit{
			Name: "Traffic IN",
			Data: yAxisData,
		})
	case domain.ChartType_POD_CALENDAR.String():
		query := "sum(increase(kube_pod_container_status_restarts_total{namespace!=\"kube-system\"}[1h]))"
		result, err := thanosClient.FetchRange(query, int(now.Unix())-(60*60*24*30), int(now.Unix()), 60*60*24)
		if err != nil {
			return res, err
		}
		xAxisData := []string{}
		yAxisData := []string{}
		for _, val := range result.Data.Result {
			for _, vals := range val.Values {
				x := int(math.Round(vals.([]interface{})[0].(float64)))
				y, err := strconv.ParseFloat(vals.([]interface{})[1].(string), 32)
				if err != nil {
					y = 0
				}
				xAxisData = append(xAxisData, strconv.Itoa(x))
				yAxisData = append(yAxisData, fmt.Sprintf("%f", y))
			}
		}
		chartData.XAxis.Data = xAxisData
		chartData.Series = append(chartData.Series, domain.Unit{
			Name: "date",
			Data: xAxisData,
		})
		chartData.Series = append(chartData.Series, domain.Unit{
			Name: "podRestartCount",
			Data: yAxisData,
		})
		chartData.Series = append(chartData.Series, domain.Unit{
			Name: "totalPodCount",
			Data: []string{"100"},
		})

		/*
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
		*/
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
	default:
		return domain.DashboardChart{}, fmt.Errorf("No data")
	}

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

}

func (u *DashboardUsecase) getThanosUrl(organizationId string) (out string, err error) {
	const prefix = "CACHE_KEY_THANOS_URL"
	value, found := u.cache.Get(prefix + organizationId)
	if found {
		return value.(string), nil
	}

	organization, err := u.organizationRepo.Get(organizationId)
	if err != nil {
		return out, errors.Wrap(err, "Failed to get organization")
	}

	//organization.PrimaryClusterId = "c6ayyhbul"
	if organization.PrimaryClusterId == "" {
		return out, fmt.Errorf("Invalid primary clusterId")
	}

	clientset_user, err := kubernetes.GetClientFromClusterId(organization.PrimaryClusterId)
	if err != nil {
		return out, err
	}
	service, err := clientset_user.CoreV1().Services("lma").Get(context.TODO(), "thanos-query", metav1.GetOptions{})
	if err != nil {
		return out, errors.Wrap(err, "Failed to get services.")
	}

	// LoadBalaner 일경우, aws address 형태의 경우만 가정한다.
	if service.Spec.Type != "LoadBalancer" {
		return out, fmt.Errorf("Service type is not LoadBalancer. [%s] ", service.Spec.Type)
	}

	lbs := service.Status.LoadBalancer.Ingress
	ports := service.Spec.Ports
	if len(lbs) > 0 && len(ports) > 0 {
		out = ports[0].TargetPort.StrVal + "://" + lbs[0].Hostname + ":" + strconv.Itoa(int(ports[0].Port))
		u.cache.Set(prefix+organizationId, out, gcache.DefaultExpiration)
	}

	return
}
