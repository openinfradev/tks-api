package usecase

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/helper"
	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/internal/serializer"
	argowf "github.com/openinfradev/tks-api/pkg/argo-client"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/kubernetes"
	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type IStackUsecase interface {
	Get(ctx context.Context, stackId domain.StackId) (model.Stack, error)
	GetByName(ctx context.Context, organizationId string, name string) (model.Stack, error)
	Fetch(ctx context.Context, organizationId string, pg *pagination.Pagination) ([]model.Stack, error)
	Create(ctx context.Context, dto model.Stack) (stackId domain.StackId, err error)
	Install(ctx context.Context, stackId domain.StackId) (err error)
	Update(ctx context.Context, dto model.Stack) error
	Delete(ctx context.Context, dto model.Stack) error
	GetKubeConfig(ctx context.Context, stackId domain.StackId) (kubeConfig string, err error)
	GetStepStatus(ctx context.Context, stackId domain.StackId) (out []domain.StackStepStatus, stackStatus string, err error)
	SetFavorite(ctx context.Context, stackId domain.StackId) error
	DeleteFavorite(ctx context.Context, stackId domain.StackId) error
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

func (u *StackUsecase) Create(ctx context.Context, dto model.Stack) (stackId domain.StackId, err error) {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return "", httpErrors.NewUnauthorizedError(fmt.Errorf("Invalid token"), "A_INVALID_TOKEN", "")
	}

	_, err = u.GetByName(ctx, dto.OrganizationId, dto.Name)
	if err == nil {
		return "", httpErrors.NewBadRequestError(httpErrors.DuplicateResource, "S_CREATE_ALREADY_EXISTED_NAME", "")
	}

	stackTemplate, err := u.stackTemplateRepo.Get(ctx, dto.StackTemplateId)
	if err != nil {
		return "", httpErrors.NewInternalServerError(errors.Wrap(err, "Invalid stackTemplateId"), "S_INVALID_STACK_TEMPLATE", "")
	}

	clusters, err := u.clusterRepo.FetchByOrganizationId(ctx, dto.OrganizationId, user.GetUserId(), nil)
	if err != nil {
		return "", httpErrors.NewInternalServerError(errors.Wrap(err, "Failed to get clusters"), "S_FAILED_GET_CLUSTERS", "")
	}
	isPrimary := false
	if len(clusters) == 0 {
		isPrimary = true
	}
	log.Debug(ctx, "isPrimary ", isPrimary)

	if dto.CloudService == domain.CloudService_BYOH {
		if dto.ClusterEndpoint == "" {
			return "", httpErrors.NewBadRequestError(fmt.Errorf("Invalid clusterEndpoint"), "S_INVALID_ADMINCLUSTER_URL", "")
		}
		arr := strings.Split(dto.ClusterEndpoint, ":")
		if len(arr) != 2 {
			return "", httpErrors.NewBadRequestError(fmt.Errorf("Invalid clusterEndpoint"), "S_INVALID_ADMINCLUSTER_URL", "")
		}
	} else {
		if _, err = u.cloudAccountRepo.Get(ctx, dto.CloudAccountId); err != nil {
			return "", httpErrors.NewInternalServerError(errors.Wrap(err, "Invalid cloudAccountId"), "S_INVALID_CLOUD_ACCOUNT", "")
		}
	}

	// Make stack nodes
	// [TODO] to be advanced feature
	dto.Conf.TksCpNodeMax = dto.Conf.TksCpNode
	dto.Conf.TksInfraNodeMax = dto.Conf.TksInfraNode
	dto.Conf.TksUserNodeMax = dto.Conf.TksUserNode
	if stackTemplate.CloudService == "AWS" && stackTemplate.KubeType == "AWS" {
		if dto.Conf.TksCpNode == 0 {
			dto.Conf.TksCpNode = 3
			dto.Conf.TksCpNodeMax = 3
			dto.Conf.TksInfraNode = 3
			dto.Conf.TksInfraNodeMax = 3
		}

		// user 노드는 MAX_AZ_NUM의 배수로 요청한다.
		if dto.Conf.TksUserNode%domain.MAX_AZ_NUM != 0 {
			return "", httpErrors.NewInternalServerError(errors.Wrap(err, "Invalid node count"), "", "")
		}
	}

	var conf domain.StackConfResponse
	if err := serializer.Map(ctx, dto.Conf, &conf); err != nil {
		log.Error(ctx, err)
		return "", httpErrors.NewInternalServerError(errors.Wrap(err, "Invalid node conf"), "", "")
	}

	workflow := "tks-stack-create"
	workflowId, err := u.argo.SumbitWorkflowFromWftpl(ctx, workflow, argowf.SubmitOptions{
		Parameters: []string{
			fmt.Sprintf("tks_api_url=%s", viper.GetString("external-address")),
			"cluster_name=" + dto.Name,
			"description=" + dto.Description,
			"organization_id=" + dto.OrganizationId,
			"cloud_account_id=" + dto.CloudAccountId.String(),
			"stack_template_id=" + dto.StackTemplateId.String(),
			"creator=" + user.GetUserId().String(),
			"base_repo_branch=" + viper.GetString("revision"),
			"infra_conf=" + strings.Replace(helper.ModelToJson(conf), "\"", "\\\"", -1),
			"cloud_service=" + dto.CloudService,
			"cluster_endpoint=" + dto.ClusterEndpoint,
			"policy_ids=" + strings.Join(dto.PolicyIds, ","),
		},
	})
	if err != nil {
		log.Error(ctx, err)
		return "", httpErrors.NewInternalServerError(err, "S_FAILED_TO_CALL_WORKFLOW", "")
	}
	log.Debug(ctx, "Submitted workflow: ", workflowId)

	// wait & get clusterId ( max 1min 	)
	dto.ID = domain.StackId("")
	for i := 0; i < 60; i++ {
		time.Sleep(time.Second * 5)
		workflow, err := u.argo.GetWorkflow(ctx, "argo", workflowId)
		if err != nil {
			return "", err
		}

		log.Debug(ctx, "workflow ", workflow)
		if workflow.Status.Phase != "" && workflow.Status.Phase != "Running" {
			return "", fmt.Errorf("Invalid workflow status [%s]", workflow.Status.Phase)
		}

		cluster, err := u.clusterRepo.GetByName(ctx, dto.OrganizationId, dto.Name)
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

	_, err = u.stackTemplateRepo.Get(ctx, cluster.StackTemplateId)
	if err != nil {
		return httpErrors.NewInternalServerError(errors.Wrap(err, "Invalid stackTemplateId"), "S_INVALID_STACK_TEMPLATE", "")
	}

	clusters, err := u.clusterRepo.FetchByOrganizationId(ctx, cluster.OrganizationId, uuid.Nil, nil)
	if err != nil {
		return httpErrors.NewInternalServerError(errors.Wrap(err, "Failed to get clusters"), "S_FAILED_GET_CLUSTERS", "")
	}
	isPrimary := false
	if len(clusters) == 0 {
		isPrimary = true
	}
	log.Debug(ctx, "isPrimary ", isPrimary)

	if cluster.CloudService != domain.CloudService_BYOH {
		return httpErrors.NewBadRequestError(fmt.Errorf("Invalid cloud service"), "S_INVALID_CLOUD_SERVICE", "")
	}

	// Make stack nodes
	var stackConf domain.StackConfResponse
	if err = serializer.Map(ctx, cluster, &stackConf); err != nil {
		log.Info(ctx, err)
	}

	workflow := "tks-stack-install"
	workflowId, err := u.argo.SumbitWorkflowFromWftpl(ctx, workflow, argowf.SubmitOptions{
		Parameters: []string{
			fmt.Sprintf("tks_api_url=%s", viper.GetString("external-address")),
			"cluster_id=" + cluster.ID.String(),
			"description=" + cluster.Description,
			"organization_id=" + cluster.OrganizationId,
			"stack_template_id=" + cluster.StackTemplateId.String(),
			"creator=" + (*cluster.CreatorId).String(),
			"base_repo_branch=" + viper.GetString("revision"),
		},
	})
	if err != nil {
		log.Error(ctx, err)
		return httpErrors.NewInternalServerError(err, "S_FAILED_TO_CALL_WORKFLOW", "")
	}
	log.Debug(ctx, "Submitted workflow: ", workflowId)

	return nil
}

func (u *StackUsecase) Get(ctx context.Context, stackId domain.StackId) (out model.Stack, err error) {
	cluster, err := u.clusterRepo.Get(ctx, domain.ClusterId(stackId))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return out, httpErrors.NewNotFoundError(err, "S_FAILED_FETCH_CLUSTER", "")
		}
		return out, err
	}

	organization, err := u.organizationRepo.Get(ctx, cluster.OrganizationId)
	if err != nil {
		return out, httpErrors.NewInternalServerError(errors.Wrap(err, fmt.Sprintf("Failed to get organization for clusterId %s", domain.ClusterId(stackId))), "S_FAILED_FETCH_ORGANIZATION", "")
	}

	appGroups, err := u.appGroupRepo.Fetch(ctx, domain.ClusterId(stackId), nil)
	if err != nil {
		return out, err
	}

	out = reflectClusterToStack(ctx, cluster, appGroups)

	if organization.PrimaryClusterId == cluster.ID.String() {
		out.PrimaryCluster = true
	}

	stackResources, _ := u.dashbordUsecase.GetStacks(ctx, cluster.OrganizationId)
	for _, resource := range stackResources {
		if resource.ID == domain.StackId(cluster.ID) {
			if err := serializer.Map(ctx, resource, &out.Resource); err != nil {
				log.Error(ctx, err)
			}
		}
	}

	appGroupsInPrimaryCluster, err := u.appGroupRepo.Fetch(ctx, domain.ClusterId(organization.PrimaryClusterId), nil)
	if err != nil {
		return out, err
	}

	for _, appGroup := range appGroupsInPrimaryCluster {
		if appGroup.AppGroupType == domain.AppGroupType_LMA {
			applications, err := u.appGroupRepo.GetApplications(ctx, appGroup.ID, domain.ApplicationType_GRAFANA)
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

func (u *StackUsecase) GetByName(ctx context.Context, organizationId string, name string) (out model.Stack, err error) {
	cluster, err := u.clusterRepo.GetByName(ctx, organizationId, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return out, httpErrors.NewNotFoundError(err, "S_FAILED_FETCH_CLUSTER", "")
		}
		return out, err
	}

	appGroups, err := u.appGroupRepo.Fetch(ctx, cluster.ID, nil)
	if err != nil {
		return out, err
	}

	out = reflectClusterToStack(ctx, cluster, appGroups)
	return
}

func (u *StackUsecase) Fetch(ctx context.Context, organizationId string, pg *pagination.Pagination) (out []model.Stack, err error) {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return out, httpErrors.NewUnauthorizedError(fmt.Errorf("Invalid token"), "A_INVALID_TOKEN", "")
	}

	organization, err := u.organizationRepo.Get(ctx, organizationId)
	if err != nil {
		return out, httpErrors.NewInternalServerError(errors.Wrap(err, fmt.Sprintf("Failed to get organization for clusterId %s", organizationId)), "S_FAILED_FETCH_ORGANIZATION", "")
	}

	clusters, err := u.clusterRepo.FetchByOrganizationId(ctx, organizationId, user.GetUserId(), pg)
	if err != nil {
		return out, err
	}

	stackResources, _ := u.dashbordUsecase.GetStacks(ctx, organizationId)

	for _, cluster := range clusters {
		appGroups, err := u.appGroupRepo.Fetch(ctx, cluster.ID, nil)
		if err != nil {
			return nil, err
		}

		outStack := reflectClusterToStack(ctx, cluster, appGroups)
		if organization.PrimaryClusterId == cluster.ID.String() {
			outStack.PrimaryCluster = true
		}

		for _, appGroup := range appGroups {
			if appGroup.AppGroupType == domain.AppGroupType_LMA {
				applications, err := u.appGroupRepo.GetApplications(ctx, appGroup.ID, domain.ApplicationType_GRAFANA)
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
				if err := serializer.Map(ctx, resource, &outStack.Resource); err != nil {
					log.Error(ctx, err)
				}
			}
		}

		if cluster.Favorites != nil && len(*cluster.Favorites) > 0 {
			outStack.Favorited = true
		} else {
			outStack.Favorited = false
		}

		out = append(out, outStack)
	}

	sort.Slice(out, func(i, j int) bool {
		return string(out[i].ID) == organization.PrimaryClusterId
	})

	return
}

