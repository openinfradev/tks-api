package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/servicequotas"
	"github.com/openinfradev/tks-api/internal/helper"
	"github.com/openinfradev/tks-api/internal/kubernetes"
	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/internal/serializer"
	argowf "github.com/openinfradev/tks-api/pkg/argo-client"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	byoh "github.com/vmware-tanzu/cluster-api-provider-bringyourownhost/apis/infrastructure/v1beta1"
	"gorm.io/gorm"
)

type IStackUsecase interface {
	Get(ctx context.Context, stackId domain.StackId) (domain.Stack, error)
	GetByName(ctx context.Context, organizationId string, name string) (domain.Stack, error)
	Fetch(ctx context.Context, organizationId string, pg *pagination.Pagination) ([]domain.Stack, error)
	Create(ctx context.Context, dto domain.Stack) (stackId domain.StackId, err error)
	Install(ctx context.Context, stackId domain.StackId) (err error)
	Update(ctx context.Context, dto domain.Stack) error
	Delete(ctx context.Context, dto domain.Stack) error
	GetKubeConfig(ctx context.Context, stackId domain.StackId) (kubeConfig string, err error)
	GetStepStatus(ctx context.Context, stackId domain.StackId) (out []domain.StackStepStatus, stackStatus string, err error)
	SetFavorite(ctx context.Context, stackId domain.StackId) error
	DeleteFavorite(ctx context.Context, stackId domain.StackId) error
	GetNodes(ctx context.Context, stackId domain.StackId) (out domain.Stack, err error)
}

type StackUsecase struct {
	clusterRepo       repository.IClusterRepository
	appGroupRepo      repository.IAppGroupRepository
	cloudAccountRepo  repository.ICloudAccountRepository
	organizationRepo  repository.IOrganizationRepository
	stackTemplateRepo repository.IStackTemplateRepository
	appServeAppRepo   repository.IAppServeAppRepository
	argo              argowf.ArgoClient
	dashbordUsecase   IDashboardUsecase
}

func NewStackUsecase(r repository.Repository, argoClient argowf.ArgoClient, dashbordUsecase IDashboardUsecase) IStackUsecase {
	return &StackUsecase{
		clusterRepo:       r.Cluster,
		appGroupRepo:      r.AppGroup,
		cloudAccountRepo:  r.CloudAccount,
		organizationRepo:  r.Organization,
		stackTemplateRepo: r.StackTemplate,
		appServeAppRepo:   r.AppServeApp,
		argo:              argoClient,
		dashbordUsecase:   dashbordUsecase,
	}
}

