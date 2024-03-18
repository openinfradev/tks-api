package usecase

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/helper"
	"github.com/openinfradev/tks-api/internal/kubernetes"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/internal/serializer"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
	thanos "github.com/openinfradev/tks-api/pkg/thanos-client"
	gcache "github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/thoas/go-funk"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/strings/slices"
)

type IDashboardUsecase interface {
	GetCharts(ctx context.Context, organizationId string, chartType domain.ChartType, duration string, interval string, year string, month string) (res []domain.DashboardChart, err error)
	GetStacks(ctx context.Context, organizationId string) (out []domain.DashboardStack, err error)
	GetResources(ctx context.Context, organizationId string) (out domain.DashboardResource, err error)
}

type DashboardUsecase struct {
	organizationRepo       repository.IOrganizationRepository
	clusterRepo            repository.IClusterRepository
	appGroupRepo           repository.IAppGroupRepository
	systemNotificationRepo repository.ISystemNotificationRepository
	cache                  *gcache.Cache
}

func NewDashboardUsecase(r repository.Repository, cache *gcache.Cache) IDashboardUsecase {
	return &DashboardUsecase{
		organizationRepo:       r.Organization,
		clusterRepo:            r.Cluster,
		appGroupRepo:           r.AppGroup,
		systemNotificationRepo: r.SystemNotification,
		cache:                  cache,
	}
}

func (u *DashboardUsecase) GetCharts(ctx context.Context, organizationId string, chartType domain.ChartType, duration string, interval string, year string, month string) (out []domain.DashboardChart, err error) {
	_, err = u.organizationRepo.Get(ctx, organizationId)
	if err != nil {
		return nil, errors.Wrap(err, "invalid organization")
	}

	for _, strType := range chartType.All() {
		if chartType != domain.ChartType_ALL && chartType.String() != strType {
			continue
		}

		chart, err := u.getChartFromPrometheus(ctx, organizationId, strType, duration, interval, year, month)
		if err != nil {
			return nil, err
		}

		out = append(out, chart)
	}

	return
}

func (u *DashboardUsecase) GetStacks(ctx context.Context, organizationId string) (out []domain.DashboardStack, err error) {
	clusters, err := u.clusterRepo.FetchByOrganizationId(ctx, organizationId, uuid.Nil, nil)
	if err != nil {
		return out, err
	}

	thanosUrl, err := u.getThanosUrl(ctx, organizationId)
	if err != nil {
		log.Error(ctx, err)
		return out, httpErrors.NewInternalServerError(err, "D_INVALID_PRIMARY_STACK", "")
	}
	address, port := helper.SplitAddress(ctx, thanosUrl)
	thanosClient, err := thanos.New(address, port, false, "")
	if err != nil {
		return out, errors.Wrap(err, "failed to create thanos client")
	}
	stackMemoryDisk, err := thanosClient.Get(ctx, "sum by (__name__, taco_cluster) ({__name__=~\"node_memory_MemFree_bytes|machine_memory_bytes|kubelet_volume_stats_used_bytes|kubelet_volume_stats_capacity_bytes\"})")
	if err != nil {
		return out, err
	}

	stackCpu, err := thanosClient.Get(ctx, "avg by (taco_cluster) (instance:node_cpu:ratio*100)")
	if err != nil {
		return out, err
	}

	for _, cluster := range clusters {
		appGroups, err := u.appGroupRepo.Fetch(ctx, cluster.ID, nil)
		if err != nil {
			return nil, err
		}
		stack := reflectClusterToStack(ctx, cluster, appGroups)
		dashboardStack := domain.DashboardStack{}
		if err := serializer.Map(ctx, stack, &dashboardStack); err != nil {
			log.Info(ctx, err)
		}

		memory, disk := u.getStackMemoryDisk(stackMemoryDisk.Data.Result, cluster.ID.String())
		cpu := u.getStackCpu(stackCpu.Data.Result, cluster.ID.String())

		if cpu != "" {
			cpu = cpu + "%"
		}
		if memory != "" {
			memory = memory + "%"
		}
		if disk != "" {
			disk = disk + "%"
		}

		dashboardStack.Cpu = cpu
		dashboardStack.Memory = memory
		dashboardStack.Storage = disk

		out = append(out, dashboardStack)
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].Status == domain.StackStatus_RUNNING.String()
	})

	return
}

