package usecase

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/helper"
	"github.com/openinfradev/tks-api/internal/model"
	policytemplate "github.com/openinfradev/tks-api/internal/policy-template"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/internal/serializer"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/kubernetes"
	"github.com/openinfradev/tks-api/pkg/log"
	thanos "github.com/openinfradev/tks-api/pkg/thanos-client"
	gcache "github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/thoas/go-funk"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/strings/slices"
)

type IDashboardUsecase interface {
	CreateDashboard(ctx context.Context, dashboard *model.Dashboard) (string, error)
	GetDashboard(ctx context.Context, organizationId string, userId string, dashboardKey string) (*model.Dashboard, error)
	UpdateDashboard(ctx context.Context, dashboard *model.Dashboard) error
	GetCharts(ctx context.Context, organizationId string, chartType domain.ChartType, duration string, interval string, year string, month string) (res []domain.DashboardChart, err error)
	GetStacks(ctx context.Context, organizationId string) (out []domain.DashboardStack, err error)
	GetResources(ctx context.Context, organizationId string) (out domain.DashboardResource, err error)
	GetPolicyUpdate(ctx context.Context, policyTemplates []policytemplate.TKSPolicyTemplate, policies []policytemplate.TKSPolicy) (domain.DashboardPolicyUpdate, error)
	GetPolicyEnforcement(ctx context.Context, organizationId string, primaryClusterId string) (*domain.BarChartData, error)
	GetPolicyViolation(ctx context.Context, organizationId string, duration string, interval string) (*domain.BarChartData, error)
	GetPolicyViolationLog(ctx context.Context, organizationId string) (*domain.GetDashboardPolicyViolationLogResponse, error)
	GetWorkload(ctx context.Context, organizationId string) (*domain.GetDashboardWorkloadResponse, error)
	GetPolicyViolationTop5(ctx context.Context, organizationId string, duration string, interval string) (*domain.BarChartData, error)
	GetThanosClient(ctx context.Context, organizationId string) (thanos.ThanosClient, error)
}

type DashboardUsecase struct {
	dashboardRepo          repository.IDashboardRepository
	organizationRepo       repository.IOrganizationRepository
	clusterRepo            repository.IClusterRepository
	appGroupRepo           repository.IAppGroupRepository
	systemNotificationRepo repository.ISystemNotificationRepository
	policyTemplateRepo     repository.IPolicyTemplateRepository
	policyRepo             repository.IPolicyRepository
	cache                  *gcache.Cache
}

func NewDashboardUsecase(r repository.Repository, cache *gcache.Cache) IDashboardUsecase {
	return &DashboardUsecase{
		dashboardRepo:          r.Dashboard,
		organizationRepo:       r.Organization,
		clusterRepo:            r.Cluster,
		appGroupRepo:           r.AppGroup,
		systemNotificationRepo: r.SystemNotification,
		policyTemplateRepo:     r.PolicyTemplate,
		policyRepo:             r.Policy,
		cache:                  cache,
	}
}

func (u *DashboardUsecase) CreateDashboard(ctx context.Context, dashboard *model.Dashboard) (string, error) {
	dashboardId, err := u.dashboardRepo.CreateDashboard(ctx, dashboard)
	if err != nil {
		return "", errors.Wrap(err, "Failed to create dashboard.")
	}

	return dashboardId, nil
}

func (u *DashboardUsecase) GetDashboard(ctx context.Context, organizationId string, userId string, dashboardKey string) (*model.Dashboard, error) {
	dashboard, err := u.dashboardRepo.GetDashboardByUserId(ctx, organizationId, userId, dashboardKey)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get dashboard.")
	}
	return dashboard, err
}