func (u *StackUsecase) Create(ctx context.Context, dto domain.Stack) (stackId domain.StackId, err error) {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return "", httpErrors.NewUnauthorizedError(fmt.Errorf("Invalid token"), "A_INVALID_TOKEN", "")
	}

	_, err = u.GetByName(ctx, dto.OrganizationId, dto.Name)
	if err == nil {
		return "", httpErrors.NewBadRequestError(httpErrors.DuplicateResource, "S_CREATE_ALREADY_EXISTED_NAME", "")
	}

	_, err = u.stackTemplateRepo.Get(dto.StackTemplateId)
	if err != nil {
		return "", httpErrors.NewInternalServerError(errors.Wrap(err, "Invalid stackTemplateId"), "S_INVALID_STACK_TEMPLATE", "")
	}

	clusters, err := u.clusterRepo.FetchByOrganizationId(dto.OrganizationId, nil)
	if err != nil {
		return "", httpErrors.NewInternalServerError(errors.Wrap(err, "Failed to get clusters"), "S_FAILED_GET_CLUSTERS", "")
	}
	isPrimary := false
	if len(clusters) == 0 {
		isPrimary = true
	}
	log.DebugWithContext(ctx, "isPrimary ", isPrimary)

	if dto.CloudService == domain.CloudService_BYOH {
		if dto.ClusterEndpoint == "" {
			return "", httpErrors.NewBadRequestError(fmt.Errorf("Invalid userClusterDomain"), "S_INVALID_ADMINCLUSTER_URL", "")
		}
		if dto.ClusterId == "" {
			return "", httpErrors.NewBadRequestError(fmt.Errorf("Invalid clusterId"), "S_INVALID_CLUSTER_ID", "")
		}
	} else {
		if _, err = u.cloudAccountRepo.Get(dto.CloudAccountId); err != nil {
			return "", httpErrors.NewInternalServerError(errors.Wrap(err, "Invalid cloudAccountId"), "S_INVALID_CLOUD_ACCOUNT", "")
		}
	}

	// Make stack nodes
	var stackConf domain.StackConfResponse
	if err = domain.Map(dto.Conf, &stackConf); err != nil {
		log.InfoWithContext(ctx, err)
	}

	workflow := "tks-stack-create"
	workflowId, err := u.argo.SumbitWorkflowFromWftpl(workflow, argowf.SubmitOptions{
		Parameters: []string{
			fmt.Sprintf("tks_api_url=%s", viper.GetString("external-address")),
			"cluster_name=" + dto.Name,
			"description=" + dto.Description,
			"organization_id=" + dto.OrganizationId,
			"cloud_account_id=" + dto.CloudAccountId.String(),
			"stack_template_id=" + dto.StackTemplateId.String(),
			"creator=" + user.GetUserId().String(),
			"base_repo_branch=" + viper.GetString("revision"),
			"infra_conf=" + strings.Replace(helper.ModelToJson(stackConf), "\"", "\\\"", -1),
			"cloud_service=" + dto.CloudService,
		},
	})
	if err != nil {
		log.ErrorWithContext(ctx, err)
		return "", httpErrors.NewInternalServerError(err, "S_FAILED_TO_CALL_WORKFLOW", "")
	}
	log.DebugWithContext(ctx, "Submitted workflow: ", workflowId)

	// wait & get clusterId ( max 1min 	)
	dto.ID = domain.StackId("")
	for i := 0; i < 60; i++ {
		time.Sleep(time.Second * 5)
		workflow, err := u.argo.GetWorkflow("argo", workflowId)
		if err != nil {
			return "", err
		}

		log.DebugWithContext(ctx, "workflow ", workflow)
		if workflow.Status.Phase != "" && workflow.Status.Phase != "Running" {
			return "", fmt.Errorf("Invalid workflow status [%s]", workflow.Status.Phase)
		}

		cluster, err := u.clusterRepo.GetByName(dto.OrganizationId, dto.Name)
		if err != nil {
			continue
		}
		if cluster.Name == dto.Name {
			dto.ID = domain.StackId(cluster.ID)
			break
		}
	}

	return dto.ID, nil
}

func (u *StackUsecase) Install(ctx context.Context, stackId domain.StackId) (err error) {
	cluster, err := u.Get(ctx, stackId)
	if err != nil {
		return httpErrors.NewBadRequestError(fmt.Errorf("Invalid stackId"), "S_INVALID_STACK_ID", "")
	}

	_, err = u.stackTemplateRepo.Get(cluster.StackTemplateId)
	if err != nil {
		return httpErrors.NewInternalServerError(errors.Wrap(err, "Invalid stackTemplateId"), "S_INVALID_STACK_TEMPLATE", "")
	}

	clusters, err := u.clusterRepo.FetchByOrganizationId(cluster.OrganizationId, nil)
	if err != nil {
		return httpErrors.NewInternalServerError(errors.Wrap(err, "Failed to get clusters"), "S_FAILED_GET_CLUSTERS", "")
	}
	isPrimary := false
	if len(clusters) == 0 {
		isPrimary = true
	}
	log.DebugWithContext(ctx, "isPrimary ", isPrimary)

	if cluster.CloudService != domain.CloudService_BYOH {
		return httpErrors.NewBadRequestError(fmt.Errorf("Invalid cloud service"), "S_INVALID_CLOUD_SERVICE", "")
	}

	// Make stack nodes
	var stackConf domain.StackConfResponse
	if err = domain.Map(cluster.Conf, &stackConf); err != nil {
		log.InfoWithContext(ctx, err)
	}

	workflow := "tks-stack-create"
	workflowId, err := u.argo.SumbitWorkflowFromWftpl(workflow, argowf.SubmitOptions{
		Parameters: []string{
			fmt.Sprintf("tks_api_url=%s", viper.GetString("external-address")),
			"cluster_name=" + cluster.Name,
			"description=" + cluster.Description,
			"organization_id=" + cluster.OrganizationId,
			"cloud_account_id=NULL",
			"stack_template_id=" + cluster.StackTemplateId.String(),
			"creator=" + (*cluster.CreatorId).String(),
			"base_repo_branch=" + viper.GetString("revision"),
			"infra_conf=" + strings.Replace(helper.ModelToJson(stackConf), "\"", "\\\"", -1),
			"cloud_service=" + cluster.CloudService,
		},
	})
	if err != nil {
		log.ErrorWithContext(ctx, err)
		return httpErrors.NewInternalServerError(err, "S_FAILED_TO_CALL_WORKFLOW", "")
	}
	log.DebugWithContext(ctx, "Submitted workflow: ", workflowId)

	return nil
}