func (u *StackUsecase) Update(ctx context.Context, dto model.Stack) (err error) {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return httpErrors.NewBadRequestError(fmt.Errorf("Invalid token"), "", "")
	}

	_, err = u.clusterRepo.Get(ctx, domain.ClusterId(dto.ID))
	if err != nil {
		return httpErrors.NewNotFoundError(err, "S_FAILED_FETCH_CLUSTER", "")
	}

	updatorId := user.GetUserId()
	dtoCluster := model.Cluster{
		ID:          domain.ClusterId(dto.ID),
		Description: dto.Description,
		UpdatorId:   &updatorId,
	}

	err = u.clusterRepo.Update(ctx, dtoCluster)
	if err != nil {
		return err
	}

	return nil
}

func (u *StackUsecase) Delete(ctx context.Context, dto model.Stack) (err error) {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return httpErrors.NewBadRequestError(fmt.Errorf("Invalid token"), "", "")
	}

	cluster, err := u.clusterRepo.Get(ctx, domain.ClusterId(dto.ID))
	if err != nil {
		return httpErrors.NewBadRequestError(errors.Wrap(err, "Failed to get cluster"), "S_FAILED_FETCH_CLUSTER", "")
	}

	// 지우려고 하는 stack 이 primary cluster 라면, organization 내에 cluster 가 자기 자신만 남아있을 경우이다.
	organizations, err := u.organizationRepo.Fetch(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "Failed to get organizations")
	}

	for _, organization := range *organizations {
		if organization.PrimaryClusterId == cluster.ID.String() {

			clusters, err := u.clusterRepo.FetchByOrganizationId(ctx, organization.ID, user.GetUserId(), nil)
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
	appGroups, err := u.appGroupRepo.Fetch(ctx, domain.ClusterId(dto.ID), nil)
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

	// Check AppServing
	appsCnt, err := u.appServeAppRepo.GetNumOfAppsOnStack(ctx, dto.OrganizationId, dto.ID.String())
	if err != nil {
		return errors.Wrap(err, "Failed to get numOfAppsOnStack")
	}
	if appsCnt > 0 {
		return httpErrors.NewBadRequestError(fmt.Errorf("existed appServeApps in %s", dto.OrganizationId), "S_FAILED_DELETE_EXISTED_ASA", "")
	}

	// Policy 삭제

	workflow := "tks-stack-delete"
	workflowId, err := u.argo.SumbitWorkflowFromWftpl(ctx, workflow, argowf.SubmitOptions{
		Parameters: []string{
			fmt.Sprintf("tks_api_url=%s", viper.GetString("external-address")),
			"organization_id=" + dto.OrganizationId,
			"cluster_id=" + dto.ID.String(),
			"cloud_account_id=" + cluster.CloudAccount.ID.String(),
			"stack_template_id=" + cluster.StackTemplate.ID.String(),
		},
	})
	if err != nil {
		log.Error(ctx, err)
		return err
	}
	log.Debug(ctx, "Submitted workflow: ", workflowId)

	// Remove Cluster & AppGroup status description
	if err := u.appGroupRepo.InitWorkflowDescription(ctx, cluster.ID); err != nil {
		log.Error(ctx, err)
	}
	if err := u.clusterRepo.InitWorkflowDescription(ctx, cluster.ID); err != nil {
		log.Error(ctx, err)
	}

	// wait & get clusterId ( max 1min 	)
	for i := 0; i < 60; i++ {
		time.Sleep(time.Second * 2)
		workflow, err := u.argo.GetWorkflow(ctx, "argo", workflowId)
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
	kubeconfig, err := kubernetes.GetKubeConfig(ctx, stackId.String(), kubernetes.KubeconfigForUser)
	//kubeconfig, err := kubernetes.GetKubeConfig("cmsai5k5l")
	if err != nil {
		return "", err
	}

	return string(kubeconfig[:]), nil
}

// [TODO] need more pretty...
func (u *StackUsecase) GetStepStatus(ctx context.Context, stackId domain.StackId) (out []domain.StackStepStatus, stackStatus string, err error) {
	cluster, err := u.clusterRepo.Get(ctx, domain.ClusterId(stackId))
	if err != nil {
		return out, "", err
	}

	organization, err := u.organizationRepo.Get(ctx, cluster.OrganizationId)
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
	if cluster.StackTemplate.TemplateType == "STANDARD" {
		// LMA
		out = append(out, domain.StackStepStatus{
			Status:  domain.AppGroupStatus_PENDING.String(),
			Stage:   "LMA",
			Step:    0,
			MaxStep: domain.MAX_STEP_LMA_CREATE_MEMBER,
		})
	} else {
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

	appGroups, err := u.appGroupRepo.Fetch(ctx, domain.ClusterId(stackId), nil)
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

	err := u.clusterRepo.SetFavorite(ctx, domain.ClusterId(stackId), user.GetUserId())
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

	err := u.clusterRepo.DeleteFavorite(ctx, domain.ClusterId(stackId), user.GetUserId())
	if err != nil {
		return err
	}

	return nil
}

func reflectClusterToStack(ctx context.Context, cluster model.Cluster, appGroups []model.AppGroup) (out model.Stack) {
	if err := serializer.Map(ctx, cluster, &out); err != nil {
		log.Error(ctx, err)
	}

	if err := serializer.Map(ctx, cluster, &out.Conf); err != nil {
		log.Error(ctx, err)
	}

	status, statusDesc := getStackStatus(cluster, appGroups)

	out.ID = domain.StackId(cluster.ID)
	out.Status = status
	out.StatusDesc = statusDesc

	/*
		return model.Stack{
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
func getStackStatus(cluster model.Cluster, appGroups []model.AppGroup) (domain.StackStatus, string) {
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

	if cluster.Status == domain.ClusterStatus_BOOTSTRAPPING {
		return domain.StackStatus_CLUSTER_BOOTSTRAPPING, cluster.StatusDesc
	}
	if cluster.Status == domain.ClusterStatus_BOOTSTRAPPED {
		return domain.StackStatus_CLUSTER_BOOTSTRAPPED, cluster.StatusDesc
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
	if cluster.StackTemplate.TemplateType == "STANDARD" {
		if len(appGroups) < 1 {
			return domain.StackStatus_APPGROUP_INSTALLING, "(0/0)"
		} else {
			for _, appGroup := range appGroups {
				if appGroup.Status == domain.AppGroupStatus_DELETED {
					return domain.StackStatus_CLUSTER_DELETING, "(0/0)"
				}
			}
		}
	} else if cluster.StackTemplate.TemplateType == "MSA" {
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
