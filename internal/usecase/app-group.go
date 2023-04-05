package usecase

import (
	"fmt"

	"github.com/openinfradev/tks-api/internal/repository"
	argowf "github.com/openinfradev/tks-api/pkg/argo-client"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/spf13/viper"
)

type IAppGroupUsecase interface {
	Fetch(clusterId domain.ClusterId) ([]domain.AppGroup, error)
	Create(dto domain.AppGroup) (id domain.AppGroupId, err error)
	Get(id domain.AppGroupId) (out domain.AppGroup, err error)
	Delete(id domain.AppGroupId) (err error)
	GetApplications(id domain.AppGroupId, applicationType domain.ApplicationType) (out []domain.Application, err error)
	UpdateApplication(dto domain.Application) (err error)
}

type AppGroupUsecase struct {
	repo        repository.IAppGroupRepository
	clusterRepo repository.IClusterRepository
	argo        argowf.ArgoClient
}

func NewAppGroupUsecase(r repository.IAppGroupRepository, clusterRepo repository.IClusterRepository, argoClient argowf.ArgoClient) IAppGroupUsecase {
	return &AppGroupUsecase{
		repo:        r,
		clusterRepo: clusterRepo,
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

func (u *AppGroupUsecase) Create(dto domain.AppGroup) (id domain.AppGroupId, err error) {
	// Check Cluster
	_, err = u.clusterRepo.Get(dto.ClusterId)
	if err != nil {
		return "", fmt.Errorf("Failed to get cluster info err %s", err)
	}

	resAppGroups, err := u.repo.Fetch(dto.ClusterId)
	if err != nil {
		return "", fmt.Errorf("Failed to get appgroup info err %s", err)
	}

	for _, resAppGroup := range resAppGroups {
		if resAppGroup.Name == dto.Name &&
			resAppGroup.AppGroupType == dto.AppGroupType {
			dto.ID = resAppGroup.ID
			break
		}
	}

	dto.ID, err = u.repo.Create(dto)
	if err != nil {
		return "", fmt.Errorf("Failed to create appGroup. err %s", err)
	}

	workflowTemplate := ""
	opts := argowf.SubmitOptions{}
	opts.Parameters = []string{
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
		log.Error("invalid appGroup type ", dto.AppGroupType)
		return "", fmt.Errorf("Invalid appGroup type. err %s", dto.AppGroupType)
	}

	workflowId, err := u.argo.SumbitWorkflowFromWftpl(workflowTemplate, opts)
	if err != nil {
		log.Error("failed to submit argo workflow template. err : ", err)
		return "", fmt.Errorf("Failed to call argo workflow : %s", err)
	}

	log.Debug("submited workflow name : ", workflowId)

	if err := u.repo.InitWorkflow(dto.ID, workflowId, domain.AppGroupStatus_INSTALLING); err != nil {
		return "", fmt.Errorf("Failed to initialize appGroup status. err : %s", err)
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

func (u *AppGroupUsecase) Delete(id domain.AppGroupId) (err error) {
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
		err = u.repo.Delete(appGroupId)
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
