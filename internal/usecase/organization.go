package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/helper"
	"github.com/openinfradev/tks-api/internal/keycloak"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/repository"
	argowf "github.com/openinfradev/tks-api/pkg/argo-client"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type IOrganizationUsecase interface {
	Create(context.Context, *model.Organization) (organizationId string, err error)
	Fetch(pg *pagination.Pagination) (*[]model.Organization, error)
	Get(organizationId string) (model.Organization, error)
	Update(organizationId string, in domain.UpdateOrganizationRequest) (model.Organization, error)
	UpdatePrimaryClusterId(organizationId string, clusterId string) (err error)
	Delete(organizationId string, accessToken string) error
}

type OrganizationUsecase struct {
	repo     repository.IOrganizationRepository
	roleRepo repository.IRoleRepository
	argo     argowf.ArgoClient
	kc       keycloak.IKeycloak
}

func NewOrganizationUsecase(r repository.Repository, argoClient argowf.ArgoClient, kc keycloak.IKeycloak) IOrganizationUsecase {
	return &OrganizationUsecase{
		repo:     r.Organization,
		roleRepo: r.Role,
		argo:     argoClient,
		kc:       kc,
	}
}

func (u *OrganizationUsecase) Create(ctx context.Context, in *model.Organization) (organizationId string, err error) {
	creator := uuid.Nil
	if in.Creator != "" {
		creator, err = uuid.Parse(in.Creator)
		if err != nil {
			return "", err
		}
	}

	// Create realm in keycloak
	if organizationId, err = u.kc.CreateRealm(helper.GenerateOrganizationId()); err != nil {
		return "", err
	}

	// Create organization in DB
	_, err = u.repo.Create(organizationId, in.Name, creator, in.Phone, in.Description)
	if err != nil {
		return "", err
	}

	workflowId, err := u.argo.SumbitWorkflowFromWftpl(
		"tks-create-contract-repo",
		argowf.SubmitOptions{
			Parameters: []string{
				"contract_id=" + organizationId,
				"base_repo_branch=" + viper.GetString("revision"),
				"keycloak_url=" + strings.TrimSuffix(viper.GetString("keycloak-address"), "/auth"),
			},
		})
	if err != nil {
		log.ErrorWithContext(ctx, "failed to submit argo workflow template. err : ", err)
		return "", errors.Wrap(err, "Failed to call argo workflow")
	}
	log.InfoWithContext(ctx, "submited workflow :", workflowId)

	if err := u.repo.InitWorkflow(organizationId, workflowId, domain.OrganizationStatus_CREATING); err != nil {
		return "", errors.Wrap(err, "Failed to init workflow")
	}

	return organizationId, nil
}
func (u *OrganizationUsecase) Fetch(pg *pagination.Pagination) (out *[]model.Organization, err error) {
	organizations, err := u.repo.Fetch(pg)
	if err != nil {
		return nil, err
	}
	return organizations, nil
}
func (u *OrganizationUsecase) Get(organizationId string) (res model.Organization, err error) {
	res, err = u.repo.Get(organizationId)
	if err != nil {
		return model.Organization{}, httpErrors.NewNotFoundError(err, "", "")
	}
	return res, nil
}

func (u *OrganizationUsecase) Delete(organizationId string, accessToken string) (err error) {
	_, err = u.Get(organizationId)
	if err != nil {
		return err
	}

	// Delete realm in keycloak
	if err := u.kc.DeleteRealm(organizationId); err != nil {
		return err
	}

	// delete roles in DB
	roles, err := u.roleRepo.ListTksRoles(organizationId, nil)
	if err != nil {
		return err
	}
	for _, role := range roles {
		if err := u.roleRepo.Delete(role.ID); err != nil {
			return err
		}
	}

	// delete organization in DB
	err = u.repo.Delete(organizationId)
	if err != nil {
		return err
	}

	return nil
}

func (u *OrganizationUsecase) Update(organizationId string, in domain.UpdateOrganizationRequest) (model.Organization, error) {
	_, err := u.Get(organizationId)
	if err != nil {
		return model.Organization{}, httpErrors.NewNotFoundError(err, "", "")
	}

	res, err := u.repo.Update(organizationId, in)
	if err != nil {
		return model.Organization{}, err
	}

	return res, nil
}

func (u *OrganizationUsecase) UpdatePrimaryClusterId(organizationId string, clusterId string) (err error) {
	if clusterId != "" && !helper.ValidateClusterId(clusterId) {
		return httpErrors.NewBadRequestError(fmt.Errorf("Invalid clusterId"), "", "")
	}

	_, err = u.Get(organizationId)
	if err != nil {
		return httpErrors.NewNotFoundError(err, "", "")
	}

	err = u.repo.UpdatePrimaryClusterId(organizationId, clusterId)
	if err != nil {
		return err
	}
	return nil
}
