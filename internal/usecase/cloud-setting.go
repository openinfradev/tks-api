package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/auth/request"
	"github.com/openinfradev/tks-api/internal/repository"
	argowf "github.com/openinfradev/tks-api/pkg/argo-client"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type ICloudSettingUsecase interface {
	Get(cloudSettingId uuid.UUID) (domain.CloudSetting, error)
	GetByName(organizationId string, name string) (domain.CloudSetting, error)
	Fetch(organizationId string) ([]domain.CloudSetting, error)
	Create(ctx context.Context, dto domain.CloudSetting) (cloudSettingId uuid.UUID, err error)
	Update(ctx context.Context, dto domain.CloudSetting) error
	Delete(ctx context.Context, dto domain.CloudSetting) error
}

type CloudSettingUsecase struct {
	repo        repository.ICloudSettingRepository
	clusterRepo repository.IClusterRepository
	argo        argowf.ArgoClient
}

func NewCloudSettingUsecase(r repository.Repository, argoClient argowf.ArgoClient) ICloudSettingUsecase {
	return &CloudSettingUsecase{
		repo:        r.CloudSetting,
		clusterRepo: r.Cluster,
		argo:        argoClient,
	}
}

func (u *CloudSettingUsecase) Create(ctx context.Context, dto domain.CloudSetting) (cloudSettingId uuid.UUID, err error) {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return uuid.Nil, httpErrors.NewBadRequestError(fmt.Errorf("Invalid token"))
	}

	dto.Resource = "TODO server result or additional information"
	dto.CreatorId = user.GetUserId()

	_, err = u.GetByName(dto.OrganizationId, dto.Name)
	if err == nil {
		return uuid.Nil, httpErrors.NewBadRequestError(httpErrors.DuplicateResource)
	}

	cloudSettingId, err = u.repo.Create(dto)
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

func (u *CloudSettingUsecase) Update(ctx context.Context, dto domain.CloudSetting) error {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return httpErrors.NewBadRequestError(fmt.Errorf("Invalid token"))
	}

	dto.Resource = "TODO server result or additional information"
	dto.UpdatorId = user.GetUserId()
	err := u.repo.Update(dto)
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

func (u *CloudSettingUsecase) GetByName(organizationId string, name string) (res domain.CloudSetting, err error) {
	res, err = u.repo.GetByName(organizationId, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.CloudSetting{}, httpErrors.NewNotFoundError(err)
		}
		return domain.CloudSetting{}, err
	}
	return
}

func (u *CloudSettingUsecase) Fetch(organizationId string) (res []domain.CloudSetting, err error) {
	res, err = u.repo.Fetch(organizationId)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (u *CloudSettingUsecase) Delete(ctx context.Context, dto domain.CloudSetting) (err error) {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return httpErrors.NewBadRequestError(fmt.Errorf("Invalid token"))
	}

	_, err = u.Get(dto.ID)
	if err != nil {
		return httpErrors.NewNotFoundError(err)
	}

	dto.UpdatorId = user.GetUserId()

	err = u.repo.Delete(dto)
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
		if cluster.Status != domain.ClusterStatus_DELETED {
			cnt++
		}
	}

	return cnt
}
