package usecase

import (
	"context"
	"fmt"

	"github.com/openinfradev/tks-api/internal/middleware/auth/request"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/repository"
	argowf "github.com/openinfradev/tks-api/pkg/argo-client"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

const MAX_WORKFLOW_TIME = 30

type ICloudAccountUsecase interface {
	Get(cloudAccountId uuid.UUID) (domain.CloudAccount, error)
	GetByName(organizationId string, name string) (domain.CloudAccount, error)
	Fetch(organizationId string) ([]domain.CloudAccount, error)
	Create(ctx context.Context, dto domain.CloudAccount) (cloudAccountId uuid.UUID, err error)
	Update(ctx context.Context, dto domain.CloudAccount) error
	Delete(ctx context.Context, dto domain.CloudAccount) error
	DeleteForce(ctx context.Context, cloudAccountId uuid.UUID) error
}

type CloudAccountUsecase struct {
	repo        repository.ICloudAccountRepository
	clusterRepo repository.IClusterRepository
	argo        argowf.ArgoClient
}

func NewCloudAccountUsecase(r repository.Repository, argoClient argowf.ArgoClient) ICloudAccountUsecase {
	return &CloudAccountUsecase{
		repo:        r.CloudAccount,
		clusterRepo: r.Cluster,
		argo:        argoClient,
	}
}

func (u *CloudAccountUsecase) Create(ctx context.Context, dto domain.CloudAccount) (cloudAccountId uuid.UUID, err error) {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return uuid.Nil, httpErrors.NewBadRequestError(fmt.Errorf("Invalid token"), "", "")
	}

	dto.Resource = "TODO server result or additional information"
	dto.CreatorId = user.GetUserId()

	_, err = u.GetByName(dto.OrganizationId, dto.Name)
	if err == nil {
		return uuid.Nil, httpErrors.NewBadRequestError(httpErrors.DuplicateResource, "", "")
	}

	cloudAccountId, err = u.repo.Create(dto)
	if err != nil {
		return uuid.Nil, httpErrors.NewInternalServerError(err, "", "")
	}
	log.Info("newly created CloudAccount ID:", cloudAccountId)

	workflowId, err := u.argo.SumbitWorkflowFromWftpl(
		"tks-create-aws-cloud-account",
		argowf.SubmitOptions{
			Parameters: []string{
				"aws_region=" + "ap-northeast-2",
				"tks_cloud_account_id=" + cloudAccountId.String(),
				"aws_account_id=" + dto.AwsAccountId,
				"aws_access_key_id=" + dto.AccessKeyId,
				"aws_secret_access_key=" + dto.SecretAccessKey,
				"aws_session_token=" + dto.SessionToken,
			},
		})
	if err != nil {
		log.Error("failed to submit argo workflow template. err : ", err)
		return uuid.Nil, fmt.Errorf("Failed to call argo workflow : %s", err)
	}
	log.Info("submited workflow :", workflowId)

	if err := u.repo.InitWorkflow(cloudAccountId, workflowId, domain.CloudAccountStatus_CREATING); err != nil {
		return uuid.Nil, errors.Wrap(err, "Failed to initialize status")
	}

	return cloudAccountId, nil
}

func (u *CloudAccountUsecase) Update(ctx context.Context, dto domain.CloudAccount) error {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return httpErrors.NewBadRequestError(fmt.Errorf("Invalid token"), "", "")
	}

	dto.Resource = "TODO server result or additional information"
	dto.UpdatorId = user.GetUserId()
	err := u.repo.Update(dto)
	if err != nil {
		return httpErrors.NewInternalServerError(err, "", "")
	}
	return nil
}

func (u *CloudAccountUsecase) Get(cloudAccountId uuid.UUID) (res domain.CloudAccount, err error) {
	res, err = u.repo.Get(cloudAccountId)
	if err != nil {
		return domain.CloudAccount{}, err
	}

	res.Clusters = u.getClusterCnt(cloudAccountId)

	return
}

func (u *CloudAccountUsecase) GetByName(organizationId string, name string) (res domain.CloudAccount, err error) {
	res, err = u.repo.GetByName(organizationId, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.CloudAccount{}, httpErrors.NewNotFoundError(err, "", "")
		}
		return domain.CloudAccount{}, err
	}
	res.Clusters = u.getClusterCnt(res.ID)
	return
}

func (u *CloudAccountUsecase) Fetch(organizationId string) (cloudAccounts []domain.CloudAccount, err error) {
	cloudAccounts, err = u.repo.Fetch(organizationId)
	if err != nil {
		return nil, err
	}

	for i, cloudAccount := range cloudAccounts {
		cloudAccounts[i].Clusters = u.getClusterCnt(cloudAccount.ID)
	}
	return
}

func (u *CloudAccountUsecase) Delete(ctx context.Context, dto domain.CloudAccount) (err error) {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return httpErrors.NewBadRequestError(fmt.Errorf("Invalid token"), "", "")
	}

	_, err = u.Get(dto.ID)
	if err != nil {
		return httpErrors.NewNotFoundError(err, "", "")
	}
	dto.UpdatorId = user.GetUserId()

	if u.getClusterCnt(dto.ID) > 0 {
		return fmt.Errorf("사용 중인 클러스터가 있어 삭제할 수 없습니다.")
	}

	workflowId, err := u.argo.SumbitWorkflowFromWftpl(
		"tks-delete-aws-cloud-account",
		argowf.SubmitOptions{
			Parameters: []string{
				"aws_region=" + "ap-northeast-2",
				"tks_cloud_account_id=" + dto.ID.String(),
				"aws_account_id=" + dto.AwsAccountId,
				"aws_access_key_id=" + dto.AccessKeyId,
				"aws_secret_access_key=" + dto.SecretAccessKey,
				"aws_session_token=" + dto.SessionToken,
			},
		})
	if err != nil {
		log.Error("failed to submit argo workflow template. err : ", err)
		return fmt.Errorf("Failed to call argo workflow : %s", err)
	}
	log.Info("submited workflow :", workflowId)

	if err := u.repo.InitWorkflow(dto.ID, workflowId, domain.CloudAccountStatus_DELETING); err != nil {
		return errors.Wrap(err, "Failed to initialize status")
	}

	return nil
}

func (u *CloudAccountUsecase) DeleteForce(ctx context.Context, cloudAccountId uuid.UUID) (err error) {
	if u.getClusterCnt(cloudAccountId) > 0 {
		return fmt.Errorf("사용 중인 클러스터가 있어 삭제할 수 없습니다.")
	}

	err = u.repo.Delete(cloudAccountId)
	if err != nil {
		return err
	}

	return nil
}

func (u *CloudAccountUsecase) getClusterCnt(cloudAccountId uuid.UUID) (cnt int) {
	cnt = 0

	clusters, err := u.clusterRepo.FetchByCloudAccountId(cloudAccountId)
	if err != nil {
		log.Error("Failed to get clusters by cloudAccountId. err : ", err)
		return cnt
	}

	for _, cluster := range clusters {
		if cluster.Status != domain.ClusterStatus_DELETED {
			cnt++
		}
	}

	return cnt
}
