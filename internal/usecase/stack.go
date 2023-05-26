package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/openinfradev/tks-api/internal/middleware/auth/request"

	"github.com/openinfradev/tks-api/internal/helper"
	"github.com/openinfradev/tks-api/internal/kubernetes"
	"github.com/openinfradev/tks-api/internal/repository"
	argowf "github.com/openinfradev/tks-api/pkg/argo-client"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type IStackUsecase interface {
	Get(ctx context.Context, stackId domain.StackId) (domain.Stack, error)
	GetByName(ctx context.Context, organizationId string, name string) (domain.Stack, error)
	Fetch(ctx context.Context, organizationId string) ([]domain.Stack, error)
	Create(ctx context.Context, dto domain.Stack) (stackId domain.StackId, err error)
	Update(ctx context.Context, dto domain.Stack) error
	Delete(ctx context.Context, dto domain.Stack) error
	GetKubeConfig(ctx context.Context, stackId domain.StackId) (kubeConfig string, err error)
	GetStepStatus(ctx context.Context, stackId domain.StackId) (out []domain.StackStepStatus, stackStatus string, err error)
}

type StackUsecase struct {
	clusterRepo       repository.IClusterRepository
	appGroupRepo      repository.IAppGroupRepository
	cloudAccountRepo  repository.ICloudAccountRepository
	organizationRepo  repository.IOrganizationRepository
	stackTemplateRepo repository.IStackTemplateRepository
	argo              argowf.ArgoClient
}

