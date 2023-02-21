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

type IOrganizationUsecase interface {
	Fetch() ([]domain.Organization, error)
	Get(organizationId string) (domain.Organization, error)
	Create(domain.Organization) (organizationId string, err error)
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
	log.Info("newly created Organization Id:", organizationId)

	workflowId, err := u.argo.SumbitWorkflowFromWftpl(
		"tks-create-organization-repo",
		argowf.SubmitOptions{
			Parameters: []string{
				"organization_id=" + organizationId,
				"revision=" + viper.GetString("revision"),
			},
		})
	if err != nil {
		log.Error("failed to submit argo workflow template. err : ", err)
		return "", fmt.Errorf("Failed to call argo workflow : %s", err)
	}
	log.Info("submited workflow :", workflowId)

	return organizationId, nil
}

func (u *OrganizationUsecase) Get(organizationId string) (res domain.Organization, err error) {
	res, err = u.repo.Get(organizationId)
	if err != nil {
		return domain.Organization{}, err
	}
	return res, nil
}