func (u *DashboardUsecase) GetResources(ctx context.Context, organizationId string) (out domain.DashboardResource, err error) {
	thanosUrl, err := u.getThanosUrl(ctx, organizationId)
	if err != nil {
		log.Error(ctx, err)
		return out, httpErrors.NewInternalServerError(err, "D_INVALID_PRIMARY_STACK", "")
	}
	address, port := helper.SplitAddress(ctx, thanosUrl)
	thanosClient, err := thanos.New(address, port, false, "")
	if err != nil {
		return out, errors.Wrap(err, "failed to create thanos client")
	}

	// Stack
	clusters, err := u.clusterRepo.FetchByOrganizationId(ctx, organizationId, uuid.Nil, nil)
	if err != nil {
		return out, err
	}

	filteredClusters := funk.Filter(clusters, func(x model.Cluster) bool {
		return x.Status != domain.ClusterStatus_DELETED
	})
	if filteredClusters != nil {
		out.Stack = fmt.Sprintf("%d 개", len(filteredClusters.([]model.Cluster)))
	} else {
		out.Stack = "0 개"
	}

	// CPU
	/*
		{"data":{"result":[{"metric":{"taco_cluster":"cmsai5k5l"},"value":[1683608185.65,"32"]},{"metric":{"taco_cluster":"crjfh12oc"},"value":[1683608185.65,"12"]}],"vector":""},"status":"success"}
	*/
	result, err := thanosClient.Get(ctx, "sum by (taco_cluster) (machine_cpu_cores)")
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
	result, err = thanosClient.Get(ctx, "sum by (taco_cluster) (machine_memory_bytes)")
	if err != nil {
		return out, err
	}
	memory := float64(0)
	for _, val := range result.Data.Result {
		memoryVal, err := strconv.Atoi(val.Value[1].(string))
		if err != nil {
			continue
		}
		if memoryVal > 0 {
			memory_ := float64(memoryVal) / float64(1024) / float64(1024) / float64(1024)
			memory = memory + memory_
		}
	}
	out.Memory = fmt.Sprintf("%v GiB", math.Round(memory))

	// Storage
	result, err = thanosClient.Get(ctx, "sum by (taco_cluster) (kubelet_volume_stats_capacity_bytes)")
	if err != nil {
		return out, err
	}
	storage := float64(0)
	for _, val := range result.Data.Result {
		storageVal, err := strconv.Atoi(val.Value[1].(string))
		if err != nil {
			continue
		}
		if storageVal > 0 {
			storage_ := float64(storageVal) / float64(1024) / float64(1024) / float64(1024)
			storage = storage + storage_
		}
	}
	out.Storage = fmt.Sprintf("%v GiB", math.Round(storage))

	return
}

