package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/repository"
	argowf "github.com/openinfradev/tks-api/pkg/argo-client"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type IAppGroupUsecase interface {
	Fetch(ctx context.Context, clusterId domain.ClusterId, pg *pagination.Pagination) ([]domain.AppGroup, error)
	Create(ctx context.Context, dto domain.AppGroup) (id domain.AppGroupId, err error)
	Get(ctx context.Context, id domain.AppGroupId) (out domain.AppGroup, err error)
	Delete(ctx context.Context, id domain.AppGroupId) (err error)
	GetApplications(ctx context.Context, id domain.AppGroupId, applicationType domain.ApplicationType) (out []domain.Application, err error)
	UpdateApplication(ctx context.Context, dto domain.Application) (err error)
}

type AppGroupUsecase struct {
	repo             repository.IAppGroupRepository
	clusterRepo      repository.IClusterRepository
	cloudAccountRepo repository.ICloudAccountRepository
	argo             argowf.ArgoClient
}

func NewAppGroupUsecase(r repository.Repository, argoClient argowf.ArgoClient) IAppGroupUsecase {
	return &AppGroupUsecase{
		repo:             r.AppGroup,
		clusterRepo:      r.Cluster,
		cloudAccountRepo: r.CloudAccount,
		argo:             argoClient,
	}
}

func (u *AppGroupUsecase) Fetch(ctx context.Context, clusterId domain.ClusterId, pg *pagination.Pagination) (out []domain.AppGroup, err error) {
	out, err = u.repo.Fetch(clusterId, pg)
	if err != nil {
		return nil, err
	}
	return
}

func (u *AppGroupUsecase) Create(ctx context.Context, dto domain.AppGroup) (id domain.AppGroupId, err error) {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return "", httpErrors.NewUnauthorizedError(fmt.Errorf("Invalid token"), "A_INVALID_TOKEN", "")
	}
	userId := user.GetUserId()
	dto.CreatorId = &userId

	cluster, err := u.clusterRepo.Get(dto.ClusterId)
	if err != nil {
		return "", httpErrors.NewBadRequestError(err, "AG_NOT_FOUND_CLUSTER", "")
	}

	resAppGroups, err := u.repo.Fetch(dto.ClusterId, nil)
	if err != nil {
		return "", httpErrors.NewBadRequestError(err, "AG_NOT_FOUND_APPGROUP", "")
	}

	for _, resAppGroup := range resAppGroups {
		if resAppGroup.AppGroupType == dto.AppGroupType {
			if resAppGroup.Status == domain.AppGroupStatus_INSTALLING ||
				resAppGroup.Status == domain.AppGroupStatus_DELETING {
				return "", fmt.Errorf("In progress appgroup status [%s]", resAppGroup.Status.String())
			}
			dto.ID = resAppGroup.ID
		}
	}

	// check cloudAccount
	cloudAccounts, err := u.cloudAccountRepo.Fetch(cluster.OrganizationId, nil)
	if err != nil {
		return "", httpErrors.NewBadRequestError(fmt.Errorf("Failed to get cloudAccounts"), "", "")
	}
	tksCloudAccountId := cluster.CloudAccountId.String()
	isExist := false
	for _, ca := range cloudAccounts {
		if ca.ID == cluster.CloudAccountId {

			// FOR TEST. ADD MAGIC KEYWORD
			if strings.Contains(ca.Name, domain.CLOUD_ACCOUNT_INCLUSTER) {
				tksCloudAccountId = ""
			}
			isExist = true
			break
		}
	}
	if !isExist {
		return "", httpErrors.NewBadRequestError(fmt.Errorf("Not found cloudAccountId[%s] in organization[%s]", cluster.CloudAccountId, cluster.OrganizationId), "", "")
	}

	if dto.ID == "" {
		dto.ID, err = u.repo.Create(dto)
	} else {
		err = u.repo.Update(dto)
	}
	if err != nil {
		return "", httpErrors.NewInternalServerError(err, "AG_FAILED_TO_CREATE_APPGROUP", "")
	}

	workflowTemplate := ""
	opts := argowf.SubmitOptions{}
	opts.Parameters = []string{
		"organization_id=" + cluster.OrganizationId,
		"site_name=" + dto.ClusterId.String(),
		"cluster_id=" + dto.ClusterId.String(),
		"github_account=" + viper.GetString("git-account"),
		"manifest_repo_url=" + viper.GetString("git-base-url") + "/" + viper.GetString("git-account") + "/" + dto.ClusterId.String() + "-manifests",
		"base_repo_branch=" + viper.GetString("revision"),
		"app_group_id=" + dto.ID.String(),
		"keycloak_url=" + strings.TrimSuffix(viper.GetString("keycloak-address"), "/auth"),
		"console_url=" + viper.GetString("console-address"),
		"alert_tks=" + viper.GetString("external-address") + "/system-api/1.0/alerts",
		"alert_slack=" + viper.GetString("alert-slack"),
		"cloud_account_id=" + tksCloudAccountId,
	}

	switch dto.AppGroupType {
	case domain.AppGroupType_LMA:
		workflowTemplate = "tks-lma-federation"
		opts.Parameters = append(opts.Parameters, "logging_component=loki")

	case domain.AppGroupType_SERVICE_MESH:
		workflowTemplate = "tks-service-mesh"

	default:
		log.ErrorWithContext(ctx, "invalid appGroup type ", dto.AppGroupType.String())
		return "", errors.Wrap(err, fmt.Sprintf("Invalid appGroup type. %s", dto.AppGroupType.String()))
	}

	workflowId, err := u.argo.SumbitWorkflowFromWftpl(workflowTemplate, opts)
	if err != nil {
		log.ErrorWithContext(ctx, "failed to submit argo workflow template. err : ", err)
		return "", httpErrors.NewInternalServerError(err, "AG_FAILED_TO_CALL_WORKFLOW", "")
	}

	if err := u.repo.InitWorkflow(dto.ID, workflowId, domain.AppGroupStatus_INSTALLING); err != nil {
		return "", errors.Wrap(err, "Failed to initialize appGroup status")
	}

	return dto.ID, nil
}

