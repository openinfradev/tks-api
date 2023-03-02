package usecase

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/domain"
	"github.com/openinfradev/tks-api/internal/repository"
	argowf "github.com/openinfradev/tks-api/pkg/argo-client"
	"github.com/openinfradev/tks-api/pkg/log"
)

type IOrganizationUsecase interface {
	Fetch() ([]domain.Organization, error)
	Get(organizationId string) (domain.Organization, error)
	Create(domain.Organization) (organizationId string, err error)
	Delete(organizationId string) (string, error)
}

type OrganizationUsecase struct {
	repo repository.IOrganizationRepository
	argo argowf.ArgoClient
}

func NewOrganizationUsecase(r repository.IOrganizationRepository, argoClient argowf.ArgoClient) IOrganizationUsecase {
	return &OrganizationUsecase{
		repo: r,
		argo: argoClient,
	}
}

func (u *OrganizationUsecase) Fetch() (out []domain.Organization, err error) {
	organizations, err := u.repo.Fetch()
	if err != nil {
		return nil, err
	}
	return organizations, nil
}

func (u *OrganizationUsecase) Create(in domain.Organization) (organizationId string, err error) {
	creator := uuid.Nil
	if in.Creator != "" {
		creator, err = uuid.Parse(in.Creator)
		if err != nil {
			return "", err
		}
	}
	organizationId, err = u.repo.Create(in.Name, creator, in.Description)
	if err != nil {
		return "", err
	}
	log.Info("newly created Organization ID:", organizationId)

	workflowId, err := u.argo.SumbitWorkflowFromWftpl(
		"tks-create-contract-repo",
		argowf.SubmitOptions{
			Parameters: []string{
				"contract_id=" + organizationId,
			},
		})
	if err != nil {
		log.Error("failed to submit argo workflow template. err : ", err)
		return "", fmt.Errorf("Failed to call argo workflow : %s", err)
	}
	log.Info("submited workflow :", workflowId)

	if err := u.repo.InitWorkflow(organizationId, workflowId); err != nil {
		return "", fmt.Errorf("Failed to initialize organization status to 'CREATING'. err : %s", err)
	}

	return organizationId, nil
}

func (u *OrganizationUsecase) Get(organizationId string) (res domain.Organization, err error) {
	res, err = u.repo.Get(organizationId)
	if err != nil {
		return domain.Organization{}, err
	}
	return res, nil
}

func (u *OrganizationUsecase) Delete(organizationId string) (res string, err error) {
	_, err = u.Get(organizationId)
	if err != nil {
		return "", err
	}

	// [TODO] validation
	// cluster 나 appgroup 등이 삭제 되었는지 확인

	res, err = u.repo.Delete(organizationId)
	if err != nil {
		return "", err
	}

	return res, nil
}