func NewStackUsecase(r repository.Repository, argoClient argowf.ArgoClient) IStackUsecase {
	return &StackUsecase{
		clusterRepo:       r.Cluster,
		appGroupRepo:      r.AppGroup,
		cloudAccountRepo:  r.CloudAccount,
		organizationRepo:  r.Organization,
		stackTemplateRepo: r.StackTemplate,
		argo:              argoClient,
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

	stackTemplate, err := u.stackTemplateRepo.Get(dto.StackTemplateId)
	if err != nil {
		return "", httpErrors.NewInternalServerError(errors.Wrap(err, "Invalid stackTemplateId"), "S_INVALID_STACK_TEMPLATE", "")
	}

	_, err = u.cloudAccountRepo.Get(dto.CloudAccountId)
	if err != nil {
		return "", httpErrors.NewInternalServerError(errors.Wrap(err, "Invalid cloudAccountId"), "S_INVALID_CLOUD_ACCOUNT", "")
	}

	// [TODO] check primary cluster
	clusters, err := u.clusterRepo.FetchByOrganizationId(dto.OrganizationId)
	if err != nil {
		return "", httpErrors.NewInternalServerError(errors.Wrap(err, "Failed to get clusters"), "S_FAILED_GET_CLUSTERS", "")
	}
	isPrimary := false
	if len(clusters) == 0 {
		isPrimary = true
	}
	log.DebugWithContext(ctx, "isPrimary ", isPrimary)

	workflow := ""
	if strings.Contains(stackTemplate.Template, "aws-reference") || strings.Contains(stackTemplate.Template, "eks-reference") {
		workflow = "tks-stack-create-aws"
	} else if strings.Contains(stackTemplate.Template, "aws-msa-reference") || strings.Contains(stackTemplate.Template, "eks-msa-reference") {
		workflow = "tks-stack-create-aws-msa"
	} else {
		log.ErrorWithContext(ctx, "Invalid template  : ", stackTemplate.Template)
		return "", httpErrors.NewInternalServerError(fmt.Errorf("Invalid stackTemplate. %s", stackTemplate.Template), "", "")
	}

	var stackConf domain.StackConfResponse
	if err = domain.Map(dto.Conf, &stackConf); err != nil {
		log.InfoWithContext(ctx, err)
	}

	workflowId, err := u.argo.SumbitWorkflowFromWftpl(workflow, argowf.SubmitOptions{
		Parameters: []string{
			fmt.Sprintf("tks_api_url=%s", viper.GetString("external-address")),
			"cluster_name=" + dto.Name,
			"description=" + dto.Description,
			"organization_id=" + dto.OrganizationId,
			"cloud_account_id=" + dto.CloudAccountId.String(),
			"stack_template_id=" + dto.StackTemplateId.String(),
			"creator=" + user.GetUserId().String(),
			"infra_conf=" + strings.Replace(helper.ModelToJson(stackConf), "\"", "\\\"", -1),
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
		return out, httpErrors.NewInternalServerError(errors.Wrap(err, fmt.Sprintf("Failed to get organization for clusterId %s", cluster.OrganizationId)), "S_FAILED_FETCH_ORGANIZATION", "")
	}

	appGroups, err := u.appGroupRepo.Fetch(domain.ClusterId(stackId))
	if err != nil {
		return out, err
	}

	out = reflectClusterToStack(cluster, appGroups)
	if organization.PrimaryClusterId == cluster.ID.String() {
		out.PrimaryCluster = true
	}

	for _, appGroup := range appGroups {
		if appGroup.AppGroupType == domain.AppGroupType_LMA {
			applications, err := u.appGroupRepo.GetApplications(appGroup.ID, domain.ApplicationType_GRAFANA)
			if err != nil {
				return out, err
			}
			if len(applications) > 0 {
				out.GrafanaUrl = applications[0].Endpoint
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

	appGroups, err := u.appGroupRepo.Fetch(cluster.ID)
	if err != nil {
		return out, err
	}

	out = reflectClusterToStack(cluster, appGroups)
	return
}

func (u *StackUsecase) Fetch(ctx context.Context, organizationId string) (out []domain.Stack, err error) {
	organization, err := u.organizationRepo.Get(organizationId)
	if err != nil {
		return out, httpErrors.NewInternalServerError(errors.Wrap(err, fmt.Sprintf("Failed to get organization for clusterId %s", organizationId)), "S_FAILED_FETCH_ORGANIZATION", "")
	}

	clusters, err := u.clusterRepo.FetchByOrganizationId(organizationId)
	if err != nil {
		return out, err
	}

	for _, cluster := range clusters {
		appGroups, err := u.appGroupRepo.Fetch(cluster.ID)
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
	organizations, err := u.organizationRepo.Fetch()
	if err != nil {
		return errors.Wrap(err, "Failed to get organizations")
	}

	for _, organization := range *organizations {
		if organization.PrimaryClusterId == cluster.ID.String() {

			clusters, err := u.clusterRepo.FetchByOrganizationId(organization.ID)
			if err != nil {
				return errors.Wrap(err, "Failed to get organizations")
			}

			for _, cl := range clusters {
				if cl.ID != cluster.ID &&
					cl.Status != domain.ClusterStatus_DELETED &&
					cl.Status != domain.ClusterStatus_ERROR {
					return httpErrors.NewBadRequestError(fmt.Errorf("Failed to delete 'Primary' cluster. The clusters remain in organization"), "S_REMAIN_CLUSTER_FOR_DELETION", "")
				}
			}
			break
		}
	}

	appGroups, err := u.appGroupRepo.Fetch(domain.ClusterId(dto.ID))
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

	workflow := ""
	if strings.Contains(cluster.StackTemplate.Template, "aws-reference") || strings.Contains(cluster.StackTemplate.Template, "eks-reference") {
		workflow = "tks-stack-delete-aws"
	} else if strings.Contains(cluster.StackTemplate.Template, "aws-msa-reference") || strings.Contains(cluster.StackTemplate.Template, "eks-msa-reference") {
		workflow = "tks-stack-delete-aws-msa"
	} else {
		log.ErrorWithContext(ctx, "Invalid template  : ", cluster.StackTemplate.Template)
		return httpErrors.NewInternalServerError(fmt.Errorf("Invalid stack-template %s", cluster.StackTemplate.Template), "", "")
	}

	workflowId, err := u.argo.SumbitWorkflowFromWftpl(workflow, argowf.SubmitOptions{
		Parameters: []string{
			fmt.Sprintf("tks_api_url=%s", viper.GetString("external-address")),
			"organization_id=" + dto.OrganizationId,
			"cluster_id=" + dto.ID.String(),
			"cloud_account_id=" + cluster.CloudAccount.ID.String(),
		},
	})
	if err != nil {
		log.ErrorWithContext(ctx, err)
		return err
	}
	log.DebugWithContext(ctx, "Submitted workflow: ", workflowId)

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

	appGroups, err := u.appGroupRepo.Fetch(domain.ClusterId(stackId))
	for _, appGroup := range appGroups {
		for i, step := range out {
			if step.Stage == appGroup.AppGroupType.String() {
				step := parseStatusDescription(appGroup.StatusDesc)

				out[i].Status = appGroup.Status.String()
				out[i].MaxStep = domain.MAX_STEP_LMA_CREATE_MEMBER
				if organization.PrimaryClusterId == cluster.ID.String() {
					out[i].MaxStep = domain.MAX_STEP_LMA_CREATE_PRIMARY
				}
				if appGroup.Status == domain.AppGroupStatus_DELETING {
					out[i].MaxStep = domain.MAX_STEP_LMA_REMOVE
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

func reflectClusterToStack(cluster domain.Cluster, appGroups []domain.AppGroup) domain.Stack {
	status, statusDesc := getStackStatus(cluster, appGroups)
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
			CpNodeCnt:   cluster.Conf.CpNodeCnt,
			TksNodeCnt:  cluster.Conf.TksNodeCnt,
			UserNodeCnt: cluster.Conf.UserNodeCnt,
		},
	}
}

// [TODO] more pretty
func getStackStatus(cluster domain.Cluster, appGroups []domain.AppGroup) (domain.StackStatus, string) {
	for _, appGroup := range appGroups {
		if appGroup.Status == domain.AppGroupStatus_INSTALLING {
			return domain.StackStatus_APPGROUP_INSTALLING, appGroup.StatusDesc
		}
		if appGroup.Status == domain.AppGroupStatus_DELETING {
			return domain.StackStatus_APPGROUP_DELETING, appGroup.StatusDesc
		}
		if appGroup.Status == domain.AppGroupStatus_ERROR {
			return domain.StackStatus_APPGROUP_ERROR, appGroup.StatusDesc
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
	if cluster.Status == domain.ClusterStatus_ERROR {
		return domain.StackStatus_CLUSTER_ERROR, cluster.StatusDesc
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
