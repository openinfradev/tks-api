package usecase

import (
	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/repository"
	argowf "github.com/openinfradev/tks-api/pkg/argo-client"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
)

type ICloudSettingUsecase interface {
	Get(cloudSettingId uuid.UUID) (domain.CloudSetting, error)
	Fetch(organizationId string) ([]domain.CloudSetting, error)
	Create(organizationId string, in domain.CreateCloudSettingRequest, creator uuid.UUID) (cloudSettingId uuid.UUID, err error)
	Update(cloudSettingId uuid.UUID, in domain.UpdateCloudSettingRequest, updator uuid.UUID) (err error)
	Delete(cloudSettingId uuid.UUID) error
}

type CloudSettingUsecase struct {
	repo        repository.ICloudSettingRepository
	clusterRepo repository.IClusterRepository
	argo        argowf.ArgoClient
}

func NewCloudSettingUsecase(r repository.ICloudSettingRepository, cr repository.IClusterRepository, argoClient argowf.ArgoClient) ICloudSettingUsecase {
	return &CloudSettingUsecase{
		repo:        r,
		clusterRepo: cr,
		argo:        argoClient,
	}
}

func (u *CloudSettingUsecase) Create(organizationId string, in domain.CreateCloudSettingRequest, creator uuid.UUID) (cloudSettingId uuid.UUID, err error) {
	resource := "TODO server result or additional information"
	cloudSettingId, err = u.repo.Create(organizationId, in, resource, creator)
	if err != nil {
		return uuid.Nil, httpErrors.NewInternalServerError(err)
	}
	log.Info("newly created CloudSetting ID:", cloudSettingId)

	/*
		workflowId, err := u.argo.SumbitWorkflowFromWftpl(
			"tks-create-contract-repo",
			argowf.SubmitOptions{
				Parameters: []string{
					"contract_id=" + cloudSettingId,
				},
			})
		if err != nil {
			log.Error("failed to submit argo workflow template. err : ", err)
			return "", fmt.Errorf("Failed to call argo workflow : %s", err)
		}
		log.Info("submited workflow :", workflowId)
		if err := u.repo.InitWorkflow(cloudSettingId, workflowId); err != nil {
			return "", fmt.Errorf("Failed to initialize cloudSetting status to 'CREATING'. err : %s", err)
		}
	*/

	return cloudSettingId, nil
}

func (u *CloudSettingUsecase) Update(cloudSettingId uuid.UUID, in domain.UpdateCloudSettingRequest, updator uuid.UUID) (err error) {
	resource := "TODO server result or additional information"
	err = u.repo.Update(cloudSettingId, in, resource, updator)
	if err != nil {
		return httpErrors.NewInternalServerError(err)
	}
	return nil
}

func (u *CloudSettingUsecase) Get(cloudSettingId uuid.UUID) (res domain.CloudSetting, err error) {
	res, err = u.repo.Get(cloudSettingId)
	if err != nil {
		return domain.CloudSetting{}, err
	}

	res.Clusters = u.getClusterCnt(cloudSettingId)

	return
}

func (u *CloudSettingUsecase) Fetch(organizationId string) (res []domain.CloudSetting, err error) {
	res, err = u.repo.Fetch(organizationId)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (u *CloudSettingUsecase) Delete(cloudSettingId uuid.UUID) (err error) {
	_, err = u.Get(cloudSettingId)
	if err != nil {
		return httpErrors.NewNotFoundError(err)
	}

	err = u.repo.Delete(cloudSettingId)
	if err != nil {
		return err
	}

	return nil
}

func (u *CloudSettingUsecase) getClusterCnt(cloudSettingId uuid.UUID) (cnt int) {
	cnt = 0

	clusters, err := u.clusterRepo.FetchByCloudSettingId(cloudSettingId)
	if err != nil {
		log.Error("Failed to get clusters by cloudSettingId. err : ", err)
		return cnt
	}

	for _, cluster := range clusters {
		if cluster.Status != "DELETED" {
			cnt++
		}
	}

	return cnt
}
