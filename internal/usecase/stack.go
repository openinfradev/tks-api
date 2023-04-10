package usecase

import (
	"context"
	"fmt"
	"net/http"
	"time"

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

	workflow := ""
	if input.TemplateId == "aws-reference" || input.TemplateId == "eks-reference" {
		workflow = "tks-stack-create-aws"
	} else if input.TemplateId == "aws-msa-reference" || input.TemplateId == "eks-msa-reference" {
		workflow = "tks-stack-create-aws-msa"
	} else {
		log.Error("Invalid templateId  : ", input.TemplateId)
		ErrorJSON(w, err.Error(), http.StatusBadRequest)
		return
	}

	{
		nameSpace := "argo"
		opts := argowf.SubmitOptions{}
		opts.Parameters = []string{
			fmt.Sprintf("tks_info_url=%s:%d", viper.GetString("info-address"), viper.GetInt("info-port")),
			fmt.Sprintf("tks_contract_url=%s:%d", viper.GetString("contract-address"), viper.GetInt("contract-port")),
			fmt.Sprintf("tks_cluster_lcm_url=%s:%d", viper.GetString("lcm-address"), viper.GetInt("lcm-port")),
			"cluster_name=" + input.Name,
			"contract_id=" + input.ProjectId,
			"csp_id=" + cspIds[0],
			"creator=" + userId,
			"description=" + input.TemplateId + "-" + input.Description,
			"template_name=" + input.TemplateId,
			/*
				"machine_type=" + input.MachineType,
				"num_of_az=" + input.NumberOfAz,
				"machine_replicas=" + input.MachineReplicas,
			*/
		}

		workflowId, err := argowfClient.SumbitWorkflowFromWftpl(workflow, nameSpace, opts)
		if err != nil {
			log.Error(err)
			InternalServerError(w)
			return
		}
		log.Info("Submitted workflow: ", workflowId)

		// wait & get clusterId ( max 1min 	)
		cnt := 0
		for range time.Tick(2 * time.Second) {
			if cnt >= 60 { // max wait 60sec
				break
			}

			workflow, err := argowfClient.GetWorkflow("argo", workflowId)
			if err != nil {
				log.Error(err)
				InternalServerError(w)
				break
			}

			if workflow.Status.Phase != "Running" {
				log.Error(err)
				InternalServerError(w)
				break
			}

			if workflow.Status.Progress == "1/2" { // start creating cluster
				time.Sleep(time.Second * 5) // Buffer
				break
			}
			cnt += 1
		}
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
