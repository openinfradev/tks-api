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
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type IStackUsecase interface {
	Get(stackId uuid.UUID) (domain.Stack, error)
	GetByName(organizationId string, name string) (domain.Stack, error)
	Fetch(organizationId string) ([]domain.Stack, error)
	Create(ctx context.Context, dto domain.Stack) (stackId uuid.UUID, err error)
	Update(ctx context.Context, dto domain.Stack) error
	Delete(ctx context.Context, dto domain.Stack) error
}

type StackUsecase struct {
	clusterRepo repository.IClusterRepository
	argo        argowf.ArgoClient
}

func NewStackUsecase(cr repository.IClusterRepository, argoClient argowf.ArgoClient) IStackUsecase {
	return &StackUsecase{
		clusterRepo: cr,
		argo:        argoClient,
	}
}

func (u *StackUsecase) Create(ctx context.Context, dto domain.Stack) (stackId uuid.UUID, err error) {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return uuid.Nil, httpErrors.NewBadRequestError(fmt.Errorf("Invalid token"))
	}

	/***************************
	 * Pre-process cluster conf *
	 ***************************/
	clConf, err := u.constructClusterConf(&domain.ClusterConf{
		Region:          dto.Conf.Region,
		NumOfAz:         dto.Conf.NumOfAz,
		SshKeyName:      "",
		MachineType:     dto.Conf.MachineType,
		MachineReplicas: dto.Conf.MachineReplicas,
	},
	)
	if err != nil {
		return uuid.Nil, err
	}

	userId := user.GetUserId()
	dto.CreatorId = &userId
	dto.Conf = *clConf

	clusterId, err = u.repo.Create(dto)
	if err != nil {
		return "", errors.Wrap(err, "Failed to create cluster")
	}

	// Call argo workflow
	workflowId, err := u.argo.SumbitWorkflowFromWftpl(
		"create-tks-usercluster",
		argowf.SubmitOptions{
			Parameters: []string{
				"contract_id=" + dto.OrganizationId,
				"cluster_id=" + clusterId.String(),
				"site_name=" + clusterId.String(),
				"template_name=" + dto.TemplateId,
				"git_account=" + viper.GetString("git-account"),
				//"manifest_repo_url=" + viper.GetString("git-base-url") + "/" + viper.GetString("git-account") + "/" + clusterId + "-manifests",
			},
		})
	if err != nil {
		log.Error("failed to submit argo workflow template. err : ", err)
		return "", err
	}
	log.Info("Successfully submited workflow: ", workflowId)

	if err := u.repo.InitWorkflow(clusterId, workflowId, domain.ClusterStatus_INSTALLING); err != nil {
		return "", errors.Wrap(err, "Failed to initialize status")
	}

	return clusterId, nil
}

func (u *StackUsecase) Update(ctx context.Context, dto domain.Stack) error {
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

func (u *StackUsecase) Get(stackId uuid.UUID) (res domain.Stack, err error) {
	res, err = u.repo.Get(stackId)
	if err != nil {
		return domain.Stack{}, err
	}

	res.Clusters = u.getClusterCnt(stackId)

	return
}

func (u *StackUsecase) GetByName(organizationId string, name string) (res domain.Stack, err error) {
	res, err = u.repo.GetByName(organizationId, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Stack{}, httpErrors.NewNotFoundError(err)
		}
		return domain.Stack{}, err
	}
	return
}

func (u *StackUsecase) Fetch(organizationId string) (res []domain.Stack, err error) {
	res, err = u.repo.Fetch(organizationId)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (u *StackUsecase) Delete(ctx context.Context, dto domain.Stack) (err error) {
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

func (u *StackUsecase) getClusterCnt(stackId uuid.UUID) (cnt int) {
	cnt = 0

	clusters, err := u.clusterRepo.FetchByStackId(stackId)
	if err != nil {
		log.Error("Failed to get clusters by stackId. err : ", err)
		return cnt
	}

	for _, cluster := range clusters {
		if cluster.Status != domain.ClusterStatus_DELETED {
			cnt++
		}
	}

	return cnt
}