func (u *StackUsecase) Get(ctx context.Context, stackId domain.StackId) (out domain.Stack, err error) {
	cluster, err := u.clusterRepo.Get(domain.ClusterId(stackId))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return out, httpErrors.NewNotFoundError(err, "S_FAILED_FETCH_CLUSTER", "")
		}
		return out, err
	}

	organization, err := u.organizationRepo.Get(cluster.OrganizationId)
	if err != nil {
		return out, httpErrors.NewInternalServerError(errors.Wrap(err, fmt.Sprintf("Failed to get organization for clusterId %s", domain.ClusterId(stackId))), "S_FAILED_FETCH_ORGANIZATION", "")
	}

	appGroups, err := u.appGroupRepo.Fetch(domain.ClusterId(stackId), nil)
	if err != nil {
		return out, err
	}

	out = reflectClusterToStack(cluster, appGroups)
	if organization.PrimaryClusterId == cluster.ID.String() {
		out.PrimaryCluster = true
	}

	appGroupsInPrimaryCluster, err := u.appGroupRepo.Fetch(domain.ClusterId(organization.PrimaryClusterId), nil)
	if err != nil {
		return out, err
	}

	for _, appGroup := range appGroupsInPrimaryCluster {
		if appGroup.AppGroupType == domain.AppGroupType_LMA {
			applications, err := u.appGroupRepo.GetApplications(appGroup.ID, domain.ApplicationType_GRAFANA)
			if err != nil {
				return out, err
			}
			if len(applications) > 0 {
				out.GrafanaUrl = applications[0].Endpoint + "/d/tks_cluster_dashboard/tks-kubernetes-view-cluster-global?var-taco_cluster=" + cluster.ID.String()
			}
		}
	}

	return
}

func (u *StackUsecase) GetByName(ctx context.Context, organizationId string, name string) (out domain.Stack, err error) {
	cluster, err := u.clusterRepo.GetByName(organizationId, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return out, httpErrors.NewNotFoundError(err, "S_FAILED_FETCH_CLUSTER", "")
		}
		return out, err
	}

	appGroups, err := u.appGroupRepo.Fetch(cluster.ID, nil)
	if err != nil {
		return out, err
	}

	out = reflectClusterToStack(cluster, appGroups)
	return
}

func (u *StackUsecase) Fetch(ctx context.Context, organizationId string, pg *pagination.Pagination) (out []domain.Stack, err error) {
	organization, err := u.organizationRepo.Get(organizationId)
	if err != nil {
		return out, httpErrors.NewInternalServerError(errors.Wrap(err, fmt.Sprintf("Failed to get organization for clusterId %s", organizationId)), "S_FAILED_FETCH_ORGANIZATION", "")
	}

	clusters, err := u.clusterRepo.FetchByOrganizationId(organizationId, pg)
	if err != nil {
		return out, err
	}

	stackResources, _ := u.dashbordUsecase.GetStacks(ctx, organizationId)

	for _, cluster := range clusters {
		appGroups, err := u.appGroupRepo.Fetch(cluster.ID, nil)
		if err != nil {
			return nil, err
		}

		outStack := reflectClusterToStack(cluster, appGroups)
		if organization.PrimaryClusterId == cluster.ID.String() {
			outStack.PrimaryCluster = true
		}

		for _, appGroup := range appGroups {
			if appGroup.AppGroupType == domain.AppGroupType_LMA {
				applications, err := u.appGroupRepo.GetApplications(appGroup.ID, domain.ApplicationType_GRAFANA)
				if err != nil {
					return nil, err
				}
				if len(applications) > 0 {
					outStack.GrafanaUrl = applications[0].Endpoint
				}
			}
		}

		for _, resource := range stackResources {
			if resource.ID == domain.StackId(cluster.ID) {
				if err := serializer.Map(resource, &outStack.Resource); err != nil {
					log.Error(err)
				}
			}
		}

		out = append(out, outStack)
	}

	return
}

