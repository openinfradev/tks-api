package usecase

import (
	"context"
	"fmt"

	"github.com/openinfradev/tks-api/internal/middleware/auth/request"

	"github.com/openinfradev/tks-api/internal/repository"
	argowf "github.com/openinfradev/tks-api/pkg/argo-client"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type IAppGroupUsecase interface {
	Fetch(clusterId domain.ClusterId) ([]domain.AppGroup, error)
	Create(ctx context.Context, dto domain.AppGroup) (id domain.AppGroupId, err error)
	Get(id domain.AppGroupId) (out domain.AppGroup, err error)
	Delete(organizationId string, id domain.AppGroupId) (err error)
	GetApplications(id domain.AppGroupId, applicationType domain.ApplicationType) (out []domain.Application, err error)
	UpdateApplication(dto domain.Application) (err error)
}

type AppGroupUsecase struct {
	repo        repository.IAppGroupRepository
	clusterRepo repository.IClusterRepository
	argo        argowf.ArgoClient
}

func NewAppGroupUsecase(r repository.Repository, argoClient argowf.ArgoClient) IAppGroupUsecase {
	return &AppGroupUsecase{
		repo:        r.AppGroup,
		clusterRepo: r.Cluster,
		argo:        argoClient,
	}
}

func (u *AppGroupUsecase) Fetch(clusterId domain.ClusterId) (out []domain.AppGroup, err error) {
	out, err = u.repo.Fetch(clusterId)
	if err != nil {
		return nil, err
	}
	return
}

func (u *AppGroupUsecase) Create(ctx context.Context, dto domain.AppGroup) (id domain.AppGroupId, err error) {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return "", httpErrors.NewBadRequestError(fmt.Errorf("Invalid token"), "")
	}
	userId := user.GetUserId()
	dto.CreatorId = &userId

	cluster, err := u.clusterRepo.Get(dto.ClusterId)
	if err != nil {
		return "", errors.Wrap(err, "Failed to get cluster info")
	}

	resAppGroups, err := u.repo.Fetch(dto.ClusterId)
	if err != nil {
		return "", errors.Wrap(err, "Failed to get appgroups")
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

	if dto.ID == "" {
		dto.ID, err = u.repo.Create(dto)
	} else {
		err = u.repo.Update(dto)
	}
	if err != nil {
		return "", errors.Wrap(err, "Failed to create appGroup.")
	}

	workflowTemplate := ""
	opts := argowf.SubmitOptions{}
	opts.Parameters = []string{
		"organization_id=" + cluster.OrganizationId,
		"site_name=" + dto.ClusterId.String(),
		"cluster_id=" + dto.ClusterId.String(),
		"github_account=" + viper.GetString("git-account"),
		"manifest_repo_url=" + viper.GetString("git-base-url") + "/" + viper.GetString("git-account") + "/" + dto.ClusterId.String() + "-manifests",
		"revision=" + viper.GetString("revision"),
		"app_group_id=" + dto.ID.String(),
	}

	switch dto.AppGroupType {
	case domain.AppGroupType_LMA:
		workflowTemplate = "tks-lma-federation"
		opts.Parameters = append(opts.Parameters, "logging_component=loki")

	case domain.AppGroupType_SERVICE_MESH:
		workflowTemplate = "tks-service-mesh"

	default:
		log.Error("invalid appGroup type ", dto.AppGroupType.String())
		return "", errors.Wrap(err, fmt.Sprintf("Invalid appGroup type. %s", dto.AppGroupType.String()))
	}

	workflowId, err := u.argo.SumbitWorkflowFromWftpl(workflowTemplate, opts)
	if err != nil {
		log.Error("failed to submit argo workflow template. err : ", err)
		return "", errors.Wrap(err, "Failed to call argo workflow")
	}

	if err := u.repo.InitWorkflow(dto.ID, workflowId, domain.AppGroupStatus_INSTALLING); err != nil {
		return "", errors.Wrap(err, "Failed to initialize appGroup status")
	}

	return dto.ID, nil
}

func (u *AppGroupUsecase) Get(id domain.AppGroupId) (out domain.AppGroup, err error) {
	appGroup, err := u.repo.Get(id)
	if err != nil {
		return domain.AppGroup{}, err
	}
	return appGroup, nil
}

func (u *AppGroupUsecase) Delete(organizationId string, id domain.AppGroupId) (err error) {
	appGroup, err := u.repo.Get(id)
	if err != nil {
		return fmt.Errorf("No appGroup for deletiing : %s", id)
	}

	clusterId := appGroup.ClusterId

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
		"cluster_id=" + clusterId.String(),
		"app_group_id=" + id.String(),
	}

	workflowId, err := u.argo.SumbitWorkflowFromWftpl(workflowTemplate, opts)
	if err != nil {
		return fmt.Errorf("Failed to call argo workflow : %s", err)
	}

	log.Debug("submited workflow name : ", workflowId)

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

func (u *AppGroupUsecase) GetApplications(id domain.AppGroupId, applicationType domain.ApplicationType) (out []domain.Application, err error) {
	out, err = u.repo.GetApplications(id, applicationType)
	if err != nil {
		return nil, err
	}
	return
}

func (u *AppGroupUsecase) UpdateApplication(dto domain.Application) (err error) {
	err = u.repo.UpsertApplication(dto)
	if err != nil {
		return err
	}
	return nil
}
