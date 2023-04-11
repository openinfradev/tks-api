package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/openinfradev/tks-api/internal/auth/request"
	"github.com/openinfradev/tks-api/internal/repository"
	argowf "github.com/openinfradev/tks-api/pkg/argo-client"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type IStackUsecase interface {
	Get(stackId domain.StackId) (domain.Stack, error)
	Fetch(organizationId string) ([]domain.Stack, error)
	Create(ctx context.Context, dto domain.Stack) (err error)
	Delete(ctx context.Context, dto domain.Stack) error
}

type StackUsecase struct {
	clusterRepo       repository.IClusterRepository
	stackTemplateRepo repository.IStackTemplateRepository
	argo              argowf.ArgoClient
}

func NewStackUsecase(r repository.Repository, argoClient argowf.ArgoClient) IStackUsecase {
	return &StackUsecase{
		clusterRepo:       r.Cluster,
		stackTemplateRepo: r.StackTemplate,
		argo:              argoClient,
	}
}

func (u *StackUsecase) Create(ctx context.Context, dto domain.Stack) (err error) {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return httpErrors.NewBadRequestError(fmt.Errorf("Invalid token"))
	}

	stackTemplate, err := u.stackTemplateRepo.Get(dto.StackTemplateId)
	if err != nil {
		return httpErrors.NewInternalServerError(errors.Wrap(err, "Invalid stackTemplateId"))
	}

	workflow := ""
	if stackTemplate.Template == "aws-reference" || stackTemplate.Template == "eks-reference" {
		workflow = "tks-stack-create-aws"
	} else if stackTemplate.Template == "aws-msa-reference" || stackTemplate.Template == "eks-msa-reference" {
		workflow = "tks-stack-create-aws-msa"
	} else {
		log.Error("Invalid template  : ", stackTemplate.Template)
		return httpErrors.NewInternalServerError(fmt.Errorf("Invalid stackTemplate. %s", stackTemplate.Template))
	}

	opts := argowf.SubmitOptions{}
	opts.Parameters = []string{
		fmt.Sprintf("tks_api_url=%s", viper.GetString("external-address")),
		"cluster_name=" + dto.Name,
		"description=" + dto.Description,
		"organization_id=" + dto.OrganizationId,
		"cloud_account_id=" + dto.CloudSettingId.String(),
		"stack_template_name=" + stackTemplate.Template,
		"creator=" + user.GetUserId().String(),
		/*
			"machine_type=" + input.MachineType,
			"num_of_az=" + input.NumberOfAz,
			"machine_replicas=" + input.MachineReplicas,
		*/
	}

	workflowId, err := u.argo.SumbitWorkflowFromWftpl(workflow, opts)
	if err != nil {
		log.Error(err)
		return errors.Wrap(err, "Failed to call workflow. workflow")
	}
	log.Info("Submitted workflow: ", workflowId)

	// wait & get clusterId ( max 1min 	)
	cnt := 0

	ticker := time.NewTicker(time.Second * 2)
	for range ticker.C {
		if cnt >= 60 { // max wait 60sec
			break
		}

		workflow, err := u.argo.GetWorkflow("argo", workflowId)
		if err != nil {
			return err
		}

		if workflow.Status.Phase != "Running" {
			return err
		}

		if workflow.Status.Progress == "1/2" { // start creating cluster
			time.Sleep(time.Second * 5) // Buffer
			break
		}
		cnt += 1
	}

	// [TODO] need clusterId?
	return nil
}

func (u *StackUsecase) Get(stackId domain.StackId) (out domain.Stack, err error) {
	res, err := u.clusterRepo.Get(domain.ClusterId(stackId))
	if err != nil {
		return domain.Stack{}, err
	}
	out = reflectClusterToStack(res)
	return
}

func (u *StackUsecase) Fetch(organizationId string) (out []domain.Stack, err error) {
	res, err := u.clusterRepo.FetchByOrganizationId(organizationId)
	if err != nil {
		return nil, err
	}

	for _, cluster := range res {
		out = append(out, reflectClusterToStack(cluster))
	}

	return
}

func (u *StackUsecase) Delete(ctx context.Context, dto domain.Stack) (err error) {
	return fmt.Errorf("Need implementaion")
}

func reflectClusterToStack(cluster domain.Cluster) domain.Stack {
	status := domain.StackStatus_PENDING
	statusDesc := ""
	return domain.Stack{
		ID:              domain.StackId(cluster.ID),
		OrganizationId:  cluster.OrganizationId,
		Name:            cluster.Name,
		Description:     cluster.Description,
		Status:          status,
		StatusDesc:      statusDesc,
		CloudSettingId:  cluster.CloudSettingId,
		CloudSetting:    cluster.CloudSetting,
		StackTemplateId: cluster.StackTemplateId,
		StackTemplate:   cluster.StackTemplate,
		CreatorId:       cluster.CreatorId,
		Creator:         cluster.Creator,
		UpdatorId:       cluster.UpdatorId,
		Updator:         cluster.Updator,
		CreatedAt:       cluster.CreatedAt,
		UpdatedAt:       cluster.UpdatedAt,
	}
}