func (u *StackUsecase) Update(ctx context.Context, dto domain.Stack) (err error) {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return httpErrors.NewBadRequestError(fmt.Errorf("Invalid token"), "", "")
	}

	_, err = u.clusterRepo.Get(domain.ClusterId(dto.ID))
	if err != nil {
		return httpErrors.NewNotFoundError(err, "S_FAILED_FETCH_CLUSTER", "")
	}

	updatorId := user.GetUserId()
	dtoCluster := domain.Cluster{
		ID:          domain.ClusterId(dto.ID),
		Description: dto.Description,
		UpdatorId:   &updatorId,
	}

	err = u.clusterRepo.Update(dtoCluster)
	if err != nil {
		return err
	}

	return nil
}

func (u *StackUsecase) Delete(ctx context.Context, dto domain.Stack) (err error) {
	cluster, err := u.clusterRepo.Get(domain.ClusterId(dto.ID))
	if err != nil {
		return httpErrors.NewBadRequestError(errors.Wrap(err, "Failed to get cluster"), "S_FAILED_FETCH_CLUSTER", "")
	}

	// 지우려고 하는 stack 이 primary cluster 라면, organization 내에 cluster 가 자기 자신만 남아있을 경우이다.
	organizations, err := u.organizationRepo.Fetch(nil)
	if err != nil {
		return errors.Wrap(err, "Failed to get organizations")
	}

	for _, organization := range *organizations {
		if organization.PrimaryClusterId == cluster.ID.String() {

			clusters, err := u.clusterRepo.FetchByOrganizationId(organization.ID, nil)
			if err != nil {
				return errors.Wrap(err, "Failed to get organizations")
			}

			for _, cl := range clusters {
				if cl.ID != cluster.ID && (cl.Status == domain.ClusterStatus_RUNNING ||
					cl.Status == domain.ClusterStatus_INSTALLING ||
					cl.Status == domain.ClusterStatus_DELETING) {
					return httpErrors.NewBadRequestError(fmt.Errorf("Failed to delete 'Primary' cluster. The clusters remain in organization"), "S_REMAIN_CLUSTER_FOR_DELETION", "")
				}
			}
			break
		}
	}
	appGroups, err := u.appGroupRepo.Fetch(domain.ClusterId(dto.ID), nil)
	if err != nil {
		return errors.Wrap(err, "Failed to get appGroups")
	}
	if len(appGroups) > 0 {
		for _, appGroup := range appGroups {
			if appGroup.Status != domain.AppGroupStatus_RUNNING {
				return fmt.Errorf("Appgroup status is not 'RUNNING'. status [%s]", appGroup.Status.String())
			}
		}
	}

	appsCnt, err := u.appServeAppRepo.GetNumOfAppsOnStack(dto.OrganizationId, dto.ID.String())
	if err != nil {
		return errors.Wrap(err, "Failed to get numOfAppsOnStack")
	}
	if appsCnt > 0 {
		return httpErrors.NewBadRequestError(fmt.Errorf("existed appServeApps in %s", dto.OrganizationId), "S_FAILED_DELETE_EXISTED_ASA", "")
	}

	// [TODO] BYOH 삭제는 어떻게 처리하는게 좋은가?

	workflow := "tks-stack-delete"
	workflowId, err := u.argo.SumbitWorkflowFromWftpl(workflow, argowf.SubmitOptions{
		Parameters: []string{
			fmt.Sprintf("tks_api_url=%s", viper.GetString("external-address")),
			"organization_id=" + dto.OrganizationId,
			"cluster_id=" + dto.ID.String(),
			"cloud_account_id=" + cluster.CloudAccount.ID.String(),
			"stack_template_id=" + cluster.StackTemplate.ID.String(),
		},
	})
	if err != nil {
		log.ErrorWithContext(ctx, err)
		return err
	}
	log.DebugWithContext(ctx, "Submitted workflow: ", workflowId)

	// Remove Cluster & AppGroup status description
	if err := u.appGroupRepo.InitWorkflowDescription(cluster.ID); err != nil {
		log.ErrorWithContext(ctx, err)
	}
	if err := u.clusterRepo.InitWorkflowDescription(cluster.ID); err != nil {
		log.ErrorWithContext(ctx, err)
	}

	// wait & get clusterId ( max 1min 	)
	for i := 0; i < 60; i++ {
		time.Sleep(time.Second * 2)
		workflow, err := u.argo.GetWorkflow("argo", workflowId)
		if err != nil {
			return err
		}

		if workflow.Status.Phase != "" && workflow.Status.Phase != "Running" {
			return fmt.Errorf("Invalid workflow status")
		}

		if workflow.Status.Progress == "1/2" { // start creating cluster
			time.Sleep(time.Second * 5) // Buffer
			break
		}
	}

	return nil
}

