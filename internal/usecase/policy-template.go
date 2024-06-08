package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/openinfradev/tks-api/pkg/log"
	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	policytemplate "github.com/openinfradev/tks-api/internal/policy-template"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
)

type IPolicyTemplateUsecase interface {
	Create(ctx context.Context, policyTemplate model.PolicyTemplate) (policyTemplateId uuid.UUID, err error)
	Fetch(ctx context.Context, organizationId *string, pg *pagination.Pagination) (policyTemplates []model.PolicyTemplate, err error)
	Update(ctx context.Context, organizationId *string, policyTemplateId uuid.UUID, templateName *string, description *string,
		severity *string, deorecated *bool, permittedOrganizationIds *[]string) (err error)
	Get(ctx context.Context, organizationId *string, policyTemplateId uuid.UUID) (policyTemplates *model.PolicyTemplate, err error)
	Delete(ctx context.Context, organizationId *string, policyTemplateId uuid.UUID) (err error)
	IsPolicyTemplateNameExist(ctx context.Context, organizationId *string, policyTemplateName string) (bool, error)
	IsPolicyTemplateKindExist(ctx context.Context, organizationId *string, policyTemplateKind string) (bool, error)
	GetPolicyTemplateVersion(ctx context.Context, organizationId *string, policyTemplateId uuid.UUID, version string) (policyTemplateVersionsReponse *model.PolicyTemplate, err error)
	ListPolicyTemplateVersions(ctx context.Context, organizationId *string, policyTemplateId uuid.UUID) (policyTemplateVersionsReponse *domain.ListPolicyTemplateVersionsResponse, err error)
	DeletePolicyTemplateVersion(ctx context.Context, organizationId *string, policyTemplateId uuid.UUID, version string) (err error)
	CreatePolicyTemplateVersion(ctx context.Context, organizationId *string, policyTemplateId uuid.UUID, newVersion string, schema []*domain.ParameterDef, rego string, libs []string,
		syncKinds *[]string, syncJson *string) (version string, err error)

	RegoCompile(request *domain.RegoCompileRequest, parseParameter bool) (response *domain.RegoCompileResponse, err error)

	// FillPermittedOrganizations(ctx context.Context,
	// 	policyTemplate *model.PolicyTemplate, out *admin_domain.PolicyTemplateResponse) error
	// FillPermittedOrganizationsForList(ctx context.Context,
	// 	policyTemplates *[]model.PolicyTemplate, outs *[]admin_domain.PolicyTemplateResponse) error

	ListPolicyTemplateStatistics(ctx context.Context, organizationId *string, policyTemplateId uuid.UUID) (statistics []model.UsageCount, err error)
	GetPolicyTemplateDeploy(ctx context.Context, organizationId *string, policyTemplateId uuid.UUID) (deployInfo domain.GetPolicyTemplateDeployResponse, err error)

	ExtractPolicyParameters(ctx context.Context, organizationId *string, policyTemplateId uuid.UUID, version string, rego string, libs []string) (response *domain.RegoCompileResponse, err error)

	AddPermittedPolicyTemplatesForOrganization(ctx context.Context, organizationId string, policyTemplateIds []uuid.UUID) (err error)
	UpdatePermittedPolicyTemplatesForOrganization(ctx context.Context, organizationId string, policyTemplateIds []uuid.UUID) (err error)
	DeletePermittedPolicyTemplatesForOrganization(ctx context.Context, organizationId string, policyTemplateIds []uuid.UUID) (err error)
}

type PolicyTemplateUsecase struct {
	organizationRepo repository.IOrganizationRepository
	clusterRepo      repository.IClusterRepository
	policyRepo       repository.IPolicyRepository
	repo             repository.IPolicyTemplateRepository
}

func NewPolicyTemplateUsecase(r repository.Repository) IPolicyTemplateUsecase {
	return &PolicyTemplateUsecase{
		repo:             r.PolicyTemplate,
		policyRepo:       r.Policy,
		organizationRepo: r.Organization,
		clusterRepo:      r.Cluster,
	}
}