func (u *AppGroupUsecase) Get(ctx context.Context, id domain.AppGroupId) (out domain.AppGroup, err error) {
	appGroup, err := u.repo.Get(id)
	if err != nil {
		return domain.AppGroup{}, err
	}
	return appGroup, nil
}

func (u *AppGroupUsecase) Delete(ctx context.Context, id domain.AppGroupId) (err error) {
	appGroup, err := u.repo.Get(id)
	if err != nil {
		return fmt.Errorf("No appGroup for deletiing : %s", id)
	}
	cluster, err := u.clusterRepo.Get(appGroup.ClusterId)
	if err != nil {
		return httpErrors.NewBadRequestError(err, "AG_NOT_FOUND_CLUSTER", "")
	}
	organizationId := cluster.OrganizationId

	// Call argo workflow template
	workflowTemplate := ""
	appGroupName := ""

	switch appGroup.AppGroupType {
	case domain.AppGroupType_LMA:
		workflowTemplate = "tks-remove-lma-federation"
		appGroupName = "lma"

	case domain.AppGroupType_SERVICE_MESH:
		workflowTemplate = "tks-remove-servicemesh"
		appGroupName = "service-mesh"

	default:
		return fmt.Errorf("Invalid appGroup type %s", appGroup.AppGroupType)
	}

	opts := argowf.SubmitOptions{}
	opts.Parameters = []string{
		"organization_id=" + organizationId,
		"app_group=" + appGroupName,
		"github_account=" + viper.GetString("git-account"),
		"cluster_id=" + cluster.ID.String(),
		"app_group_id=" + id.String(),
		"keycloak_url=" + strings.TrimSuffix(viper.GetString("keycloak-address"), "/auth"),
		"base_repo_branch=" + viper.GetString("revision"),
	}

	workflowId, err := u.argo.SumbitWorkflowFromWftpl(workflowTemplate, opts)
	if err != nil {
		return fmt.Errorf("Failed to call argo workflow : %s", err)
	}

	log.DebugWithContext(ctx, "submited workflow name : ", workflowId)

	if err := u.repo.InitWorkflow(id, workflowId, domain.AppGroupStatus_DELETING); err != nil {
		return fmt.Errorf("Failed to initialize appGroup status. err : %s", err)
	}

	/*
		err = u.userRepository.Delete(appGroupId)
		if err != nil {
			return fmt.Errorf("Fatiled to deleting appGroup : %s", appGroupId)
		}
	*/

	return nil
}

func (u *AppGroupUsecase) GetApplications(ctx context.Context, id domain.AppGroupId, applicationType domain.ApplicationType) (out []domain.Application, err error) {
	out, err = u.repo.GetApplications(id, applicationType)
	if err != nil {
		return nil, err
	}
	return
}

func (u *AppGroupUsecase) UpdateApplication(ctx context.Context, dto domain.Application) (err error) {
	err = u.repo.UpsertApplication(dto)
	if err != nil {
		return err
	}
	return nil
}
