package usecase

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/open-policy-agent/opa/ast"
	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
)

type IPolicyTemplateUsecase interface {
	Create(ctx context.Context, policyTemplate model.PolicyTemplate) (policyTemplateId uuid.UUID, err error)
	Fetch(ctx context.Context, pg *pagination.Pagination) (policyTemplates []model.PolicyTemplate, err error)
	Update(ctx context.Context, policyTemplateId uuid.UUID, templateName *string, description *string,
		severity *string, deorecated *bool, permittedOrganizationIds *[]string) (err error)
	Get(ctx context.Context, policyTemplateId uuid.UUID) (policyTemplates *model.PolicyTemplate, err error)
	Delete(ctx context.Context, policyTemplateId uuid.UUID) (err error)
	IsPolicyTemplateNameExist(ctx context.Context, policyTemplateName string) (bool, error)
	IsPolicyTemplateKindExist(ctx context.Context, policyTemplateKind string) (bool, error)
	GetPolicyTemplateVersion(ctx context.Context, policyTemplateId uuid.UUID, version string) (policyTemplateVersionsReponse *model.PolicyTemplate, err error)
	ListPolicyTemplateVersions(ctx context.Context, policyTemplateId uuid.UUID) (policyTemplateVersionsReponse *domain.ListPolicyTemplateVersionsResponse, err error)
	DeletePolicyTemplateVersion(ctx context.Context, policyTemplateId uuid.UUID, version string) (err error)
	CreatePolicyTemplateVersion(ctx context.Context, policyTemplateId uuid.UUID, newVersion string, schema []domain.ParameterDef, rego string, libs []string) (version string, err error)

	RegoCompile(request *domain.RegoCompileRequest, parseParameter bool) (response *domain.RegoCompileResponse, err error)

	UpdatePermittedOrganizations(ctx context.Context,
		policyTemplate *model.PolicyTemplate, out *domain.PolicyTemplateResponse) error
	UpdatePermittedOrganizationsForList(ctx context.Context,
		policyTemplates *[]model.PolicyTemplate, outs *[]domain.PolicyTemplateResponse) error
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

func (u *PolicyTemplateUsecase) Create(ctx context.Context, dto model.PolicyTemplate) (policyTemplateId uuid.UUID, err error) {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return uuid.Nil, httpErrors.NewUnauthorizedError(fmt.Errorf("invalid token"), "A_INVALID_TOKEN", "")
	}

	exists, err := u.repo.ExistByName(ctx, dto.TemplateName)
	if err == nil && exists {
		return uuid.Nil, httpErrors.NewBadRequestError(httpErrors.DuplicateResource, "PT_CREATE_ALREADY_EXISTED_NAME", "policy template name already exists")
	}

	exists, err = u.repo.ExistByKind(ctx, dto.Kind)
	if err == nil && exists {
		return uuid.Nil, httpErrors.NewBadRequestError(httpErrors.DuplicateResource, "PT_CREATE_ALREADY_EXISTED_KIND", "policy template kind already exists")
	}

	dto.PermittedOrganizations = make([]model.Organization, len(dto.PermittedOrganizationIds))
	for i, organizationId := range dto.PermittedOrganizationIds {

		organization, err := u.organizationRepo.Get(ctx, organizationId)
		if err != nil {
			return uuid.Nil, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "C_INVALID_ORGANIZATION_ID", "")
		}
		dto.PermittedOrganizations[i] = organization
	}

	userId := user.GetUserId()
	dto.CreatorId = &userId

	// 시스템 템플릿 속성 설정
	dto.Type = "tks"
	dto.Mandatory = false

	id, err := u.repo.Create(ctx, dto)

	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func (u *PolicyTemplateUsecase) Fetch(ctx context.Context, pg *pagination.Pagination) (policyTemplates []model.PolicyTemplate, err error) {
	policyTemplates, err = u.repo.Fetch(ctx, pg)

	if err != nil {
		return nil, err
	}

	return policyTemplates, nil
}

func (u *PolicyTemplateUsecase) UpdatePermittedOrganizations(ctx context.Context,
	policyTemplate *model.PolicyTemplate, out *domain.PolicyTemplateResponse) error {
	organizations, err := u.organizationRepo.Fetch(ctx, nil)

	if err != nil {
		return err
	}

	u.updatePermittedOrganizations(ctx, organizations, policyTemplate, out)

	return nil
}

