package usecase

import (
	"context"
	"fmt"

	"github.com/openinfradev/tks-api/internal/auth/request"
	"github.com/openinfradev/tks-api/internal/helper"
	"github.com/openinfradev/tks-api/internal/keycloak"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/pkg/errors"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/repository"
	argowf "github.com/openinfradev/tks-api/pkg/argo-client"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
)

type IOrganizationUsecase interface {
	Create(context.Context, *domain.Organization) (organizationId string, err error)
	Fetch() (*[]domain.Organization, error)
	Get(organizationId string) (domain.Organization, error)
	Update(organizationId string, in domain.UpdateOrganizationRequest) (domain.Organization, error)
	UpdatePrimaryClusterId(organizationId string, clusterId string) (err error)
	Delete(organizationId string, accessToken string) error
}

type OrganizationUsecase struct {
	repo repository.IOrganizationRepository
	argo argowf.ArgoClient
	kc   keycloak.IKeycloak
}

func NewOrganizationUsecase(r repository.IOrganizationRepository, argoClient argowf.ArgoClient, kc keycloak.IKeycloak) IOrganizationUsecase {
	return &OrganizationUsecase{
		repo: r,
		argo: argoClient,
		kc:   kc,
	}
}

func (u *OrganizationUsecase) Create(ctx context.Context, in *domain.Organization) (organizationId string, err error) {
	creator := uuid.Nil
	if in.Creator != "" {
		creator, err = uuid.Parse(in.Creator)
		if err != nil {
			return "", err
		}
	}
	token, ok := request.TokenFrom(ctx)
	if !ok {
		return "", fmt.Errorf("token in the context is empty")
	}

	// Create realm in keycloak
	if organizationId, err = u.kc.CreateRealm(helper.GenerateOrganizationId(), domain.Organization{}, token); err != nil {
		return "", err
	}

	_, err = u.repo.Create(organizationId, in.Name, creator, in.Phone, in.Description)
	if err != nil {
		return "", err
	}
	workflowId, err := u.argo.SumbitWorkflowFromWftpl(
		"tks-create-contract-repo",
		argowf.SubmitOptions{
			Parameters: []string{
				"contract_id=" + organizationId,
			},
		})
	if err != nil {
		log.Error("failed to submit argo workflow template. err : ", err)
		return "", errors.Wrap(err, "Failed to call argo workflow")
	}
	log.Info("submited workflow :", workflowId)

	if err := u.repo.InitWorkflow(organizationId, workflowId, domain.OrganizationStatus_CREATING); err != nil {
		return "", errors.Wrap(err, "Failed to init workflow")
	}

	return organizationId, nil
}
func (u *OrganizationUsecase) Fetch() (out *[]domain.Organization, err error) {
	organizations, err := u.repo.Fetch()
	if err != nil {
		return nil, err
	}
	return organizations, nil
}
func (u *OrganizationUsecase) Get(organizationId string) (res domain.Organization, err error) {
	res, err = u.repo.Get(organizationId)
	if err != nil {
		return domain.Organization{}, httpErrors.NewNotFoundError(err)
	}
	return res, nil
}

func (u *OrganizationUsecase) Delete(organizationId string, accessToken string) (err error) {
	_, err = u.Get(organizationId)
	if err != nil {
		return err
	}

	// Delete realm in keycloak
	if err := u.kc.DeleteRealm(organizationId, accessToken); err != nil {
		return err
	}

	// [TODO] validation
	// cluster 나 appgroup 등이 삭제 되었는지 확인
	err = u.repo.Delete(organizationId)
	if err != nil {
		return err
	}

	return nil
}

func (u *OrganizationUsecase) Update(organizationId string, in domain.UpdateOrganizationRequest) (domain.Organization, error) {
	_, err := u.Get(organizationId)
	if err != nil {
		return domain.Organization{}, httpErrors.NewNotFoundError(err)
	}

	res, err := u.repo.Update(organizationId, in)
	if err != nil {
		return domain.Organization{}, err
	}

	return res, nil
}

func (u *OrganizationUsecase) UpdatePrimaryClusterId(organizationId string, clusterId string) (err error) {
	if !helper.ValidateClusterId(clusterId) {
		return httpErrors.NewBadRequestError(fmt.Errorf("Invalid clusterId"))
	}

	organization, err := u.Get(organizationId)
	if err != nil {
		return httpErrors.NewNotFoundError(err)
	}

	// [TODO] need refactoring about reflect
	in := domain.UpdateOrganizationRequest{
		Description:      organization.Description,
		Phone:            organization.Phone,
		PrimaryClusterId: clusterId,
	}

	_, err = u.repo.Update(organizationId, in)
	if err != nil {
		return err
	}
	return nil
}