func (u *PolicyTemplateUsecase) Create(ctx context.Context, dto model.PolicyTemplate) (policyTemplateId uuid.UUID, err error) {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return uuid.Nil, httpErrors.NewUnauthorizedError(fmt.Errorf("invalid token"), "A_INVALID_TOKEN", "")
	}

	if dto.SyncKinds != nil && dto.SyncJson != nil {
		return uuid.Nil, httpErrors.NewBadRequestError(httpErrors.DuplicateResource, "PT_INVALID_SYNC", "both sync kinds and json specified")
	}

	if dto.SyncKinds != nil {
		if _, err := policytemplate.CheckAndConvertToSyncData(*dto.SyncKinds); err != nil {
			return uuid.Nil, httpErrors.NewBadRequestError(httpErrors.DuplicateResource, "PT_INVALID_SYNC", "invalid sync kind")
		}
	}

	if dto.SyncJson != nil {
		result, err := policytemplate.ParseAndCheckSyncData(*dto.SyncJson)

		if err != nil {
			return uuid.Nil, httpErrors.NewBadRequestError(httpErrors.DuplicateResource, "PT_INVALID_SYNC", "invalid sync json")
		}

		formatted, err := policytemplate.MarshalSyncData(result)

		if err == nil {
			dto.SyncJson = &formatted
		}
	}

	compileResult, err := u.RegoCompile(&domain.RegoCompileRequest{Rego: dto.Rego, Libs: dto.Libs}, false)

	if err != nil {
		return uuid.Nil, httpErrors.NewBadRequestError(fmt.Errorf("rego compile error"), "PT_INVALID_REGO_SYNTAX", "")
	}

	if len(compileResult.Errors) > 0 {
		compileErrors := []string{}

		for _, errorResult := range compileResult.Errors {
			compileErrors = append(compileErrors, errorResult.Text)
		}

		detail := strings.Join(compileErrors, "\n")

		return uuid.Nil, httpErrors.NewBadRequestError(fmt.Errorf("rego compile error"), "PT_INVALID_REGO_SYNTAX", detail)
	}

	if dto.IsTksTemplate() {
		exists, err := u.repo.ExistByName(ctx, dto.TemplateName)
		if err == nil && exists {
			return uuid.Nil, httpErrors.NewBadRequestError(httpErrors.DuplicateResource, "PT_CREATE_ALREADY_EXISTED_NAME", "policy template name already exists")
		}

		exists, err = u.repo.ExistByKind(ctx, dto.Kind)
		if err == nil && exists {
			return uuid.Nil, httpErrors.NewBadRequestError(httpErrors.DuplicateResource, "PT_CREATE_ALREADY_EXISTED_KIND", "policy template kind already exists")
		}
	} else {
		exists, err := u.repo.ExistByNameInOrganization(ctx, *dto.OrganizationId, dto.TemplateName)
		if err == nil && exists {
			return uuid.Nil, httpErrors.NewBadRequestError(httpErrors.DuplicateResource, "PT_CREATE_ALREADY_EXISTED_NAME", "policy template name already exists")
		}

		exists, err = u.repo.ExistByKindInOrganization(ctx, *dto.OrganizationId, dto.Kind)
		if err == nil && exists {
			return uuid.Nil, httpErrors.NewBadRequestError(httpErrors.DuplicateResource, "PT_CREATE_ALREADY_EXISTED_KIND", "policy template kind already exists")
		}
	}

	if err := policytemplate.ValidateParamDefs(dto.ParametersSchema); err != nil {
		return uuid.Nil, httpErrors.NewBadRequestError(err, "PT_INVALID_PARAMETER_SCHEMA", "")
	}

	if dto.IsTksTemplate() {
		// TKS 템블릿이면
		dto.Mandatory = false
		dto.OrganizationId = nil
		dto.PermittedOrganizations = make([]model.Organization, len(dto.PermittedOrganizationIds))

		for i, organizationId := range dto.PermittedOrganizationIds {
			organization, err := u.organizationRepo.Get(ctx, organizationId)
			if err != nil {
				return uuid.Nil, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "C_INVALID_ORGANIZATION_ID", "")
			}
			dto.PermittedOrganizations[i] = organization
		}
	} else {
		dto.PermittedOrganizationIds = make([]string, 0)
		dto.PermittedOrganizations = make([]model.Organization, 0)
	}

	userId := user.GetUserId()
	dto.CreatorId = &userId

	dto.Rego = policytemplate.FormatRegoCode(dto.Rego)
	dto.Libs = policytemplate.FormatLibCode(dto.Libs)

	id, err := u.repo.Create(ctx, dto)

	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func (u *PolicyTemplateUsecase) Fetch(ctx context.Context, organizationId *string, pg *pagination.Pagination) (policyTemplates []model.PolicyTemplate, err error) {
	if organizationId == nil {
		policyTemplates, err = u.repo.Fetch(ctx, pg)
	} else {
		policyTemplates, err = u.repo.FetchForOrganization(ctx, *organizationId, pg)
	}

	if err != nil {
		log.Errorf(ctx, "error is :%s(%T)", err.Error(), err)
	}

	organizations, err := u.organizationRepo.Fetch(ctx, nil)

	if err != nil {
		log.Errorf(ctx, "error is :%s(%T)", err.Error(), err)
	}

	var primaryClusterId string

	if organizationId != nil {
		org, err := u.organizationRepo.Get(ctx, *organizationId)

		if err != nil {
			log.Errorf(ctx, "error is :%s(%T)", err.Error(), err)
		} else {
			primaryClusterId = org.PrimaryClusterId
		}
	}

	for i := range policyTemplates {
		// 단순히 참조하면 업데이트가 안되므로 pointer derefrencing
		policyTemplate := &policyTemplates[i]
		if policyTemplate.IsTksTemplate() && len(policyTemplate.PermittedOrganizations) == 0 {
			if organizations != nil {
				(*policyTemplate).PermittedOrganizations = *organizations
			}
		}

		if organizationId != nil {
			(*policyTemplate).LatestVersion = policyTemplate.Version

			// 에러 없이 primaryClusterId를 가져왔을 때만 처리
			if len(primaryClusterId) > 0 {

				templateCR, err := policytemplate.GetTksPolicyTemplateCR(ctx, primaryClusterId, policyTemplate.ResoureName())

				if err == nil && templateCR != nil {
					(*policyTemplate).CurrentVersion = templateCR.Spec.Version
				} else if errors.IsNotFound(err) || templateCR == nil {
					// 템플릿이 존재하지 않으면 최신 버전으로 배포되므로
					(*policyTemplate).CurrentVersion = policyTemplate.LatestVersion
				} else {
					// 통신 실패 등 기타 에러, 버전을 세팅하지 않아 에러임을 알수 있도록 로그를 남김
					log.Errorf(ctx, "error is :%s(%T)", err.Error(), err)
				}
			}
		}
	}

	return policyTemplates, err
}