func (u *DashboardUsecase) getChartFromPrometheus(ctx context.Context, organizationId string, chartType string, duration string, interval string, year string, month string) (res domain.DashboardChart, err error) {
	thanosUrl, err := u.getThanosUrl(ctx, organizationId)
	if err != nil {
		log.Error(ctx, err)
		return res, httpErrors.NewInternalServerError(err, "D_INVALID_PRIMARY_STACK", "")
	}
	address, port := helper.SplitAddress(ctx, thanosUrl)
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
	}

	query := ""

	switch chartType {
	case domain.ChartType_CPU.String():
		//query := "sum (avg(1-rate(node_cpu_seconds_total{mode=\"idle\"}[1h])) by (taco_cluster))"
		query = "avg by (taco_cluster) (1-irate(node_cpu_seconds_total{mode=\"idle\"}[" + interval + "]))"

	case domain.ChartType_MEMORY.String():
		query = "avg by (taco_cluster) (sum(node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes) by (taco_cluster) / sum(node_memory_MemTotal_bytes) by (taco_cluster))"

	case domain.ChartType_POD.String():
		query = "sum by (taco_cluster) (changes(kube_pod_container_status_restarts_total{namespace!=\"kube-system\"}[" + interval + "]))"

	case domain.ChartType_TRAFFIC.String():
		query = "avg by (taco_cluster) (irate(container_network_receive_bytes_total[" + interval + "]))"

	case domain.ChartType_POD_CALENDAR.String():
		// 입력받은 년,월 을 date 형식으로
		yearInt, _ := strconv.Atoi(year)
		monthInt, _ := strconv.Atoi(month)
		startDate := time.Date(yearInt, time.Month(monthInt), 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(yearInt, time.Month(monthInt+1), 1, 0, 0, 0, 0, time.UTC)

		if now.Year() < yearInt {
			return res, fmt.Errorf("Invalid year")
		} else if now.Year() == yearInt && int(now.Month()) < monthInt {
			return res, fmt.Errorf("Invalid month")
		}

		systemNotifications, err := u.systemNotificationRepo.FetchPodRestart(ctx, organizationId, startDate, endDate)
		if err != nil {
			return res, err
		}

		organization, err := u.organizationRepo.Get(ctx, organizationId)
		if err != nil {
			return res, err
		}

		log.Info(ctx, organization.CreatedAt.Format("2006-01-02"))

		podCounts := []domain.PodCount{}
		for day := rangeDate(startDate, endDate); ; {
			d := day()
			if d.IsZero() {
				break
			}

			baseDate := d.Format("2006-01-02")
			cntPodRestart := 0

			if baseDate <= now.Format("2006-01-02") && baseDate >= organization.CreatedAt.Format("2006-01-02") {
				for _, systemNotification := range systemNotifications {
					strDate := systemNotification.CreatedAt.Format("2006-01-02")

					if strDate == baseDate {
						cntPodRestart += 1
					}
				}
				pd := domain.PodCount{
					Day:   d.Day(),
					Value: int(cntPodRestart),
				}
				podCounts = append(podCounts, pd)
			}
		}
		chartData.XAxis = nil
		chartData.YAxis = nil
		chartData.PodCounts = podCounts

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

	result, err := thanosClient.FetchRange(ctx, query, int(now.Unix())-durationSec, int(now.Unix()), intervalSec)
	if err != nil {
		return res, err
	}

	// 모든 x축 부터 계산
	xAxisData := []string{}
	for _, val := range result.Data.Result {
		for _, vals := range val.Values {
			x := int(math.Round(vals.([]interface{})[0].(float64)))

			if !slices.Contains(xAxisData, strconv.Itoa(x)) {
				xAxisData = append(xAxisData, strconv.Itoa(x))
			}
		}
	}
	sort.Slice(xAxisData, func(i, j int) bool {
		a, _ := strconv.Atoi(xAxisData[i])
		b, _ := strconv.Atoi(xAxisData[j])
		return a < b
	})

	// cluster 별 y축 계산
	for _, val := range result.Data.Result {
		yAxisData := []string{}

		for _, xAxis := range xAxisData {
			percentage := false
			if chartType == domain.ChartType_CPU.String() || chartType == domain.ChartType_MEMORY.String() {
				percentage = true
			}
			yAxisData = append(yAxisData, u.getChartYValue(val.Values, xAxis, percentage))
		}

		clusterName, err := u.getClusterNameFromId(ctx, val.Metric.TacoCluster)
		if err != nil {
			clusterName = val.Metric.TacoCluster
		}

		chartData.Series = append(chartData.Series, domain.Unit{
			Name: clusterName,
			Data: yAxisData,
		})

	}
	chartData.XAxis = &domain.Axis{}
	chartData.XAxis.Data = xAxisData

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

func (u *DashboardUsecase) getThanosUrl(ctx context.Context, organizationId string) (out string, err error) {
	const prefix = "CACHE_KEY_THANOS_URL"
	value, found := u.cache.Get(prefix + organizationId)
	if found {
		return value.(string), nil
	}

	organization, err := u.organizationRepo.Get(ctx, organizationId)
	if err != nil {
		return out, errors.Wrap(err, "Failed to get organization")
	}

	//organization.PrimaryClusterId = "cmnl6zqmb"
	if organization.PrimaryClusterId == "" {
		return out, fmt.Errorf("Invalid primary clusterId")
	}

	clientset_admin, err := kubernetes.GetClientAdminCluster(ctx)
	if err != nil {
		return out, errors.Wrap(err, "Failed to get client set for user cluster")
	}

	// tks-endpoint-secret 이 있다면 그 secret 내의 endpoint 를 사용한다.
	secrets, err := clientset_admin.CoreV1().Secrets(organization.PrimaryClusterId).Get(context.TODO(), "tks-endpoint-secret", metav1.GetOptions{})
	if err != nil {
		log.Info(ctx, "cannot found tks-endpoint-secret. so use LoadBalancer...")

		clientset_user, err := kubernetes.GetClientFromClusterId(ctx, organization.PrimaryClusterId)
		if err != nil {
			return out, errors.Wrap(err, "Failed to get client set for user cluster")
		}

		service, err := clientset_user.CoreV1().Services("lma").Get(context.TODO(), "thanos-query-frontend", metav1.GetOptions{})
		if err != nil {
			service, err = clientset_user.CoreV1().Services("lma").Get(context.TODO(), "thanos-query", metav1.GetOptions{})
			if err != nil {
				return out, errors.Wrap(err, "Failed to get services.")
			}
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
			return out, nil
		}
	} else {
		out = "http://" + string(secrets.Data["thanos"])
		log.Info(ctx, "thanosUrl : ", out)
		u.cache.Set(prefix+organizationId, out, gcache.DefaultExpiration)
		return out, nil
	}

	return
}

func (u *DashboardUsecase) getChartYValue(values []interface{}, xData string, percentage bool) (out string) {
	for _, vals := range values {
		x := int(math.Round(vals.([]interface{})[0].(float64)))
		y, err := strconv.ParseFloat(vals.([]interface{})[1].(string), 32)
		if err != nil {
			return ""
		}
		if strconv.Itoa(x) == xData {
			if percentage {
				y = y * 100
			}
			return fmt.Sprintf("%f", y)
		}
	}
	return ""
}

func (u *DashboardUsecase) getStackMemoryDisk(result []thanos.MetricDataResult, clusterId string) (memory string, disk string) {
	// node_memory_MemFree_bytes|machine_memory_bytes|kubelet_volume_stats_used_bytes|kubelet_volume_stats_capacity_bytes

	free := 0
	machine := 0
	used := 0
	capacity := 0
	for _, val := range result {
		if val.Metric.TacoCluster == clusterId {
			if val.Metric.Name == "node_memory_MemFree_bytes" {
				free, _ = strconv.Atoi(val.Value[1].(string))
			} else if val.Metric.Name == "machine_memory_bytes" {
				machine, _ = strconv.Atoi(val.Value[1].(string))
			}

			if val.Metric.Name == "kubelet_volume_stats_used_bytes" {
				used, _ = strconv.Atoi(val.Value[1].(string))
			} else if val.Metric.Name == "kubelet_volume_stats_capacity_bytes" {
				capacity, _ = strconv.Atoi(val.Value[1].(string))
			}
		}
	}

	if machine > 0 {
		m := 1 - (float32(free) / float32(machine))
		memory = fmt.Sprintf("%0.2f", m*100)
	}

	if capacity > 0 {
		d := float32(used) / float32(capacity)
		disk = fmt.Sprintf("%0.2f", d*100)
	}

	return
}

func (u *DashboardUsecase) getStackCpu(result []thanos.MetricDataResult, clusterId string) (cpu string) {
	for _, val := range result {
		if val.Metric.TacoCluster == clusterId {
			if s, err := strconv.ParseFloat(val.Value[1].(string), 32); err == nil {
				cpu = fmt.Sprintf("%0.2f", s)
			}

			return cpu
		}
	}
	return
}

func (u *DashboardUsecase) getClusterNameFromId(ctx context.Context, clusterId string) (clusterName string, err error) {
	const prefix = "CACHE_KEY_CLUSTER_NAME_FROM_ID"
	value, found := u.cache.Get(prefix + clusterId)
	if found {
		return value.(string), nil
	}

	cluster, err := u.clusterRepo.Get(ctx, domain.ClusterId(clusterId))
	if err != nil {
		return clusterName, errors.Wrap(err, "Failed to get cluster")
	}
	clusterName = cluster.Name

	u.cache.Set(prefix+clusterId, clusterName, gcache.DefaultExpiration)
	return
}

func rangeDate(start, end time.Time) func() time.Time {
	y, m, d := start.Date()
	start = time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	y, m, d = end.Date()
	end = time.Date(y, m, d, 0, 0, 0, 0, time.UTC)

	return func() time.Time {
		if start.After(end) {
			return time.Time{}
		}
		date := start
		start = start.AddDate(0, 0, 1)
		return date
	}
}
