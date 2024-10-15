package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/helper"
	"github.com/openinfradev/tks-api/internal/keycloak"
	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
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
	Fetch(ctx context.Context, pg *pagination.Pagination) (*[]model.Organization, error)
	Get(ctx context.Context, organizationId string) (model.Organization, error)
	Update(ctx context.Context, organizationId string, dto model.Organization) (model.Organization, error)
	UpdatePrimaryClusterId(ctx context.Context, organizationId string, clusterId string) (err error)
	ChangeAdminId(ctx context.Context, organizationId string, adminId uuid.UUID) error
	Delete(ctx context.Context, organizationId string, accessToken string) error
}

type OrganizationUsecase struct {
	repo                           repository.IOrganizationRepository
	userRepo                       repository.IUserRepository
	roleRepo                       repository.IRoleRepository
	clusterRepo                    repository.IClusterRepository
	stackTemplateRepo              repository.IStackTemplateRepository
	systemNotificationRuleRepo     repository.ISystemNotificationRuleRepository
	systemNotificationTemplateRepo repository.ISystemNotificationTemplateRepository
	argo                           argowf.ArgoClient
	kc                             keycloak.IKeycloak
}

func NewOrganizationUsecase(r repository.Repository, argoClient argowf.ArgoClient, kc keycloak.IKeycloak) IOrganizationUsecase {
	return &OrganizationUsecase{
		repo:                           r.Organization,
		userRepo:                       r.User,
		roleRepo:                       r.Role,
		clusterRepo:                    r.Cluster,
		stackTemplateRepo:              r.StackTemplate,
		systemNotificationRuleRepo:     r.SystemNotificationRule,
		systemNotificationTemplateRepo: r.SystemNotificationTemplate,
		argo:                           argoClient,
		kc:                             kc,
	}
}

func (u *OrganizationUsecase) Create(ctx context.Context, in *model.Organization) (organizationId string, err error) {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return "", httpErrors.NewBadRequestError(fmt.Errorf("Invalid token"), "", "")
	}
	userId := user.GetUserId()
	in.CreatorId = &userId

	pg := pagination.NewPaginationWithFilter("name", "", "$eq", []string{in.Name})
	organizations, err := u.repo.Fetch(ctx, pg)
	if err != nil {
		return "", err
	}
	if organizations != nil && len(*organizations) > 0 {
		return "", httpErrors.NewBadRequestError(fmt.Errorf("duplicate organization name"), "O_CREATE_ALREADY_EXISTED_NAME", "")
	}

	// Create realm in keycloak
	if organizationId, err = u.kc.CreateRealm(ctx, helper.GenerateOrganizationId()); err != nil {
		return "", err
	}
	in.ID = organizationId

	// Create organization in DB
	_, err = u.repo.Create(ctx, in)
	if err != nil {
		return "", err
	}

	workflowId, err := u.argo.SumbitWorkflowFromWftpl(
		ctx,
		"tks-create-contract-repo",
		argowf.SubmitOptions{
			Parameters: []string{
				"contract_id=" + organizationId,
				"base_repo_branch=" + viper.GetString("revision"),
				"keycloak_url=" + strings.TrimSuffix(viper.GetString("keycloak-address"), "/auth"),
			},
		})
	if err != nil {
		log.Error(ctx, "failed to submit argo workflow template. err : ", err)
		return "", errors.Wrap(err, "Failed to call argo workflow")
	}
	log.Info(ctx, "submited workflow :", workflowId)

	if err := u.repo.InitWorkflow(ctx, organizationId, workflowId, domain.OrganizationStatus_CREATING); err != nil {
		return "", errors.Wrap(err, "Failed to init workflow")
	}

	return organizationId, nil
}
func (u *OrganizationUsecase) Fetch(ctx context.Context, pg *pagination.Pagination) (out *[]model.Organization, err error) {
	organizations, err := u.repo.Fetch(ctx, pg)
	if err != nil {
		return nil, err
	}

	// Make Admin object
	users, err := u.userRepo.List(ctx)
	if err == nil {
		for i, organization := range *organizations {
			if organization.AdminId == nil {
				continue
			}
			for _, user := range *users {
				user := user
				if user.ID == *organization.AdminId {
					(*organizations)[i].Admin = &user
				}
			}
		}
	}

	return organizations, nil
}
func (u *OrganizationUsecase) Get(ctx context.Context, organizationId string) (out model.Organization, err error) {
	out, err = u.repo.Get(ctx, organizationId)
	if err != nil {
		return model.Organization{}, httpErrors.NewNotFoundError(err, "", "")
	}

	// Make Admin object
	if out.AdminId != nil {
		admin, err := u.userRepo.GetByUuid(ctx, *out.AdminId)
		if err == nil {
			out.Admin = &admin
		}
	}

	clusters, err := u.clusterRepo.FetchByOrganizationId(ctx, organizationId, uuid.Nil, nil)
	if err != nil {
		log.Info(ctx, err)
		out.ClusterCount = 0
	}
	out.ClusterCount = len(clusters)
	return out, nil

}

func (u *OrganizationUsecase) Delete(ctx context.Context, organizationId string, accessToken string) (err error) {
	_, err = u.Get(ctx, organizationId)
	if err != nil {
		return err
	}

	// Delete realm in keycloak
	if err := u.kc.DeleteRealm(ctx, organizationId); err != nil {
		return err
	}

	// delete roles in DB
	roles, err := u.roleRepo.ListTksRoles(ctx, organizationId, nil)
	if err != nil {
		return err
	}
	for _, role := range roles {
		if err := u.roleRepo.Delete(ctx, role.ID); err != nil {
			return err
		}
	}

	// delete organization in DB
	err = u.repo.Delete(ctx, organizationId)
	if err != nil {
		return err
	}

	return nil
}

func (u *OrganizationUsecase) Update(ctx context.Context, organizationId string, in model.Organization) (model.Organization, error) {
	_, err := u.Get(ctx, organizationId)
	if err != nil {
		return model.Organization{}, httpErrors.NewNotFoundError(err, "", "")
	}

	res, err := u.repo.Update(ctx, organizationId, in)
	if err != nil {
		return model.Organization{}, err
	}

	return res, nil
}

func (u *OrganizationUsecase) UpdatePrimaryClusterId(ctx context.Context, organizationId string, clusterId string) (err error) {
	if clusterId != "" && !helper.ValidateClusterId(clusterId) {
		return httpErrors.NewBadRequestError(fmt.Errorf("Invalid clusterId"), "", "")
	}

	_, err = u.Get(ctx, organizationId)
	if err != nil {
		return httpErrors.NewNotFoundError(err, "", "")
	}

	err = u.repo.UpdatePrimaryClusterId(ctx, organizationId, clusterId)
	if err != nil {
		return err
	}
	return nil
}

func (u *OrganizationUsecase) ChangeAdminId(ctx context.Context, organizationId string, adminId uuid.UUID) error {
	_, err := u.Get(ctx, organizationId)
	if err != nil {
		return httpErrors.NewNotFoundError(err, "", "")
	}

	err = u.repo.UpdateAdminId(ctx, organizationId, adminId)
	if err != nil {
		return err
	}

	return nil
}