// func (u *PolicyTemplateUsecase) FillPermittedOrganizations(ctx context.Context,
// 	policyTemplate *model.PolicyTemplate, out *admin_domain.PolicyTemplateResponse) error {
// 	organizations, err := u.organizationRepo.Fetch(ctx, nil)

// 	if err != nil {
// 		return err
// 	}

// 	u.fillPermittedOrganizations(ctx, organizations, policyTemplate, out)

// 	return nil
// }

// func (u *PolicyTemplateUsecase) FillPermittedOrganizationsForList(ctx context.Context,
// 	policyTemplates *[]model.PolicyTemplate, outs *[]admin_domain.PolicyTemplateResponse) error {

// 	organizations, err := u.organizationRepo.Fetch(ctx, nil)

// 	if err != nil {
// 		return err
// 	}

// 	results := *outs

// 	for i, policyTemplate := range *policyTemplates {
// 		u.fillPermittedOrganizations(ctx, organizations, &policyTemplate, &results[i])
// 	}

// 	return nil
// }

// // 모든 조직 목록에 대해 허용 여부 업데이트
// func (u *PolicyTemplateUsecase) fillPermittedOrganizations(_ context.Context, organizations *[]model.Organization, policyTemplate *model.PolicyTemplate, out *admin_domain.PolicyTemplateResponse) {
// 	if policyTemplate == nil || organizations == nil || out == nil {
// 		return
// 	}

// 	if policyTemplate.IsOrganizationTemplate() {
// 		return
// 	}

// 	// 정책 템플릿에서 허용된 조직 목록이 없다는 것은 모든 조직이 사용할 수 있음을 의미함
// 	allPermitted := len(policyTemplate.PermittedOrganizationIds) == 0

// 	// 허용된 조직 포함 여부를 효율적으로 처리하기 위해 ID 리스트를 셋으로 변환
// 	permittedOrganizationIdSet := mapset.NewSet(policyTemplate.PermittedOrganizationIds...)

// 	out.PermittedOrganizations = make([]admin_domain.PermittedOrganization, len(*organizations))

// 	for i, organization := range *organizations {
// 		permitted := allPermitted || permittedOrganizationIdSet.ContainsOne(organization.ID)

// 		out.PermittedOrganizations[i] = admin_domain.PermittedOrganization{
// 			OrganizationId:   organization.ID,
// 			OrganizationName: organization.Name,
// 			Permitted:        permitted,
// 		}
// 	}
// }