func (u *PolicyTemplateUsecase) UpdatePermittedOrganizationsForList(ctx context.Context,
	policyTemplates *[]model.PolicyTemplate, outs *[]domain.PolicyTemplateResponse) error {

	organizations, err := u.organizationRepo.Fetch(ctx, nil)

	if err != nil {
		return err
	}

	results := *outs

	for i, policyTemplate := range *policyTemplates {
		u.updatePermittedOrganizations(ctx, organizations, &policyTemplate, &results[i])
	}

	return nil
}

// 모든 조직 목록에 대해 허용 여부 업데이트
func (u *PolicyTemplateUsecase) updatePermittedOrganizations(ctx context.Context, organizations *[]model.Organization, policyTemplate *model.PolicyTemplate, out *domain.PolicyTemplateResponse) {
	if policyTemplate == nil || organizations == nil || out == nil {
		return
	}

	// 정책 템플릿에서 허용된 조직 목록이 없다는 것은 모든 조직이 사용할 수 있음을 의미함
	allPermitted := len(policyTemplate.PermittedOrganizationIds) == 0

	// 허용된 조직 포함 여부를 효율적으로 처리하기 위해 ID 리스트를 셋으로 변환
	permittedOrganizationIdSet := mapset.NewSet(policyTemplate.PermittedOrganizationIds...)

	out.PermittedOrganizations = make([]domain.PermittedOrganization, len(*organizations))

	for i, organization := range *organizations {
		permitted := allPermitted || permittedOrganizationIdSet.ContainsOne(organization.ID)

		out.PermittedOrganizations[i] = domain.PermittedOrganization{
			OrganizationId:   organization.ID,
			OrganizationName: organization.Name,
			Permitted:        permitted,
		}
	}
}

func (u *PolicyTemplateUsecase) Get(ctx context.Context, policyTemplateID uuid.UUID) (policyTemplates *model.PolicyTemplate, err error) {
	policyTemplate, err := u.repo.GetByID(ctx, policyTemplateID)

	if err != nil {
		return nil, err
	}

	return policyTemplate, nil
}

