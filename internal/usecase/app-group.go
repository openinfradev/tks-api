package usecase

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/domain"
	"github.com/openinfradev/tks-api/internal/repository"
	argowf "github.com/openinfradev/tks-api/pkg/argo-client"
	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/spf13/viper"
)

type IAppGroupUsecase interface {
	Fetch(clusterId string) ([]domain.AppGroup, error)
	Create(clusterId string, name string, appGroupType string, creatorId string, description string) (appGroupId string, err error)
	Get(appGroupId string) (out domain.AppGroup, err error)
	Delete(appGroupId string) (err error)
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

func (u *AppGroupUsecase) Fetch(clusterId string) (out []domain.AppGroup, err error) {
	out, err = u.repo.Fetch(clusterId)

	if err != nil {
		return nil, err
	}
	return out, nil
}

func (u *AppGroupUsecase) Create(clusterId string, name string, appGroupType string, creatorId string, description string) (appGroupId string, err error) {
	creator := uuid.Nil
	if creatorId != "" {
		creator, err = uuid.Parse(creatorId)
		if err != nil {
			return "", fmt.Errorf("Invalid Creator ID %s", creatorId)
		}
	}

	// Check Cluster
	_, err = u.clusterRepo.Get(clusterId)
	if err != nil {
		return "", fmt.Errorf("Failed to get cluster info err %s", err)
	}

	resAppGroups, err := u.repo.Fetch(clusterId)
	if err != nil {
		return "", fmt.Errorf("Failed to get appgroup info err %s", err)
	}

	for _, resAppGroup := range resAppGroups {
		if resAppGroup.Name == name &&
			resAppGroup.AppGroupType == appGroupType {
			appGroupId = resAppGroup.Id
			break
		}
	}

	appGroupId, err = u.repo.Create(clusterId, name, appGroupType, creator, description)
	if err != nil {
		return "", fmt.Errorf("Failed to create appGroup. err %s", err)
	}

	workflowTemplate := ""
	opts := argowf.SubmitOptions{}
	opts.Parameters = []string{
		"site_name=" + clusterId,
		"cluster_id=" + clusterId,
		"github_account=" + viper.GetString("git-account"),
		"manifest_repo_url=" + viper.GetString("git-base-url") + "/" + viper.GetString("git-account") + "/" + clusterId + "-manifests",
		"revision=" + viper.GetString("revision"),
		"app_group_id=" + appGroupId,
	}

	switch appGroupType {
	case "LMA":
		workflowTemplate = "tks-lma-federation"
		opts.Parameters = append(opts.Parameters, "logging_component=loki")

	case "LMA_EFK":
		workflowTemplate = "tks-lma-federation"
		opts.Parameters = append(opts.Parameters, "logging_component=efk")

	case "SERVICE_MESH":
		workflowTemplate = "tks-service-mesh"

	default:
		log.Error("invalid appGroup type ", appGroupType)
		return "", fmt.Errorf("Invalid appGroup type. err %s", appGroupType)
	}

	workflowId, err := u.argo.SumbitWorkflowFromWftpl(workflowTemplate, opts)
	if err != nil {
		log.Error("failed to submit argo workflow template. err : ", err)
		return "", fmt.Errorf("Failed to call argo workflow : %s", err)
	}

	log.Debug("submited workflow name : ", workflowId)

	if err := u.repo.UpdateAppGroupStatus(appGroupId, domain.AppGroupStatus_INSTALLING, workflowId); err != nil {
		return "", fmt.Errorf("Failed to update appGroup status to 'INSTALLING'. err : %s", err)
	}

	return appGroupId, nil
}

func (u *AppGroupUsecase) Get(appGroupId string) (out domain.AppGroup, err error) {
	appGroup, err := u.repo.Get(appGroupId)
	if err != nil {
		return domain.AppGroup{}, err
	}
	return appGroup, nil
}

func (u *AppGroupUsecase) Delete(appGroupId string) (err error) {
	appGroup, err := u.repo.Get(appGroupId)
	if err != nil {
		return fmt.Errorf("No appGroup for deletiing : %s", appGroupId)
	}

	clusterId := appGroup.ClusterId

	// Call argo workflow template
	workflowTemplate := ""
	appGroupName := ""

	switch appGroup.AppGroupType {
	case "LMA", "LMA_EFK":
		workflowTemplate = "tks-remove-lma-federation"
		appGroupName = "lma"

	case "SERVICE_MESH":
		workflowTemplate = "tks-remove-servicemesh"
		appGroupName = "service-mesh"

	default:
		return fmt.Errorf("Invalid appGroup type %s", appGroup.AppGroupType)
	}

	opts := argowf.SubmitOptions{}
	opts.Parameters = []string{
		"app_group=" + appGroupName,
		"github_account=" + viper.GetString("git-account"),
		"cluster_id=" + clusterId,
		"app_group_id=" + appGroupId,
	}

	workflowId, err := u.argo.SumbitWorkflowFromWftpl(workflowTemplate, opts)
	if err != nil {
		return fmt.Errorf("Failed to call argo workflow : %s", err)
	}

	log.Debug("submited workflow name : ", workflowId)

	if err := u.repo.UpdateAppGroupStatus(appGroupId, domain.AppGroupStatus_DELETING, workflowId); err != nil {
		return fmt.Errorf("Failed to update appGroup status to 'DELETING'. err : %s", err)
	}

	err = u.repo.Delete(appGroupId)
	if err != nil {
		return fmt.Errorf("Fatiled to deleting appGroup : %s", appGroupId)
	}

	return nil
}