func (u *PolicyTemplateUsecase) Get(ctx context.Context, organizationId *string, policyTemplateID uuid.UUID) (policyTemplates *model.PolicyTemplate, err error) {
	policyTemplate, err := u.repo.GetByID(ctx, policyTemplateID)

	if err != nil {
		return nil, err
	}

	if !policyTemplate.IsPermittedToOrganization(organizationId) {
		return nil, httpErrors.NewNotFoundError(fmt.Errorf(
			"policy template not found"),
			"PT_NOT_FOUND_POLICY_TEMPLATE", "")
	}

	if policyTemplate.IsTksTemplate() && len(policyTemplate.PermittedOrganizations) == 0 {
		organizations, err := u.organizationRepo.Fetch(ctx, nil)

		if err != nil {
			log.Errorf(ctx, "error is :%s(%T)", err.Error(), err)
		} else if organizations != nil {
			policyTemplate.PermittedOrganizations = *organizations
		}
	}

	if organizationId != nil {
		(*policyTemplate).LatestVersion = policyTemplate.Version

		var primaryClusterId string

		org, err := u.organizationRepo.Get(ctx, *organizationId)

		if err != nil {
			log.Errorf(ctx, "error is :%s(%T)", err.Error(), err)
		} else {
			// 에러 없이 primaryClusterId를 가져왔을 때만 처리
			primaryClusterId = org.PrimaryClusterId

			templateCR, err := policytemplate.GetTksPolicyTemplateCR(ctx, primaryClusterId, policyTemplate.ResoureName())

			if err == nil && templateCR != nil {
				(*policyTemplate).CurrentVersion = templateCR.Spec.Version
			} else if errors.IsNotFound(err) || templateCR == nil {
				// 템플릿이 존재하지 않으면 최신 버전으로 배포되므로
				(*policyTemplate).CurrentVersion = policyTemplate.LatestVersion
			} else {
				// 통신 실패 등 기타 에러, 버전을 세팅하지 않아 에러임을 알수 있도록 로그를 남김
				log.Errorf(ctx, "error is :%s(%T)", err.Error(), err)
			}
		}
	}

	return policyTemplate, nil
}

func (u *PolicyTemplateUsecase) Update(ctx context.Context, organizationId *string, policyTemplateId uuid.UUID, templateName *string, description *string, severity *string, deprecated *bool, permittedOrganizationIds *[]string) (err error) {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return httpErrors.NewBadRequestError(fmt.Errorf("invalid token"), "A_INVALID_TOKEN", "")
	}

	policyTemplate, err := u.repo.GetByID(ctx, policyTemplateId)
	if err != nil {
		return httpErrors.NewNotFoundError(err, "PT_FAILED_FETCH_POLICY_TEMPLATE", "")
	}

	if policyTemplate == nil {
		return httpErrors.NewBadRequestError(fmt.Errorf(
			"failed to fetch policy template"),
			"PT_FAILED_FETCH_POLICY_TEMPLATE", "")
	}

	if !policyTemplate.IsPermittedToOrganization(organizationId) {
		// 다른 Organization의 템플릿을 조작하려고 함, 보안을 위해서 해당 식별자 존재 자체를 알려주면 안되므로 not found
		if *organizationId != *policyTemplate.OrganizationId {
			return httpErrors.NewNotFoundError(fmt.Errorf(
				"policy template not found"),
				"PT_NOT_FOUND_POLICY_TEMPLATE", "")
		}

		return httpErrors.NewForbiddenError(fmt.Errorf(
			"cannot update policy template"),
			"PT_NOT_PERMITTED_ON_TKS_POLICY_TEMPLATE", "")
	}

	updateMap := make(map[string]interface{})

	// 기존 이름과 같은 경우면 체크 필요없으므로 다른 경우에만 처리하도록 추가
	if templateName != nil && *templateName != policyTemplate.TemplateName {
		if policyTemplate.IsTksTemplate() {
			exists, err := u.repo.ExistByName(ctx, *templateName)
			if err == nil && exists {
				return httpErrors.NewBadRequestError(httpErrors.DuplicateResource, "P_INVALID_POLICY_TEMPLATE_NAME", "policy template name already exists")
			}
		} else {
			exists, err := u.repo.ExistByNameInOrganization(ctx, *organizationId, *templateName)
			if err == nil && exists {
				return httpErrors.NewBadRequestError(httpErrors.DuplicateResource, "P_INVALID_POLICY_TEMPLATE_NAME", "policy template name already exists")
			}
		}

		updateMap["template_name"] = templateName
	}

	if description != nil {
		updateMap["description"] = description
	}

	if deprecated != nil {
		updateMap["deprecated"] = deprecated
	}

	if severity != nil {
		updateMap["severity"] = severity
	}

	var newPermittedOrganizations *[]model.Organization = nil

	if permittedOrganizationIds != nil {
		permittedOrganizations := make([]model.Organization, len(*permittedOrganizationIds))

		for i, organizationId := range *permittedOrganizationIds {
			organization, err := u.organizationRepo.Get(ctx, organizationId)
			if err != nil {
				return httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "C_INVALID_ORGANIZATION_ID", "")
			}

			permittedOrganizations[i] = organization
		}
		newPermittedOrganizations = &permittedOrganizations
	} else if len(updateMap) == 0 {
		// 허용된 조직도 필드 속성도 업데이트되지 않았으므로 아무것도 업데이트할 것이 없음
		return nil
	}

	updatorId := user.GetUserId()
	updateMap["updator_id"] = updatorId

	err = u.repo.Update(ctx, policyTemplateId, updateMap, newPermittedOrganizations)
	if err != nil {
		return err
	}

	return nil
}