func (u *StackUsecase) GetKubeConfig(ctx context.Context, stackId domain.StackId) (kubeConfig string, err error) {
	kubeconfig, err := kubernetes.GetKubeConfig(stackId.String())
	//kubeconfig, err := kubernetes.GetKubeConfig("cmsai5k5l")
	if err != nil {
		return "", err
	}

	return string(kubeconfig[:]), nil
}

// [TODO] need more pretty...
func (u *StackUsecase) GetStepStatus(ctx context.Context, stackId domain.StackId) (out []domain.StackStepStatus, stackStatus string, err error) {
	cluster, err := u.clusterRepo.Get(domain.ClusterId(stackId))
	if err != nil {
		return out, "", err
	}

	organization, err := u.organizationRepo.Get(cluster.OrganizationId)
	if err != nil {
		return out, "", err
	}

	// cluster status
	step := parseStatusDescription(cluster.StatusDesc)
	clusterStepStatus := domain.StackStepStatus{
		Status:  cluster.Status.String(),
		Stage:   "CLUSTER",
		Step:    step,
		MaxStep: domain.MAX_STEP_CLUSTER_CREATE,
	}
	if cluster.Status == domain.ClusterStatus_DELETING {
		clusterStepStatus.MaxStep = domain.MAX_STEP_CLUSTER_REMOVE
	}
	out = append(out, clusterStepStatus)

	// make default appgroup status
	if strings.Contains(cluster.StackTemplate.Template, "aws-reference") || strings.Contains(cluster.StackTemplate.Template, "eks-reference") {
		// LMA
		out = append(out, domain.StackStepStatus{
			Status:  domain.AppGroupStatus_PENDING.String(),
			Stage:   "LMA",
			Step:    0,
			MaxStep: domain.MAX_STEP_LMA_CREATE_MEMBER,
		})
	} else if strings.Contains(cluster.StackTemplate.Template, "aws-msa-reference") || strings.Contains(cluster.StackTemplate.Template, "eks-msa-reference") {
		// LMA + SERVICE_MESH
		out = append(out, domain.StackStepStatus{
			Status:  domain.AppGroupStatus_PENDING.String(),
			Stage:   "LMA",
			Step:    0,
			MaxStep: domain.MAX_STEP_LMA_CREATE_MEMBER,
		})
		out = append(out, domain.StackStepStatus{
			Status:  domain.AppGroupStatus_PENDING.String(),
			Stage:   "SERVICE_MESH",
			Step:    0,
			MaxStep: domain.MAX_STEP_SM_CREATE,
		})
	}

	appGroups, err := u.appGroupRepo.Fetch(domain.ClusterId(stackId), nil)
	for _, appGroup := range appGroups {
		for i, step := range out {
			if step.Stage == appGroup.AppGroupType.String() {
				step := parseStatusDescription(appGroup.StatusDesc)

				out[i].Status = appGroup.Status.String()

				if appGroup.AppGroupType == domain.AppGroupType_LMA {
					out[i].MaxStep = domain.MAX_STEP_LMA_CREATE_MEMBER
					if organization.PrimaryClusterId == cluster.ID.String() {
						out[i].MaxStep = domain.MAX_STEP_LMA_CREATE_PRIMARY
					}
					if appGroup.Status == domain.AppGroupStatus_DELETING || cluster.Status == domain.ClusterStatus_DELETING {
						out[i].MaxStep = domain.MAX_STEP_LMA_REMOVE
					}
				} else {
					out[i].MaxStep = domain.MAX_STEP_SM_CREATE
					if appGroup.Status == domain.AppGroupStatus_DELETING || cluster.Status == domain.ClusterStatus_DELETING {
						out[i].MaxStep = domain.MAX_STEP_SM_REMOVE
					}
				}
				out[i].Step = step
				if out[i].Step > out[i].MaxStep {
					out[i].Step = out[i].MaxStep
				}
			}
		}
	}

	status, _ := getStackStatus(cluster, appGroups)
	stackStatus = status.String()

	// sort
	// deleting : service_mesh -> lma -> cluster
	// installing : cluster -> lma -> service_mesh
	if status == domain.StackStatus_APPGROUP_DELETING || status == domain.StackStatus_CLUSTER_DELETING {
		reversed := make([]domain.StackStepStatus, len(out))
		j := 0
		for i := len(out) - 1; i >= 0; i-- {
			reversed[j] = out[i]
			j++
		}
		out = reversed
	}

	return
}

