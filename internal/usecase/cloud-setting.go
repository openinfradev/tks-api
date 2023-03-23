package usecase

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/repository"
	argowf "github.com/openinfradev/tks-api/pkg/argo-client"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
)

type ICloudSettingUsecase interface {
	Get(cloudSettingId uuid.UUID) (domain.CloudSetting, error)
	GetByOrganizationId(organizationId string) (domain.CloudSetting, error)
	Create(organizationId string, in domain.CreateCloudSettingRequest, resource string, creator uuid.UUID) (cloudSettingId uuid.UUID, err error)
	Delete(cloudSettingId uuid.UUID) error
}

type CloudSettingUsecase struct {
	repo repository.ICloudSettingRepository
	argo argowf.ArgoClient
}

func NewCloudSettingUsecase(r repository.ICloudSettingRepository, argoClient argowf.ArgoClient) ICloudSettingUsecase {
	return &CloudSettingUsecase{
		repo: r,
		argo: argoClient,
	}
}

func (u *CloudSettingUsecase) Create(organizationId string, in domain.CreateCloudSettingRequest, resource string, creator uuid.UUID) (cloudSettingId uuid.UUID, err error) {
	_, err = u.repo.GetByOrganizationId(organizationId)
	if err == nil {
		return uuid.Nil, httpErrors.NewRestError(http.StatusForbidden, "", fmt.Errorf("Already exist cloudSetting for organization"))
	}

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

func (u *CloudSettingUsecase) Get(cloudSettingId uuid.UUID) (res domain.CloudSetting, err error) {
	res, err = u.repo.Get(cloudSettingId)
	if err != nil {
		return domain.CloudSetting{}, err
	}
	return res, nil
}

func (u *CloudSettingUsecase) GetByOrganizationId(organizationId string) (res domain.CloudSetting, err error) {
	res, err = u.repo.GetByOrganizationId(organizationId)
	if err != nil {
		return domain.CloudSetting{}, err
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