func (u *PolicyTemplateUsecase) Delete(ctx context.Context, organizationId *string, policyTemplateId uuid.UUID) (err error) {
	policyTemplate, err := u.repo.GetByID(ctx, policyTemplateId)

	if err != nil {
		return err
	}

	if policyTemplate == nil {
		return httpErrors.NewBadRequestError(fmt.Errorf(
			"failed to fetch policy template"),
			"PT_FAILED_FETCH_POLICY_TEMPLATE", "")
	}

	if !policyTemplate.IsPermittedToOrganization(organizationId) {
		// 다른 Organization의 템플릿을 조작하려고 함, 보안을 위해서 해당 식별자 존재 자체를 알려주면 안되므로 not found
		if *organizationId != *policyTemplate.OrganizationId {
			return httpErrors.NewNotFoundError(fmt.Errorf(
				"policy template not found"),
				"PT_NOT_FOUND_POLICY_TEMPLATE", "")
		}

		return httpErrors.NewForbiddenError(fmt.Errorf(
			"cannot delete tks policy template"),
			"PT_NOT_PERMITTED_ON_TKS_POLICY_TEMPLATE", "")
	}

	return u.repo.Delete(ctx, policyTemplateId)
}

func (u *PolicyTemplateUsecase) IsPolicyTemplateNameExist(ctx context.Context, organizationId *string, policyTemplateName string) (bool, error) {
	if organizationId == nil {
		return u.repo.ExistByName(ctx, policyTemplateName)
	}

	return u.repo.ExistByNameInOrganization(ctx, *organizationId, policyTemplateName)
}

func (u *PolicyTemplateUsecase) IsPolicyTemplateKindExist(ctx context.Context, organizationId *string, policyTemplateKind string) (bool, error) {
	if organizationId == nil {
		return u.repo.ExistByKind(ctx, policyTemplateKind)
	}

	return u.repo.ExistByKindInOrganization(ctx, *organizationId, policyTemplateKind)
}

func (u *PolicyTemplateUsecase) GetPolicyTemplateVersion(ctx context.Context, organizationId *string, policyTemplateId uuid.UUID, version string) (policyTemplateVersionsReponse *model.PolicyTemplate, err error) {
	policyTemplate, err := u.repo.GetPolicyTemplateVersion(ctx, policyTemplateId, version)

	if err != nil {
		return nil, err
	}

	if !policyTemplate.IsPermittedToOrganization(organizationId) {
		return nil, httpErrors.NewNotFoundError(fmt.Errorf(
			"policy template not found"),
			"PT_NOT_FOUND_POLICY_TEMPLATE", "")
	}

	if policyTemplate.IsTksTemplate() && len(policyTemplate.PermittedOrganizations) == 0 {
		organizations, err := u.organizationRepo.Fetch(ctx, nil)

		if err != nil {
			log.Errorf(ctx, "error is :%s(%T)", err.Error(), err)
		} else if organizations != nil {
			policyTemplate.PermittedOrganizations = *organizations
		}
	}

	return policyTemplate, nil
}

func (u *PolicyTemplateUsecase) ListPolicyTemplateVersions(ctx context.Context, organizationId *string, policyTemplateId uuid.UUID) (policyTemplateVersionsReponse *domain.ListPolicyTemplateVersionsResponse, err error) {
	policyTemplate, err := u.repo.GetByID(ctx, policyTemplateId)

	if err != nil {
		return nil, err
	}

	if !policyTemplate.IsPermittedToOrganization(organizationId) {
		return nil, httpErrors.NewNotFoundError(fmt.Errorf(
			"policy template not found"),
			"PT_NOT_FOUND_POLICY_TEMPLATE", "")
	}

	return u.repo.ListPolicyTemplateVersions(ctx, policyTemplateId)
}