func (u *StackUsecase) SetFavorite(ctx context.Context, stackId domain.StackId) error {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return httpErrors.NewUnauthorizedError(fmt.Errorf("Invalid token"), "A_INVALID_TOKEN", "")
	}

	err := u.clusterRepo.SetFavorite(domain.ClusterId(stackId), user.GetUserId())
	if err != nil {
		return err
	}

	return nil
}

func (u *StackUsecase) DeleteFavorite(ctx context.Context, stackId domain.StackId) error {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return httpErrors.NewUnauthorizedError(fmt.Errorf("Invalid token"), "A_INVALID_TOKEN", "")
	}

	err := u.clusterRepo.DeleteFavorite(domain.ClusterId(stackId), user.GetUserId())
	if err != nil {
		return err
	}

	return nil
}

func (u *StackUsecase) GetNodes(ctx context.Context, stackId domain.StackId) (out domain.Stack, err error) {
	cluster, err := u.clusterRepo.Get(domain.ClusterId(stackId))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return out, httpErrors.NewNotFoundError(err, "S_FAILED_FETCH_CLUSTER", "")
		}
		return out, err
	}
	if cluster.CloudService != domain.CloudService_BYOH {
		return out, httpErrors.NewBadRequestError(fmt.Errorf("Invalid cloud service"), "", "")
	}

	client, err := kubernetes.GetClientAdminCluster()
	if err != nil {
		return out, err
	}

	hosts := byoh.ByoHostList{}
	data, err := client.RESTClient().
		Get().
		AbsPath("/apis/infrastructure.cluster.x-k8s.io/v1beta1").
		Namespace("default").
		//Namespace(cluster.ID). [TODO]
		Resource("byohosts").
		DoRaw(ctx)

	if err := json.Unmarshal(data, &hosts); err != nil {
		return out, err
	}

	/* FOR DEBUG
	for _, host := range hosts.Items {
		log.Info(host.Name)
		log.Info(host.Labels)
		log.Info(host.Status.Conditions[0].Type)
	}
	*/

	stackNodeStatus := func(targeted int, registered int) string {
		if targeted <= registered {
			return "COMPLETED"
		}
		return "INPROGRESS"
	}

	tksCpNodeRegistered, tksCpNodeRegistering, tksCpHosts := 0, 0, make([]domain.StackHost, 0)
	tksInfraNodeRegistered, tksInfraNodeRegistering, tksInfraHosts := 0, 0, make([]domain.StackHost, 0)
	tksUserNodeRegistered, tksUserNodeRegistering, tksUserHosts := 0, 0, make([]domain.StackHost, 0)
	for _, host := range hosts.Items {
		hostStatus := host.Status.Conditions[0].Type
		registered, registering := 0, 0
		if hostStatus == "K8sNodeBootstrapSucceeded" {
			registered = 1
		} else {
			registering = 1
		}

		switch host.Labels["role"] {
		case "tks":
			tksCpNodeRegistered = tksCpNodeRegistered + registered
			tksCpNodeRegistering = tksCpNodeRegistering + registering
			tksCpHosts = append(tksCpHosts, domain.StackHost{Name: host.Name, Status: string(hostStatus)})
		case "worker":
			tksInfraNodeRegistered = tksInfraNodeRegistered + registered
			tksInfraNodeRegistering = tksInfraNodeRegistering + registering
			tksInfraHosts = append(tksInfraHosts, domain.StackHost{Name: host.Name, Status: string(hostStatus)})
		case "3":
			tksUserNodeRegistered = tksUserNodeRegistered + registered
			tksUserNodeRegistering = tksUserNodeRegistering + registering
			tksUserHosts = append(tksUserHosts, domain.StackHost{Name: host.Name, Status: string(hostStatus)})
		}
	}

	out.Nodes = []domain.StackNode{
		{
			Type:        "TKS_CP_NODE",
			Targeted:    cluster.Conf.TksCpNode,
			Registered:  tksCpNodeRegistered,
			Registering: tksCpNodeRegistering,
			Status:      stackNodeStatus(cluster.Conf.TksCpNode, tksCpNodeRegistered),
			Command:     "curl -fL http://192.168.0.77/tks-byoh-hostagent-install.sh | sh -s CLUSTER-ID-control-plane",
			Validity:    3600,
			Hosts:       tksCpHosts,
		},
		{
			Type:        "TKS_INFRA_NODE",
			Targeted:    cluster.Conf.TksInfraNode,
			Registered:  tksInfraNodeRegistered,
			Registering: tksInfraNodeRegistering,
			Status:      stackNodeStatus(cluster.Conf.TksInfraNode, tksUserNodeRegistered),
			Command:     "curl -fL http://192.168.0.77/tks-byoh-hostagent-install.sh | sh -s CLUSTER-ID-tks-worker",
			Validity:    3600,
			Hosts:       tksInfraHosts,
		},
		{
			Type:        "TKS_USER_NODE",
			Targeted:    cluster.Conf.TksUserNode,
			Registered:  tksUserNodeRegistered,
			Registering: tksUserNodeRegistering,
			Status:      stackNodeStatus(cluster.Conf.TksUserNode, tksUserNodeRegistered),
			Command:     "curl -fL http://192.168.0.77/tks-byoh-hostagent-install.sh | sh -s CLUSTER-ID-user-worker",
			Validity:    3600,
			Hosts:       tksInfraHosts,
		},
	}

	// [TODO] for integration
	/*
		out.Nodes = []domain.StackNodeResponse{
			{
				ID:         "1",
				Type:       "TKS_CP_NODE",
				Targeted:   3,
				Registered: 1,
				Status:     "INPROGRESS",
				Command:    "curl -fL http://192.168.0.77/tks-byoh-hostagent-install.sh | sh -s CLUSTER-ID-control-plane",
				Validity:   3000,
			},
			{
				ID:         "2",
				Type:       "TKS_INFRA_NODE",
				Targeted:   0,
				Registered: 0,
				Status:     "PENDING",
				Command:    "curl -fL http://192.168.0.77/tks-byoh-hostagent-install.sh | sh -s CLUSTER-ID-control-plane",
				Validity:   3000,
			},
			{
				ID:         "3",
				Type:       "TKS_USER_NODE",
				Targeted:   3,
				Registered: 3,
				Status:     "COMPLETED",
				Command:    "curl -fL http://192.168.0.77/tks-byoh-hostagent-install.sh | sh -s CLUSTER-ID-control-plane",
				Validity:   3000,
			},
		}
	*/

	return
}