func (u *DashboardUsecase) UpdateDashboard(ctx context.Context, dashboard *model.Dashboard) error {
	if err := u.dashboardRepo.UpdateDashboard(ctx, dashboard); err != nil {
		return errors.Wrap(err, "Failed to update dashboard")
	}
	return nil
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
		log.Error(ctx, err)
		return out, err
	}

	filteredClusters := funk.Filter(clusters, func(x model.Cluster) bool {
		return x.Status == domain.ClusterStatus_RUNNING
	})

	var normal, abnormal int
	if filteredClusters != nil {
		for _, cluster := range filteredClusters.([]model.Cluster) {
			clientSet, err := kubernetes.GetClientFromClusterId(ctx, cluster.ID.String())
			if err != nil {
				return out, errors.Wrap(err, "Failed to get client set for user cluster")
			}
			// get cluster info
			clusterInfo, err := clientSet.CoreV1().Services("kube-system").List(context.TODO(), metav1.ListOptions{LabelSelector: "kubernetes.io/cluster-service"})
			if err != nil {
				abnormal++
				log.Debugf(ctx, "Failed to get cluster info: %v\n", err)
				continue
			}
			if clusterInfo != nil && len(clusterInfo.Items) > 0 {
				if clusterInfo.Items[0].ObjectMeta.Labels["kubernetes.io/cluster-service"] == "true" {
					normal++
				} else {
					abnormal++
				}
			}
		}
	}
	out.Stack.Normal = strconv.Itoa(normal)
	out.Stack.Abnormal = strconv.Itoa(abnormal)

	// CPU
	/*
		{"data":{"result":[{"metric":{"taco_cluster":"cmsai5k5l"},"value":[1683608185.65,"32"]},{"metric":{"taco_cluster":"crjfh12oc"},"value":[1683608185.65,"12"]}],"vector":""},"status":"success"}
	*/
	result, err := thanosClient.Get(ctx, "sum by (taco_cluster) (machine_cpu_cores)")
	if err != nil {
		log.Error(ctx, err)
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
	out.Cpu = strconv.Itoa(cpu)

	// Memory
	result, err = thanosClient.Get(ctx, "sum by (taco_cluster) (machine_memory_bytes)")
	if err != nil {
		log.Error(ctx, err)
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
	out.Memory = fmt.Sprintf("%v", math.Round(memory))

	// Storage
	result, err = thanosClient.Get(ctx, "sum by (taco_cluster) (kubelet_volume_stats_capacity_bytes)")
	if err != nil {
		log.Error(ctx, err)
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
	out.Storage = fmt.Sprintf("%v", math.Round(storage))

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

func (u *DashboardUsecase) GetPolicyUpdate(ctx context.Context, policyTemplates []policytemplate.TKSPolicyTemplate,
	policies []policytemplate.TKSPolicy) (domain.DashboardPolicyUpdate, error) {

	var outdatedTemplateIds []string
	for _, tpt := range policyTemplates {
		templateId := tpt.Labels[policytemplate.TemplateIDLabel]
		id, err := uuid.Parse(templateId)
		if err != nil {
			log.Errorf(ctx, "error is :%s(%T)", err.Error(), err)
			continue
		}
		version, err := u.policyTemplateRepo.GetLatestTemplateVersion(ctx, id)
		if err != nil {
			log.Errorf(ctx, "error is :%s(%T)", err.Error(), err)
			continue
		}

		if version != tpt.Spec.Version {
			outdatedTemplateIds = append(outdatedTemplateIds, templateId)
		}
	}

	outdatedTemplateCount := len(outdatedTemplateIds)
	outdatedPolicyCount := 0

	for _, policy := range policies {
		templateId := policy.Labels[policytemplate.TemplateIDLabel]

		if slices.Contains(outdatedTemplateIds, templateId) {
			outdatedPolicyCount++
		}
	}

	dpu := domain.DashboardPolicyUpdate{
		PolicyTemplate: outdatedTemplateCount,
		Policy:         outdatedPolicyCount,
	}

	return dpu, nil
}

func (u *DashboardUsecase) GetPolicyEnforcement(ctx context.Context, organizationId string, primaryClusterId string) (*domain.BarChartData, error) {
	type DashboardPolicyTemplate struct {
		ClusterId      string
		PolicyTemplate map[string]map[string]int
	}
	dashboardPolicyTemplates := make([]DashboardPolicyTemplate, 0)

	// get clusters from db
	dbClusters, err := u.clusterRepo.FetchByOrganizationId(ctx, organizationId, uuid.Nil, nil)
	if err != nil {
		return nil, err
	}
	filteredClusters := funk.Filter(dbClusters, func(x model.Cluster) bool {
		return x.Status == domain.ClusterStatus_RUNNING
	})
	if filteredClusters != nil {
		dbPolicyTemplates, err := u.policyTemplateRepo.GetPolicyTemplateByOrganizationIdOrTKS(ctx, organizationId)
		if err != nil {
			return nil, err
		}
		for _, cluster := range filteredClusters.([]model.Cluster) {
			policyTemplates := make(map[string]map[string]int)
			// get policytemplates by cluster
			// dbPolicyTemplates = {"K8sAllowedRepos": {"": 0}}
			for _, dpt := range dbPolicyTemplates {
				if _, ok := policyTemplates[dpt.TemplateName]; !ok {
					policyTemplates[dpt.TemplateName] = map[string]int{"": 0}
				}
			}
			dashboardPolicyTemplates = append(dashboardPolicyTemplates,
				DashboardPolicyTemplate{ClusterId: cluster.Name, PolicyTemplate: policyTemplates})
		}
	}

	// get clusters from cr
	clusters, err := policytemplate.GetTksClusterCRs(ctx, primaryClusterId)
	if err != nil {
		log.Error(ctx, "Failed to retrieve policytemplate list ", err)
		return nil, err
	}

	for _, cluster := range clusters {
		// policyTemplates = {"K8sAllowedRepos": {"members": 1}}
		policyTemplates := make(map[string]map[string]int)

		// If the cluster does not have a policytemplate, skip ahead
		// cluster.Status.Templates = {"K8sAllowedRepos": ["members"]}
		if cluster.Status.Templates == nil {
			continue
		}

		for templateName, policies := range cluster.Status.Templates {
			for _, policy := range policies {
				policyTemplates[templateName] = make(map[string]int)
				policyTemplates[templateName][policy] = 1
			}
		}

		dashboardPolicyTemplates = append(dashboardPolicyTemplates,
			DashboardPolicyTemplate{ClusterId: cluster.Name, PolicyTemplate: policyTemplates})
	}

	// fetch policies from db
	dbPolicies, err := u.policyRepo.Fetch(ctx, organizationId, nil)
	if err != nil {
		return nil, err
	}

	type TotalPolicyCount struct {
		PolicyName          string
		OptionalPolicyCount int
		RequiredPolicyCount int
	}

	// totalTemplate = {"template name": TotalPolicyCount}
	totalTemplate := make(map[string]*TotalPolicyCount)
	for _, dpt := range dashboardPolicyTemplates {
		for templateName, policies := range dpt.PolicyTemplate {
			if _, ok := totalTemplate[templateName]; !ok {
				totalTemplate[templateName] = &TotalPolicyCount{"", 0, 0}
			}
			// check if policy is required or optional
			for policy, count := range policies {
				for _, dbPolicy := range *dbPolicies {
					if policy == dbPolicy.PolicyResourceName {
						temp := totalTemplate[templateName]
						temp.PolicyName = policy
						if dbPolicy.Mandatory {
							temp.RequiredPolicyCount += count
						} else {
							temp.OptionalPolicyCount += count
						}
					}
				}
			}
		}
	}

	// desc sorting by value
	type ChartData struct {
		Name          string
		OptionalCount int
		RequiredCount int
	}
	chartData := make([]ChartData, 0)
	for key, val := range totalTemplate {
		data := ChartData{
			Name:          key,
			OptionalCount: val.OptionalPolicyCount,
			RequiredCount: val.RequiredPolicyCount,
		}
		chartData = append(chartData, data)
	}
	sort.Slice(chartData, func(i, j int) bool {
		return chartData[i].OptionalCount+chartData[i].RequiredCount >
			chartData[j].OptionalCount+chartData[j].RequiredCount
	})

	// X축
	var xAxis *domain.Axis
	xData := make([]string, 0)

	// Y축
	var series []domain.UnitNumber
	yOptionalData := make([]int, 0)
	yRequiredData := make([]int, 0)

	for _, v := range chartData {
		xData = append(xData, v.Name)
		yOptionalData = append(yOptionalData, v.OptionalCount)
		yRequiredData = append(yRequiredData, v.RequiredCount)
	}

	xAxis = &domain.Axis{
		Data: xData,
	}

	optionalUnit := domain.UnitNumber{
		Name: "선택",
		Data: yOptionalData,
	}
	series = append(series, optionalUnit)

	requiredUnit := domain.UnitNumber{
		Name: "필수",
		Data: yRequiredData,
	}
	series = append(series, requiredUnit)

	bcd := &domain.BarChartData{
		XAxis:  xAxis,
		Series: series,
	}

	return bcd, nil
}

func (u *DashboardUsecase) GetPolicyViolation(ctx context.Context, organizationId string, duration string, interval string) (*domain.BarChartData, error) {
	thanosClient, err := u.GetThanosClient(ctx, organizationId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create thanos client")
	}

	durationSec, intervalSec := getDurationAndIntervalSec(duration, interval)

	clusterIdStr, err := u.GetFlatClusterIds(ctx, organizationId)
	if err != nil {
		return nil, err
	}
	query := fmt.Sprintf("sum by(kind,name,violation_enforcement)(opa_scorecard_constraint_violations{taco_cluster=~\"%s\"})", clusterIdStr)

	now := time.Now()
	pm, err := thanosClient.FetchPolicyRange(ctx, query, int(now.Unix())-durationSec, int(now.Unix()), intervalSec)
	if err != nil {
		return nil, err
	}

	// totalViolation: {"K8sRequiredLabels": {"violation_enforcement": 2}}
	totalViolation := make(map[string]map[string]int)

	dbPolicyTemplates, err := u.policyTemplateRepo.GetPolicyTemplateByOrganizationIdOrTKS(ctx, organizationId)
	if err != nil {
		return nil, err
	}
	for _, dpt := range dbPolicyTemplates {
		if _, ok := totalViolation[dpt.TemplateName]; !ok {
			totalViolation[dpt.TemplateName] = map[string]int{"": 0}
		}
	}

	for _, res := range pm.Data.Result {
		policyTemplate := res.Metric.Kind
		if len(res.Metric.Violation) == 0 {
			continue
		}

		count := 0
		if res.Value != nil && len(res.Value) > 1 {
			count, err = strconv.Atoi(res.Value[1].(string))
			if err != nil {
				count = 0
			}
		}

		violation := res.Metric.Violation
		if val, ok := totalViolation[policyTemplate][violation]; !ok {
			totalViolation[policyTemplate] = make(map[string]int)
			totalViolation[policyTemplate][violation] = count
		} else {
			totalViolation[policyTemplate][violation] = val + count
		}
	}

	// desc sorting by value
	type ChartData struct {
		Name        string
		DenyCount   int
		WarnCount   int
		DryRunCount int
	}
	chartData := make([]ChartData, 0)
	for pt, violations := range totalViolation {
		data := ChartData{}
		data.Name = pt
		if val, ok := violations["deny"]; ok {
			data.DenyCount = val
		}
		if val, ok := violations["warn"]; ok {
			data.WarnCount = val
		}
		if val, ok := violations["dryrun"]; ok {
			data.DryRunCount = val
		}
		chartData = append(chartData, data)
	}
	sort.Slice(chartData, func(i, j int) bool {
		return chartData[i].DenyCount+chartData[i].WarnCount+chartData[i].DryRunCount >
			chartData[j].DenyCount+chartData[j].WarnCount+chartData[j].DryRunCount
	})

	// X축
	var xAxis *domain.Axis
	xData := make([]string, 0)

	// Y축
	var series []domain.UnitNumber
	yDenyData := make([]int, 0)
	yWarnData := make([]int, 0)
	yDryrunData := make([]int, 0)

	for _, v := range chartData {
		xData = append(xData, v.Name)
		yDenyData = append(yDenyData, v.DenyCount)
		yWarnData = append(yWarnData, v.WarnCount)
		yDryrunData = append(yDryrunData, v.DryRunCount)
	}

	xAxis = &domain.Axis{
		Data: xData,
	}

	denyUnit := domain.UnitNumber{
		Name: "거부",
		Data: yDenyData,
	}
	series = append(series, denyUnit)

	warnUnit := domain.UnitNumber{
		Name: "경고",
		Data: yWarnData,
	}
	series = append(series, warnUnit)

	dryrunUnit := domain.UnitNumber{
		Name: "감사",
		Data: yDryrunData,
	}
	series = append(series, dryrunUnit)

	bcd := &domain.BarChartData{
		XAxis:  xAxis,
		Series: series,
	}

	return bcd, nil
}

func (u *DashboardUsecase) GetPolicyViolationLog(ctx context.Context, organizationId string) (*domain.GetDashboardPolicyViolationLogResponse, error) {
	// TODO Implement me
	return nil, nil
}

func (u *DashboardUsecase) GetWorkload(ctx context.Context, organizationId string) (*domain.GetDashboardWorkloadResponse, error) {
	thanosClient, err := u.GetThanosClient(ctx, organizationId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create thanos client")
	}

	clusterIdStr, err := u.GetFlatClusterIds(ctx, organizationId)
	if err != nil {
		return nil, err
	}

	dwr := &domain.GetDashboardWorkloadResponse{Title: "자원별 Pod 배포 현황"}

	// Deployment pod count
	count := 0
	query := fmt.Sprintf("sum (kube_deployment_status_replicas_available{taco_cluster=~'%s'} )", clusterIdStr)
	wm, err := thanosClient.GetWorkload(ctx, query)
	if err != nil {
		return nil, err
	}
	if len(wm.Data.Result) > 0 && len(wm.Data.Result[0].Value) > 1 {
		count, _ = strconv.Atoi(wm.Data.Result[0].Value[1].(string))
	}

	dwr.Data = append(dwr.Data, domain.WorkloadData{Name: "Deployments", Value: count})

	// StatefulSet pod count
	count = 0
	query = fmt.Sprintf("sum (kube_statefulset_status_replicas_available{taco_cluster=~'%s'} )", clusterIdStr)
	wm, err = thanosClient.GetWorkload(ctx, query)
	if err != nil {
		return nil, err
	}
	if len(wm.Data.Result) > 0 && len(wm.Data.Result[0].Value) > 1 {
		count, _ = strconv.Atoi(wm.Data.Result[0].Value[1].(string))
	}
	dwr.Data = append(dwr.Data, domain.WorkloadData{Name: "StatefulSets", Value: count})

	// DaemonSet pod count
	count = 0
	query = fmt.Sprintf("sum (kube_daemonset_status_number_available{taco_cluster=~'%s'} )", clusterIdStr)
	wm, err = thanosClient.GetWorkload(ctx, query)
	if err != nil {
		return nil, err
	}
	if len(wm.Data.Result) > 0 && len(wm.Data.Result[0].Value) > 1 {
		count, _ = strconv.Atoi(wm.Data.Result[0].Value[1].(string))
	}
	dwr.Data = append(dwr.Data, domain.WorkloadData{Name: "DaemonSets", Value: count})

	// CronJob pod count
	count = 0
	query = fmt.Sprintf("sum (kube_cronjob_status_active{taco_cluster=~'%s'} )", clusterIdStr)
	wm, err = thanosClient.GetWorkload(ctx, query)
	if err != nil {
		return nil, err
	}
	if len(wm.Data.Result) > 0 && len(wm.Data.Result[0].Value) > 1 {
		count, _ = strconv.Atoi(wm.Data.Result[0].Value[1].(string))
	}
	dwr.Data = append(dwr.Data, domain.WorkloadData{Name: "CronJobs", Value: count})

	// Job pod count
	count = 0
	query = fmt.Sprintf("sum (kube_job_status_active{taco_cluster=~'%s'} )", clusterIdStr)
	wm, err = thanosClient.GetWorkload(ctx, query)
	if err != nil {
		return nil, err
	}
	if len(wm.Data.Result) > 0 && len(wm.Data.Result[0].Value) > 1 {
		count, _ = strconv.Atoi(wm.Data.Result[0].Value[1].(string))
	}
	dwr.Data = append(dwr.Data, domain.WorkloadData{Name: "Jobs", Value: count})

	return dwr, nil
}

func (u *DashboardUsecase) GetPolicyViolationTop5(ctx context.Context, organizationId string, duration string, interval string) (*domain.BarChartData, error) {
	thanosClient, err := u.GetThanosClient(ctx, organizationId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create thanos client")
	}

	durationSec, intervalSec := getDurationAndIntervalSec(duration, interval)

	clusterIdStr, err := u.GetFlatClusterIds(ctx, organizationId)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	query := fmt.Sprintf("topk (5, sum by (kind) (opa_scorecard_constraint_violations{taco_cluster=~'%s'}))", clusterIdStr)
	ptm, err := thanosClient.FetchPolicyTemplateRange(ctx, query, int(now.Unix())-durationSec, int(now.Unix()), intervalSec)
	if err != nil {
		return nil, err
	}

	templateNames := make([]string, 0)
	for _, result := range ptm.Data.Result {
		templateNames = append(templateNames, result.Metric.Kind)
	}

	// desc sorting by value
	type ChartData struct {
		Name        string
		DenyCount   int
		WarnCount   int
		DryRunCount int
	}
	chartData := make([]ChartData, 0)
	for _, templateName := range templateNames {
		//xData = append(xData, templateName)
		data := ChartData{}
		data.Name = templateName

		query = fmt.Sprintf("sum by (violation_enforcement) "+
			"(opa_scorecard_constraint_violations{taco_cluster='%s', kind='%s'})", clusterIdStr, templateName)
		pvcm, err := thanosClient.FetchPolicyViolationCountRange(ctx, query, int(now.Unix())-durationSec, int(now.Unix()), intervalSec)
		if err != nil {
			return nil, err
		}

		denyCount := 0
		warnCount := 0
		dryrunCount := 0
		for _, result := range pvcm.Data.Result {
			if result.Value == nil || len(result.Value) <= 1 {
				continue
			}

			switch policy := result.Metric.ViolationEnforcement; policy {
			case "":
				denyCount, _ = strconv.Atoi(result.Value[1].(string))
			case "warn":
				warnCount, _ = strconv.Atoi(result.Value[1].(string))
			case "dryrun":
				dryrunCount, _ = strconv.Atoi(result.Value[1].(string))
			}
		}
		data.DenyCount = denyCount
		data.WarnCount = warnCount
		data.DryRunCount = dryrunCount
		chartData = append(chartData, data)
	}
	sort.Slice(chartData, func(i, j int) bool {
		return chartData[i].DenyCount+chartData[i].WarnCount+chartData[i].DryRunCount >
			chartData[j].DenyCount+chartData[j].WarnCount+chartData[j].DryRunCount
	})

	// X축
	var xAxis *domain.Axis
	xData := make([]string, 0)

	// Y축
	var series []domain.UnitNumber
	yDenyData := make([]int, 0)
	yWarnData := make([]int, 0)
	yDryrunData := make([]int, 0)

	for _, v := range chartData {
		xData = append(xData, v.Name)
		yDenyData = append(yDenyData, v.DenyCount)
		yWarnData = append(yWarnData, v.WarnCount)
		yDryrunData = append(yDryrunData, v.DryRunCount)
	}

	xAxis = &domain.Axis{
		Data: xData,
	}

	denyUnit := domain.UnitNumber{
		Name: "거부",
		Data: yDenyData,
	}
	series = append(series, denyUnit)

	warnUnit := domain.UnitNumber{
		Name: "경고",
		Data: yWarnData,
	}
	series = append(series, warnUnit)

	dryrunUnit := domain.UnitNumber{
		Name: "감사",
		Data: yDryrunData,
	}
	series = append(series, dryrunUnit)

	bcd := &domain.BarChartData{
		XAxis:  xAxis,
		Series: series,
	}

	return bcd, nil
}

func (u *DashboardUsecase) GetThanosClient(ctx context.Context, organizationId string) (thanos.ThanosClient, error) {
	thanosUrl, err := u.getThanosUrl(ctx, organizationId)
	if err != nil {
		log.Error(ctx, err)
		return nil, httpErrors.NewInternalServerError(err, "D_INVALID_PRIMARY_STACK", "")
	}
	address, port := helper.SplitAddress(ctx, thanosUrl)

	// [TEST]
	//address = "http://a93c60de70c794ef39b495976588c989-d7cd29ca75def693.elb.ap-northeast-2.amazonaws.com"
	//port = 9090

	client, err := thanos.New(address, port, false, "")
	if err != nil {
		return nil, errors.Wrap(err, "failed to create thanos client")
	}
	return client, nil
}

func (u *DashboardUsecase) GetFlatClusterIds(ctx context.Context, organizationId string) (string, error) {
	clusters, err := u.clusterRepo.FetchByOrganizationId(ctx, organizationId, uuid.Nil, nil)
	if err != nil {
		log.Error(ctx, err)
		return "", err
	}
	var clusterIds []string
	for _, cluster := range clusters {
		clusterIds = append(clusterIds, cluster.ID.String())
	}
	clusterIdStr := strings.Join(clusterIds, "|")
	return clusterIdStr, nil
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

func getDurationAndIntervalSec(duration string, interval string) (int, int) {
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

	return durationSec, intervalSec
}