func (u *PolicyTemplateUsecase) DeletePolicyTemplateVersion(ctx context.Context, organizationId *string, policyTemplateId uuid.UUID, version string) (err error) {
	policyTemplate, err := u.repo.GetPolicyTemplateVersion(ctx, policyTemplateId, version)

	if err != nil {
		return err
	}

	if policyTemplate == nil {
		return httpErrors.NewBadRequestError(fmt.Errorf(
			"failed to fetch policy template"),
			"PT_FAILED_FETCH_POLICY_TEMPLATE", "")
	}

	if !policyTemplate.IsPermittedToOrganization(organizationId) {
		// 다른 Organization의 템플릿을 조작하려고 함, 보안을 위해서 해당 식별자 존재 자체를 알려주면 안되므로 not found
		if *organizationId != *policyTemplate.OrganizationId {
			return httpErrors.NewNotFoundError(fmt.Errorf(
				"policy template version not found"),
				"PT_NOT_FOUND_POLICY_TEMPLATE", "")
		}

		return httpErrors.NewForbiddenError(fmt.Errorf(
			"cannot delete tks policy template version"),
			"PT_NOT_PERMITTED_ON_TKS_POLICY_TEMPLATE", "")
	}

	return u.repo.DeletePolicyTemplateVersion(ctx, policyTemplateId, version)
}

func (u *PolicyTemplateUsecase) CreatePolicyTemplateVersion(ctx context.Context, organizationId *string, policyTemplateId uuid.UUID, newVersion string, schema []*domain.ParameterDef, rego string, libs []string,
	syncKinds *[]string, syncJson *string) (version string, err error) {
	policyTemplate, err := u.repo.GetByID(ctx, policyTemplateId)

	if err != nil {
		return "", err
	}

	if syncKinds != nil && syncJson != nil {
		return "", httpErrors.NewBadRequestError(httpErrors.DuplicateResource, "PT_INVALID_SYNC", "both sync kinds and json specified")
	}

	if syncKinds != nil {
		if _, err := policytemplate.CheckAndConvertToSyncData(*syncKinds); err != nil {
			return "", httpErrors.NewBadRequestError(httpErrors.DuplicateResource, "PT_INVALID_SYNC", "invalid sync kind")
		}
	}

	if syncJson != nil {
		result, err := policytemplate.ParseAndCheckSyncData(*syncJson)
		if err != nil {
			return "", httpErrors.NewBadRequestError(httpErrors.DuplicateResource, "PT_INVALID_SYNC", "invalid sync json")
		}

		formatted, err := policytemplate.MarshalSyncData(result)

		if err == nil {
			syncJson = &formatted
		}
	}

	compileResult, err := u.RegoCompile(&domain.RegoCompileRequest{Rego: rego, Libs: libs}, false)

	if err != nil {
		return "", httpErrors.NewBadRequestError(fmt.Errorf("rego compile error"), "PT_INVALID_REGO_SYNTAX", "")
	}

	if len(compileResult.Errors) > 0 {
		compileErrors := []string{}

		for _, errorResult := range compileResult.Errors {
			compileErrors = append(compileErrors, errorResult.Text)
		}

		detail := strings.Join(compileErrors, "\n")

		return "", httpErrors.NewBadRequestError(fmt.Errorf("rego compile error"), "PT_INVALID_REGO_SYNTAX", detail)
	}

	if policyTemplate == nil {
		return "", httpErrors.NewBadRequestError(fmt.Errorf(
			"failed to fetch policy template"),
			"PT_FAILED_FETCH_POLICY_TEMPLATE", "")
	}

	if !policyTemplate.IsPermittedToOrganization(organizationId) {
		// 다른 Organization의 템플릿을 조작하려고 함, 보안을 위해서 해당 식별자 존재 자체를 알려주면 안되므로 not found
		if *organizationId != *policyTemplate.OrganizationId {
			return "", httpErrors.NewNotFoundError(fmt.Errorf(
				"policy template version not found"),
				"PT_NOT_FOUND_POLICY_TEMPLATE", "")
		}

		return "", httpErrors.NewForbiddenError(fmt.Errorf(
			"cannot crate tks policy template version"),
			"PT_NOT_PERMITTED_ON_TKS_POLICY_TEMPLATE", "")
	}

	if err := policytemplate.ValidateParamDefs(policyTemplate.ParametersSchema); err != nil {
		return "", httpErrors.NewBadRequestError(err, "PT_INVALID_PARAMETER_SCHEMA", "")
	}

	rego = policytemplate.FormatRegoCode(rego)
	libs = policytemplate.FormatLibCode(libs)

	return u.repo.CreatePolicyTemplateVersion(ctx, policyTemplateId, newVersion, schema, rego, libs, syncKinds, syncJson)
}