func reflectClusterToStack(cluster domain.Cluster, appGroups []domain.AppGroup) (out domain.Stack) {
	if err := serializer.Map(cluster, &out); err != nil {
		log.Error(err)
	}

	status, statusDesc := getStackStatus(cluster, appGroups)

	out.ID = domain.StackId(cluster.ID)
	out.Status = status
	out.StatusDesc = statusDesc

	/*
		return domain.Stack{
			ID:              domain.StackId(cluster.ID),
			OrganizationId:  cluster.OrganizationId,
			Name:            cluster.Name,
			Description:     cluster.Description,
			Status:          status,
			StatusDesc:      statusDesc,
			CloudAccountId:  cluster.CloudAccountId,
			CloudAccount:    cluster.CloudAccount,
			StackTemplateId: cluster.StackTemplateId,
			StackTemplate:   cluster.StackTemplate,
			CreatorId:       cluster.CreatorId,
			Creator:         cluster.Creator,
			UpdatorId:       cluster.UpdatorId,
			Updator:         cluster.Updator,
			CreatedAt:       cluster.CreatedAt,
			UpdatedAt:       cluster.UpdatedAt,
			Conf: domain.StackConf{
				TksCpNode:        cluster.Conf.TksCpNode,
				TksCpNodeMax:     cluster.Conf.TksCpNodeMax,
				TksCpNodeType:    cluster.Conf.TksCpNodeType,
				TksInfraNode:     cluster.Conf.TksInfraNode,
				TksInfraNodeMax:  cluster.Conf.TksInfraNodeMax,
				TksInfraNodeType: cluster.Conf.TksInfraNodeType,
				TksUserNode:      cluster.Conf.TksUserNode,
				TksUserNodeMax:   cluster.Conf.TksUserNodeMax,
				TksUserNodeType:  cluster.Conf.TksUserNodeType,
			},
		}
	*/
	return
}

