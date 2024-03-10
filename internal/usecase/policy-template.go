package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
)

type IPolicyTemplateUsecase interface {
	Create(ctx context.Context, policyTemplate domain.PolicyTemplate) (policyTemplateId string, err error)
	Fetch(ctx context.Context, pg *pagination.Pagination) (policyTemplates []domain.PolicyTemplate, err error)
	Update(ctx context.Context, policyTemplateId uuid.UUID, update domain.UpdatePolicyTemplateRequest) (err error)
	Get(ctx context.Context, policyTemplateId uuid.UUID) (policyTemplates *domain.PolicyTemplate, err error)
	Delete(ctx context.Context, policyTemplateId uuid.UUID) (err error)
	IsPolicyTemplateNameExist(ctx context.Context, policyTemplateName string) (bool, error)
	IsPolicyTemplateKindExist(ctx context.Context, policyTemplateKind string) (bool, error)
	GetPolicyTemplateVersion(ctx context.Context, policyTemplateId uuid.UUID, version string) (policyTemplateVersionsReponse *domain.PolicyTemplate, err error)
	ListPolicyTemplateVersions(ctx context.Context, policyTemplateId uuid.UUID) (policyTemplateVersionsReponse *domain.ListPolicyTemplateVersionsResponse, err error)
	DeletePolicyTemplateVersion(ctx context.Context, policyTemplateId uuid.UUID, version string) (err error)
	CreatePolicyTemplateVersion(ctx context.Context, policyTemplateId uuid.UUID, newVersion string, schema []domain.ParameterDef, rego string, libs []string) (version string, err error)
}

type PolicyTemplateUsecase struct {
	organizationRepo repository.IOrganizationRepository
	clusterRepo      repository.IClusterRepository
	repo             repository.IPolicyTemplateRepository
}

func NewPolicyTemplateUsecase(r repository.Repository) IPolicyTemplateUsecase {
	return &PolicyTemplateUsecase{
		repo:             r.PolicyTemplate,
		organizationRepo: r.Organization,
		clusterRepo:      r.Cluster,
	}
}

func (u *PolicyTemplateUsecase) Create(ctx context.Context, dto domain.PolicyTemplate) (policyTemplateId string, err error) {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return "", httpErrors.NewUnauthorizedError(fmt.Errorf("Invalid token"), "P_INVALID_TOKEN", "")
	}

	exists, err := u.repo.ExistByName(dto.TemplateName)
	if err == nil && exists {
		return "", httpErrors.NewBadRequestError(httpErrors.DuplicateResource, "P_INVALID_POLICY_TEMPLATE_NAME", "policy template name already exists")
	}

	exists, err = u.repo.ExistByKind(dto.Kind)
	if err == nil && exists {
		return "", httpErrors.NewBadRequestError(httpErrors.DuplicateResource, "P_INVALID_POLICY_TEMPLATE_KIND", "policy template kind already exists")
	}

	for _, organizationId := range dto.PermittedOrganizationIds {
		_, err = u.organizationRepo.Get(organizationId)
		if err != nil {
			return "", httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId"), "", "")
		}
	}

	userId := user.GetUserId()
	dto.CreatorId = &userId
	id, err := u.repo.Create(dto)

	if err != nil {
		return "", err
	}

	return id.String(), nil
}

func (u *PolicyTemplateUsecase) Fetch(ctx context.Context, pg *pagination.Pagination) (policyTemplates []domain.PolicyTemplate, err error) {
	policyTemplates, err = u.repo.Fetch(pg)

	if err != nil {
		return nil, err
	}

	organizations, err := u.organizationRepo.Fetch(nil)
	if err == nil {
		for i, policyTemplate := range policyTemplates {
			permittedOrgIdSet := u.getPermittedOrganiationIdSet(&policyTemplate)

			u.updatePermittedOrganizations(organizations, permittedOrgIdSet, &policyTemplates[i])
		}
	}

	return policyTemplates, nil

}

func (u *PolicyTemplateUsecase) Get(ctx context.Context, policyTemplateID uuid.UUID) (policyTemplates *domain.PolicyTemplate, err error) {
	policyTemplate, err := u.repo.GetByID(policyTemplateID)

	if err != nil {
		return nil, err
	}

	permittedOrgIdSet := u.getPermittedOrganiationIdSet(policyTemplate)

	organizations, err := u.organizationRepo.Fetch(nil)
	if err == nil {
		u.updatePermittedOrganizations(organizations, permittedOrgIdSet, policyTemplate)
	}

	return policyTemplate, nil
}