func (u *PolicyTemplateUsecase) RegoCompile(request *domain.RegoCompileRequest, parseParameter bool) (response *domain.RegoCompileResponse, err error) {
	response = &domain.RegoCompileResponse{}
	response.Errors = []domain.RegoCompieError{}

	compiler, err := policytemplate.CompileRegoWithLibs(request.Rego, request.Libs)

	if err != nil {
		response.Errors = append(response.Errors, domain.RegoCompieError{
			Status:  400,
			Code:    "PT_FAILED_TO_LOAD_REGO_MODULE",
			Message: "failed to load rego module",
			Text:    err.Error(),
		})
	} else if compiler.Failed() {
		for _, compileError := range compiler.Errors {
			response.Errors = append(response.Errors, domain.RegoCompieError{
				Status:  400,
				Code:    "PT_INVALID_REGO_SYNTAX",
				Message: "invalid rego syntax",
				Text: fmt.Sprintf("[%d:%d] %s",
					compileError.Location.Row, compileError.Location.Col,
					compileError.Message),
			})
		}
	}

	if len(response.Errors) > 0 {
		return response, nil
	}

	if parseParameter {
		// 효율적인 파라미터 추출을 위한 머지
		modules, err := policytemplate.MergeAndCompileRegoWithLibs(request.Rego, request.Libs)
		if err != nil {
			response.Errors = append(response.Errors, domain.RegoCompieError{
				Status:  400,
				Code:    "PT_FAILED_TO_LOAD_REGO_MODULE",
				Message: "failed to load merged rego module",
				Text:    err.Error(),
			})
		} else {
			response.ParametersSchema = policytemplate.ExtractParameter(modules)
		}
	}

	return response, nil
}

func (u *PolicyTemplateUsecase) ListPolicyTemplateStatistics(ctx context.Context, organizationId *string, policyTemplateId uuid.UUID) (statistics []model.UsageCount, err error) {
	return u.policyRepo.GetUsageCountByTemplateId(ctx, organizationId, policyTemplateId)
}

func (u *PolicyTemplateUsecase) GetPolicyTemplateDeploy(ctx context.Context, organizationId *string, policyTemplateId uuid.UUID) (deployVersions domain.GetPolicyTemplateDeployResponse, err error) {
	policyTemplate, err := u.repo.GetByID(ctx, policyTemplateId)

	if err != nil {
		return deployVersions, err
	}

	if !policyTemplate.IsPermittedToOrganization(organizationId) {
		return deployVersions, httpErrors.NewNotFoundError(fmt.Errorf(
			"policy template not found"),
			"PT_NOT_FOUND_POLICY_TEMPLATE", "")
	}

	if organizationId == nil {
		organizations, err := u.organizationRepo.Fetch(ctx, nil)
		if err != nil {
			return deployVersions, err
		}

		deployVersions.DeployVersion = map[string]string{}
		for _, organization := range *organizations {
			tksPolicyTemplate, err := policytemplate.GetTksPolicyTemplateCR(ctx, organization.PrimaryClusterId, strings.ToLower(policyTemplate.Kind))
			if err != nil {
				log.Errorf(ctx, "error is :%s(%T)", err.Error(), err)
				continue
			}

			for clusterId, status := range tksPolicyTemplate.Status.TemplateStatus {
				deployVersions.DeployVersion[clusterId] = status.Version
			}
		}

		return deployVersions, nil
	}

	organization, err := u.organizationRepo.Get(ctx, *organizationId)
	if err != nil {
		return deployVersions, err
	}

	tksPolicyTemplate, err := policytemplate.GetTksPolicyTemplateCR(ctx, organization.PrimaryClusterId, strings.ToLower(policyTemplate.Kind))
	if err != nil {
		log.Errorf(ctx, "error is :%s(%T)", err.Error(), err)

		return deployVersions, httpErrors.NewInternalServerError(err, "P_FAILED_TO_CALL_KUBERNETES", "")
	}

	for clusterId, status := range tksPolicyTemplate.Status.TemplateStatus {
		deployVersions.DeployVersion[clusterId] = status.Version
	}

	return deployVersions, nil
}