// [TODO] more pretty
func getStackStatus(cluster domain.Cluster, appGroups []domain.AppGroup) (domain.StackStatus, string) {
	for _, appGroup := range appGroups {
		if appGroup.Status == domain.AppGroupStatus_PENDING && cluster.Status == domain.ClusterStatus_RUNNING {
			return domain.StackStatus_APPGROUP_INSTALLING, appGroup.StatusDesc
		}
		if appGroup.Status == domain.AppGroupStatus_INSTALLING {
			return domain.StackStatus_APPGROUP_INSTALLING, appGroup.StatusDesc
		}
		if appGroup.Status == domain.AppGroupStatus_DELETING {
			return domain.StackStatus_APPGROUP_DELETING, appGroup.StatusDesc
		}
		if appGroup.Status == domain.AppGroupStatus_INSTALL_ERROR {
			return domain.StackStatus_APPGROUP_INSTALL_ERROR, appGroup.StatusDesc
		}
		if appGroup.Status == domain.AppGroupStatus_DELETE_ERROR {
			return domain.StackStatus_APPGROUP_DELETE_ERROR, appGroup.StatusDesc
		}
	}

	if cluster.Status == domain.ClusterStatus_INSTALLING {
		return domain.StackStatus_CLUSTER_INSTALLING, cluster.StatusDesc
	}
	if cluster.Status == domain.ClusterStatus_DELETING {
		return domain.StackStatus_CLUSTER_DELETING, cluster.StatusDesc
	}
	if cluster.Status == domain.ClusterStatus_DELETED {
		return domain.StackStatus_CLUSTER_DELETED, cluster.StatusDesc
	}
	if cluster.Status == domain.ClusterStatus_INSTALL_ERROR {
		return domain.StackStatus_CLUSTER_INSTALL_ERROR, cluster.StatusDesc
	}
	if cluster.Status == domain.ClusterStatus_DELETE_ERROR {
		return domain.StackStatus_CLUSTER_DELETE_ERROR, cluster.StatusDesc
	}

	// workflow 중간 중간 비는 status 처리...
	if strings.Contains(cluster.StackTemplate.Template, "aws-reference") || strings.Contains(cluster.StackTemplate.Template, "eks-reference") {
		if len(appGroups) < 1 {
			return domain.StackStatus_APPGROUP_INSTALLING, "(0/0)"
		} else {
			for _, appGroup := range appGroups {
				if appGroup.Status == domain.AppGroupStatus_DELETED {
					return domain.StackStatus_CLUSTER_DELETING, "(0/0)"
				}
			}
		}
	} else if strings.Contains(cluster.StackTemplate.Template, "aws-msa-reference") || strings.Contains(cluster.StackTemplate.Template, "eks-msa-reference") {
		if len(appGroups) < 2 {
			return domain.StackStatus_APPGROUP_INSTALLING, "(0/0)"
		} else {
			deletedAppGroupCnt := 0
			for _, appGroup := range appGroups {
				if appGroup.Status == domain.AppGroupStatus_DELETED {
					deletedAppGroupCnt++
				}
			}
			if deletedAppGroupCnt == 1 {
				return domain.StackStatus_APPGROUP_DELETING, "(0/0)"
			} else if deletedAppGroupCnt == 2 {
				return domain.StackStatus_CLUSTER_DELETING, "(0/0)"
			}
		}
	}

	return domain.StackStatus_RUNNING, cluster.StatusDesc

}

func parseStatusDescription(statusDesc string) (step int) {
	// (20/20)
	if statusDesc == "" {
		return 0
	}

	maxStep := 0
	_, err := fmt.Sscanf(statusDesc, "(%d/%d)", &step, &maxStep)
	if err != nil {
		step = 0
	}
	return
}

func getServiceQuota(client *servicequotas.Client, quotaCode string, serviceCode string) (res *servicequotas.GetServiceQuotaOutput, err error) {
	res, err = client.GetServiceQuota(context.TODO(), &servicequotas.GetServiceQuotaInput{
		QuotaCode:   &quotaCode,
		ServiceCode: &serviceCode,
	}, func(o *servicequotas.Options) {
		o.Region = "ap-northeast-2"
	})
	if err != nil {
		return nil, err
	}
	return
}