func (u *PolicyTemplateUsecase) Update(ctx context.Context, policyTemplateId uuid.UUID, update domain.UpdatePolicyTemplateRequest) (err error) {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return httpErrors.NewBadRequestError(fmt.Errorf("Invalid token"), "", "")
	}

	_, err = u.repo.GetByID(policyTemplateId)
	if err != nil {
		return httpErrors.NewNotFoundError(err, "P_FAILED_FETCH_POLICY_TEMPLATE", "")
	}

	exists, err := u.repo.ExistByName(*update.TemplateName)
	if err == nil && exists {
		return httpErrors.NewBadRequestError(httpErrors.DuplicateResource, "P_INVALID_POLICY_TEMPLATE_NAME", "policy template name already exists")
	}

	if update.PermittedOrganizationIds != nil {
		for _, organizationId := range *update.PermittedOrganizationIds {
			_, err = u.organizationRepo.Get(organizationId)
			if err != nil {
				return httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId"), "", "")
			}
		}
	}

	updatorId := user.GetUserId()
	dto := domain.UpdatePolicyTemplateUpdate{
		ID:                       policyTemplateId,
		Type:                     "tks",
		UpdatorId:                updatorId,
		TemplateName:             update.TemplateName,
		Description:              update.Description,
		Severity:                 update.Severity,
		Deprecated:               update.Deprecated,
		PermittedOrganizationIds: update.PermittedOrganizationIds,
	}

	err = u.repo.Update(dto)
	if err != nil {
		return err
	}

	return nil
}

func (u *PolicyTemplateUsecase) Delete(ctx context.Context, policyTemplateId uuid.UUID) (err error) {
	return u.repo.Delete(policyTemplateId)
}

func (u *PolicyTemplateUsecase) IsPolicyTemplateNameExist(ctx context.Context, policyTemplateName string) (bool, error) {
	return u.repo.ExistByName(policyTemplateName)
}

func (u *PolicyTemplateUsecase) IsPolicyTemplateKindExist(ctx context.Context, policyTemplateKind string) (bool, error) {
	return u.repo.ExistByKind(policyTemplateKind)
}

func (u *PolicyTemplateUsecase) GetPolicyTemplateVersion(ctx context.Context, policyTemplateId uuid.UUID, version string) (policyTemplateVersionsReponse *domain.PolicyTemplate, err error) {
	policyTemplate, err := u.repo.GetPolicyTemplateVersion(policyTemplateId, version)

	if err != nil {
		return nil, err
	}

	permittedOrgIdSet := u.getPermittedOrganiationIdSet(policyTemplate)

	organizations, err := u.organizationRepo.Fetch(nil)
	if err == nil {
		u.updatePermittedOrganizations(organizations, permittedOrgIdSet, policyTemplate)
	}

	return policyTemplate, nil
}

func (*PolicyTemplateUsecase) updatePermittedOrganizations(organizations *[]domain.Organization, permittedOrgIdSet map[string]string, policyTemplate *domain.PolicyTemplate) {
	// 허용리스트가 비어있으면 모든 Org에 대해서 허용
	permitted := len(permittedOrgIdSet) == 0

	for _, organization := range *organizations {

		_, ok := permittedOrgIdSet[organization.ID]

		if !ok {
			policyTemplate.PermittedOrganizations = append(
				policyTemplate.PermittedOrganizations,
				domain.PermittedOrganization{
					OrganizationId:   organization.ID,
					OrganizationName: organization.Name,
					Permitted:        permitted,
				})
		}
	}
}

func (*PolicyTemplateUsecase) getPermittedOrganiationIdSet(policyTemplate *domain.PolicyTemplate) map[string]string {
	permittedOrgIdSet := make(map[string]string)

	for _, permittedOrg := range policyTemplate.PermittedOrganizations {
		// Set 처리를 위해서 키만 사용, 값은 아무거나
		permittedOrgIdSet[permittedOrg.OrganizationId] = "1"
	}
	return permittedOrgIdSet
}

func (u *PolicyTemplateUsecase) ListPolicyTemplateVersions(ctx context.Context, policyTemplateId uuid.UUID) (policyTemplateVersionsReponse *domain.ListPolicyTemplateVersionsResponse, err error) {
	return u.repo.ListPolicyTemplateVersions(policyTemplateId)
}

func (u *PolicyTemplateUsecase) DeletePolicyTemplateVersion(ctx context.Context, policyTemplateId uuid.UUID, version string) (err error) {
	return u.repo.DeletePolicyTemplateVersion(policyTemplateId, version)
}

func (u *PolicyTemplateUsecase) CreatePolicyTemplateVersion(ctx context.Context, policyTemplateId uuid.UUID, newVersion string, schema []domain.ParameterDef, rego string, libs []string) (version string, err error) {
	return u.repo.CreatePolicyTemplateVersion(policyTemplateId, newVersion, schema, rego, libs)
}