func (u *PolicyTemplateUsecase) ExtractPolicyParameters(ctx context.Context, organizationId *string, policyTemplateId uuid.UUID, version string, rego string, libs []string) (response *domain.RegoCompileResponse, err error) {
	policyTemplate, err := u.repo.GetPolicyTemplateVersion(ctx, policyTemplateId, version)

	if err != nil {
		return nil, err
	}

	if policyTemplate == nil {
		return nil, httpErrors.NewBadRequestError(fmt.Errorf(
			"failed to fetch policy template"),
			"PT_FAILED_FETCH_POLICY_TEMPLATE", "")
	}

	if !policyTemplate.IsPermittedToOrganization(organizationId) {
		// 다른 Organization의 템플릿을 조작하려고 함, 보안을 위해서 해당 식별자 존재 자체를 알려주면 안되므로 not found
		if *organizationId != *policyTemplate.OrganizationId {
			return nil, httpErrors.NewNotFoundError(fmt.Errorf(
				"policy template version not found"),
				"PT_NOT_FOUND_POLICY_TEMPLATE", "")
		}
	}

	response = &domain.RegoCompileResponse{}
	response.Errors = []domain.RegoCompieError{}

	compiler, err := policytemplate.CompileRegoWithLibs(rego, libs)

	if err != nil {
		response.Errors = append(response.Errors, domain.RegoCompieError{
			Status:  400,
			Code:    "PT_FAILED_TO_LOAD_REGO_MODULE",
			Message: "failed to load rego module",
			Text:    err.Error(),
		})
		return response, nil
	} else if compiler.Failed() {
		for _, compileError := range compiler.Errors {
			response.Errors = append(response.Errors, domain.RegoCompieError{
				Status:  400,
				Code:    "PT_INVALID_REGO_SYNTAX",
				Message: "invalid rego syntax",
				Text: fmt.Sprintf("[%d:%d] %s",
					compileError.Location.Row, compileError.Location.Col,
					compileError.Message),
			})
		}
		return response, nil
	}

	if len(response.Errors) > 0 {
		return response, nil
	}

	modules, err := policytemplate.MergeAndCompileRegoWithLibs(rego, libs)
	if err != nil {
		response.Errors = append(response.Errors, domain.RegoCompieError{
			Status:  400,
			Code:    "PT_FAILED_TO_LOAD_REGO_MODULE",
			Message: "failed to load merged rego module",
			Text:    err.Error(),
		})
	}

	extractedParamDefs := policytemplate.ExtractParameter(modules)
	paramdefs := policyTemplate.ParametersSchema

	newParams, err := policytemplate.GetNewExtractedParamDefs(paramdefs, extractedParamDefs)

	response.ParametersSchema = []*domain.ParameterDef{}

	if err != nil {
		response.Errors = append(response.Errors, domain.RegoCompieError{
			Status:  400,
			Code:    "PT_NOT_COMPATIBLE_PARAMS",
			Message: "new parameters are not compatible to current version",
			Text:    err.Error(),
		})
	} else {
		response.ParametersSchema = append(response.ParametersSchema, paramdefs...)
		response.ParametersSchema = append(response.ParametersSchema, newParams...)
	}

	return response, nil
}

func (u *PolicyTemplateUsecase) AddPermittedPolicyTemplatesForOrganization(ctx context.Context, organizationId string, policyTemplateIds []uuid.UUID) (err error) {
	policyTemplates := make([]model.PolicyTemplate, len(policyTemplateIds))

	for i, policyTemplateId := range policyTemplateIds {
		result, err := u.repo.GetByID(ctx, policyTemplateId)

		if err != nil || result == nil {
			return httpErrors.NewBadRequestError(fmt.Errorf(
				"failed to fetch policy template"),
				"PT_FAILED_FETCH_POLICY_TEMPLATE", "")
		}

		if result.IsOrganizationTemplate() {
			return httpErrors.NewBadRequestError(fmt.Errorf(
				"failed to permit organization to organization policy template"),
				"PT_FAILED_TO_PERMIT_ORG_TEMPLATE", "")
		}

		policyTemplates[i] = *result
	}

	return u.organizationRepo.AddPermittedPolicyTemplatesByID(ctx, organizationId, policyTemplates)
}

func (u *PolicyTemplateUsecase) UpdatePermittedPolicyTemplatesForOrganization(ctx context.Context, organizationId string, policyTemplateIds []uuid.UUID) (err error) {
	policyTemplates := make([]model.PolicyTemplate, len(policyTemplateIds))

	for i, policyTemplateId := range policyTemplateIds {
		result, err := u.repo.GetByID(ctx, policyTemplateId)

		if err != nil || result == nil {
			return httpErrors.NewBadRequestError(fmt.Errorf(
				"failed to fetch policy template"),
				"PT_FAILED_FETCH_POLICY_TEMPLATE", "")
		}

		if result.IsOrganizationTemplate() {
			return httpErrors.NewBadRequestError(fmt.Errorf(
				"failed to permit organization to organization policy template"),
				"PT_FAILED_TO_PERMIT_ORG_TEMPLATE", "")
		}

		policyTemplates[i] = *result
	}

	return u.organizationRepo.UpdatePermittedPolicyTemplatesByID(ctx, organizationId, policyTemplates)
}

func (u *PolicyTemplateUsecase) DeletePermittedPolicyTemplatesForOrganization(ctx context.Context, organizationId string, policyTemplateIds []uuid.UUID) (err error) {
	return u.organizationRepo.DeletePermittedPolicyTemplatesByID(ctx, organizationId, policyTemplateIds)
}