func (u *PolicyTemplateUsecase) Update(ctx context.Context, policyTemplateId uuid.UUID, templateName *string, description *string, severity *string, deprecated *bool, permittedOrganizationIds *[]string) (err error) {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return httpErrors.NewBadRequestError(fmt.Errorf("invalid token"), "A_INVALID_TOKEN", "")
	}

	_, err = u.repo.GetByID(ctx, policyTemplateId)
	if err != nil {
		return httpErrors.NewNotFoundError(err, "PT_FAILED_FETCH_POLICY_TEMPLATE", "")
	}

	updateMap := make(map[string]interface{})

	if templateName != nil {
		exists, err := u.repo.ExistByName(ctx, *templateName)
		if err == nil && exists {
			return httpErrors.NewBadRequestError(httpErrors.DuplicateResource, "P_INVALID_POLICY_TEMPLATE_NAME", "policy template name already exists")
		}
		updateMap["name"] = templateName
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

func (u *PolicyTemplateUsecase) Delete(ctx context.Context, policyTemplateId uuid.UUID) (err error) {
	return u.repo.Delete(ctx, policyTemplateId)
}

func (u *PolicyTemplateUsecase) IsPolicyTemplateNameExist(ctx context.Context, policyTemplateName string) (bool, error) {
	return u.repo.ExistByName(ctx, policyTemplateName)
}

func (u *PolicyTemplateUsecase) IsPolicyTemplateKindExist(ctx context.Context, policyTemplateKind string) (bool, error) {
	return u.repo.ExistByKind(ctx, policyTemplateKind)
}

func (u *PolicyTemplateUsecase) GetPolicyTemplateVersion(ctx context.Context, policyTemplateId uuid.UUID, version string) (policyTemplateVersionsReponse *model.PolicyTemplate, err error) {
	policyTemplate, err := u.repo.GetPolicyTemplateVersion(ctx, policyTemplateId, version)

	if err != nil {
		return nil, err
	}

	return policyTemplate, nil
}

func (*PolicyTemplateUsecase) updatePermittedOrganizations(ctx context.Context, organizations *[]model.Organization, permittedOrgIdSet map[string]string, policyTemplate *model.PolicyTemplate) {
	// 허용리스트가 비어있으면 모든 Org에 대해서 허용
	permitted := len(permittedOrgIdSet) == 0

	log.Info(ctx, "CHECK HERE ", permitted)

	for _, organization := range *organizations {

		_, ok := permittedOrgIdSet[organization.ID]

		if !ok {
			policyTemplate.PermittedOrganizations = append(
				policyTemplate.PermittedOrganizations,
				// ktkfree : 역시 이부분 확인 부탁 드립니다.
				/*
					domain.PermittedOrganization{
						OrganizationId:   organization.ID,
						OrganizationName: organization.Name,
						Permitted:        permitted,
					}
				*/
				model.Organization{},
			)
		}
	}
}

func (*PolicyTemplateUsecase) getPermittedOrganiationIdSet(ctx context.Context, policyTemplate *model.PolicyTemplate) map[string]string {
	permittedOrgIdSet := make(map[string]string)

	for _, permittedOrg := range policyTemplate.PermittedOrganizations {
		// Set 처리를 위해서 키만 사용, 값은 아무거나

		// ktkfree : 이부분 확인 부탁 드립니다.
		//
		//permittedOrgIdSet[permittedOrg.OrganizationId] = "1"
		log.Info(ctx, "CHECK HERE ", permittedOrg)
	}
	return permittedOrgIdSet
}

func (u *PolicyTemplateUsecase) ListPolicyTemplateVersions(ctx context.Context, policyTemplateId uuid.UUID) (policyTemplateVersionsReponse *domain.ListPolicyTemplateVersionsResponse, err error) {
	return u.repo.ListPolicyTemplateVersions(ctx, policyTemplateId)
}

func (u *PolicyTemplateUsecase) DeletePolicyTemplateVersion(ctx context.Context, policyTemplateId uuid.UUID, version string) (err error) {
	return u.repo.DeletePolicyTemplateVersion(ctx, policyTemplateId, version)
}

func (u *PolicyTemplateUsecase) CreatePolicyTemplateVersion(ctx context.Context, policyTemplateId uuid.UUID, newVersion string, schema []domain.ParameterDef, rego string, libs []string) (version string, err error) {
	return u.repo.CreatePolicyTemplateVersion(ctx, policyTemplateId, newVersion, schema, rego, libs)
}

func (u *PolicyTemplateUsecase) RegoCompile(request *domain.RegoCompileRequest, parseParameter bool) (response *domain.RegoCompileResponse, err error) {
	modules := map[string]*ast.Module{}

	response = &domain.RegoCompileResponse{}
	response.Errors = []domain.RegoCompieError{}

	mod, err := ast.ParseModuleWithOpts("rego", request.Rego, ast.ParserOptions{})
	if err != nil {
		return nil, err
	}
	modules["rego"] = mod

	compiler := ast.NewCompiler()
	compiler.Compile(modules)

	if compiler.Failed() {
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

	if parseParameter {
		response.ParametersSchema = extractParameter(request.Rego)
	}

	return response, nil
}

func extractParameter(rego string) []*domain.ParameterDef {
	pattern := `input\.parameters\.[\w\.\[\]]+`

	prefix := "input.parameters."

	// Compile the regex pattern
	regex := regexp.MustCompile(pattern)

	matches := regex.FindAllString(rego, -1)

	defStore := NewParamDefStore()

	for _, match := range matches {
		remainder := match[len(prefix):]

		// 문법 변환: aa["a"]["B"][_]->aa.a.B[_]
		regex := regexp.MustCompile(`\[\"(\w+)\"\]`)
		remainder = regex.ReplaceAllString(remainder, ".$1")

		params := strings.Split(remainder, ".")

		if len(params) == 0 {
			continue
		}

		defStore.AddDefinition(params)
	}

	return defStore.store
}

type ParamDefStore struct {
	store []*domain.ParameterDef
}

func NewParamDefStore() *ParamDefStore {
	return &ParamDefStore{store: []*domain.ParameterDef{}}
}

func (s *ParamDefStore) GetStore() []*domain.ParameterDef {
	return s.store
}

func (s *ParamDefStore) AddDefinition(params []string) {
	init := &s.store

	for i, param := range params {
		isLast := i == len(params)-1

		key := findKey(s.store, param)

		if key == nil {
			key = createKey(param, isLast)
			*init = append(*init, key)
		}

		init = &key.Children
	}
}

func findKey(defs []*domain.ParameterDef, key string) *domain.ParameterDef {
	for _, def := range defs {
		if def.Key == key || def.Key+"[_]" == key {
			return def
		}
	}

	return nil
}

func createKey(key string, isLast bool) *domain.ParameterDef {
	finalType := "any"

	pKey := key
	isArray := false

	if strings.HasSuffix(pKey, "[_]") {
		pKey, _ = strings.CutSuffix(pKey, "[_]")
		isArray = true
	}

	if isLast {
		if isArray {
			finalType = "any[]"
		} else {
			finalType = "any"
		}
	} else {
		if isArray {
			finalType = "object[]"
		} else {
			finalType = "object"
		}
	}

	newDef := &domain.ParameterDef{
		Key:      pKey,
		Type:     finalType,
		Children: []*domain.ParameterDef{},
	}

	return newDef
}
