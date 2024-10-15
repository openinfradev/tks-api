package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	policytemplate "github.com/openinfradev/tks-api/internal/policy-template"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/internal/serializer"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/utils/strings/slices"
)

type IPolicyUsecase interface {
	Create(ctx context.Context, organizationId string, dto model.Policy) (policyId uuid.UUID, err error)
	Update(ctx context.Context, organizationId string, policyId uuid.UUID,
		mandatory *bool, policyName *string, description *string, templateId *uuid.UUID, enforcementAction *string,
		parameters *string, match *domain.Match, matchYaml *string, targetClusterIds *[]string) (err error)
	Delete(ctx context.Context, organizationId string, policyId uuid.UUID) (err error)
	Get(ctx context.Context, organizationId string, policyId uuid.UUID) (policy *model.Policy, err error)
	GetForEdit(ctx context.Context, organizationId string, policyId uuid.UUID) (policy *model.Policy, err error)
	Fetch(ctx context.Context, organizationId string, pg *pagination.Pagination, filledParameter bool) (*[]model.Policy, error)
	IsPolicyIdExist(ctx context.Context, organizationId string, policyId uuid.UUID) (exists bool, err error)
	IsPolicyNameExist(ctx context.Context, organizationId string, policyName string) (exists bool, err error)
	IsPolicyResourceNameExist(ctx context.Context, organizationId string, policyResourceName string) (exists bool, err error)
	UpdatePolicyTargetClusters(ctx context.Context, organizationId string, policyId uuid.UUID, currentClusterIds []string, targetClusterIds []string) (err error)
	SetMandatoryPolicies(ctx context.Context, organizationId string, mandatoryPolicyIds []uuid.UUID, nonMandatoryPolicyIds []uuid.UUID) (err error)
	GetMandatoryPolicies(ctx context.Context, organizationId string) (response *domain.GetMandatoryPoliciesResponse, err error)
	ListStackPolicyStatus(ctx context.Context, clusterId string, pg *pagination.Pagination) (policyStatuses []domain.StackPolicyStatusResponse, err error)
	GetStackPolicyTemplateStatus(ctx context.Context, clusterId string, policyTemplateId uuid.UUID) (clusterPolicyTemplateStatusResponse *domain.GetStackPolicyTemplateStatusResponse, err error)
	UpdateStackPolicyTemplateStatus(ctx context.Context, clusterId string, policyTemplateId uuid.UUID,
		currentVersion string, targetVerson string) (err error)
	GetPolicyStatistics(ctx context.Context, organizationId string) (response *domain.PolicyStatisticsResponse, err error)
	AddPoliciesForClusterID(ctx context.Context, organizationId string, clusterId domain.ClusterId, policyIds []uuid.UUID) (err error)
	UpdatePoliciesForClusterID(ctx context.Context, organizationId string, clusterId domain.ClusterId, policyIds []uuid.UUID) (err error)
	DeletePoliciesForClusterID(ctx context.Context, organizationId string, clusterId domain.ClusterId, policyIds []uuid.UUID) (err error)
	GetStackPolicyStatistics(ctx context.Context, organizationId string, clusterId domain.ClusterId) (statistics *domain.StackPolicyStatistics, err error)
	GetPolicyIDsByClusterID(ctx context.Context, clusterId domain.ClusterId) (out *[]uuid.UUID, err error)
}

type PolicyUsecase struct {
	organizationRepo repository.IOrganizationRepository
	clusterRepo      repository.IClusterRepository
	templateRepo     repository.IPolicyTemplateRepository
	repo             repository.IPolicyRepository
}

func NewPolicyUsecase(r repository.Repository) IPolicyUsecase {
	return &PolicyUsecase{
		repo:             r.Policy,
		templateRepo:     r.PolicyTemplate,
		organizationRepo: r.Organization,
		clusterRepo:      r.Cluster,
	}
}

func randomResouceName(kind string) string {
	uuid := uuid.New().String()
	idStr := strings.Split(uuid, "-")
	return strings.ToLower(kind) + "-" + idStr[len(idStr)-1]
}

func (u *PolicyUsecase) Create(ctx context.Context, organizationId string, dto model.Policy) (policyId uuid.UUID, err error) {
	dto.OrganizationId = organizationId

	user, ok := request.UserFrom(ctx)
	if !ok {
		return uuid.Nil, httpErrors.NewUnauthorizedError(fmt.Errorf("invalid token"), "A_INVALID_TOKEN", "")
	}

	exists, err := u.repo.ExistByName(ctx, dto.OrganizationId, dto.PolicyName)
	if err != nil {
		return uuid.Nil, err
	}

	if exists {
		return uuid.Nil, httpErrors.NewBadRequestError(httpErrors.DuplicateResource, "P_CREATE_ALREADY_EXISTED_NAME", "policy name already exists")
	}

	policyTemplate, err := u.templateRepo.GetByID(ctx, dto.TemplateId)
	if err != nil {
		return uuid.Nil, err
	}

	if policyTemplate == nil {
		return uuid.Nil, httpErrors.NewBadRequestError(httpErrors.DuplicateResource, "PT_POlICY_TEMPLATE_NOT_FOUND", "policy template not found")
	}

	if len(dto.PolicyResourceName) == 0 {
		dto.PolicyResourceName = randomResouceName(policyTemplate.Kind)
	}

	exists, err = u.repo.ExistByResourceName(ctx, dto.OrganizationId, dto.PolicyResourceName)
	if err != nil {
		return uuid.Nil, err
	}

	if exists {
		return uuid.Nil, httpErrors.NewBadRequestError(httpErrors.DuplicateResource, "P_INVALID_POLICY_RESOURCE_NAME", "policy resource name already exists")
	}

	dto.TargetClusters = make([]model.Cluster, len(dto.TargetClusterIds))
	for i, clusterId := range dto.TargetClusterIds {

		cluster, err := u.clusterRepo.Get(ctx, domain.ClusterId(clusterId))
		if err != nil {
			log.Errorf(ctx, "error is :%s(%T)", err.Error(), err)

			return uuid.Nil, httpErrors.NewBadRequestError(err, "P_FAILED_FETCH_CLUSTER", "")
		}
		dto.TargetClusters[i] = cluster
	}

	if !policyTemplate.IsPermittedToOrganization(&organizationId) {
		return uuid.Nil, httpErrors.NewNotFoundError(fmt.Errorf(
			"policy template not found"),
			"PT_NOT_FOUND_POLICY_TEMPLATE", "")
	}

	userId := user.GetUserId()
	dto.CreatorId = &userId
	dto.TemplateId = policyTemplate.ID

	organization, err := u.organizationRepo.Get(ctx, organizationId)

	if err != nil {
		log.Errorf(ctx, "error is :%s(%T)", err.Error(), err)

		return uuid.Nil, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "C_INVALID_ORGANIZATION_ID", "")
	}

	if err := policytemplate.ValidateJSONusingParamdefs(policyTemplate.ParametersSchema, dto.Parameters); err != nil {
		log.Errorf(ctx, "error is :%s(%T)", err.Error(), err)

		return uuid.Nil, httpErrors.NewBadRequestError(err, "P_INVALID_POLICY_PARAMETER", "")
	}

	primaryClusterId := organization.PrimaryClusterId

	// k8s에 label로 policyID를 기록해 주기 위해 DB 컬럼 생성 시 ID를 생성하지 않고 미리 생성
	dto.ID = uuid.New()

	err = createOrUpdatPolicy(ctx, &dto, policyTemplate, primaryClusterId)
	if err != nil {
		return uuid.Nil, err
	}

	id, err := u.repo.Create(ctx, dto)

	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func createOrUpdatPolicy(ctx context.Context, dto *model.Policy, policyTemplate *model.PolicyTemplate, primaryClusterId string) error {
	policyCR := policytemplate.PolicyToTksPolicyCR(dto)

	// DB 생성 전 policy DTO에 PolicyTemplate 필드를 넣어주면 탬플릿이 생성될 수 있으므로,
	// dto에 세팅하지 않고 변환 후 필요한 템플릿 리소스 이름을 따로 넣어 줌
	if len(policyCR.Spec.Template) == 0 {
		policyCR.Spec.Template = policyTemplate.Kind
	}

	policyTemplateCR := policytemplate.PolicyTemplateToTksPolicyTemplateCR(policyTemplate)

	exists, err := policytemplate.ExistsTksPolicyTemplateCR(ctx, primaryClusterId, policyTemplateCR.Name)
	if err != nil {
		log.Errorf(ctx, "error is :%s(%T)", err.Error(), err)

		return httpErrors.NewInternalServerError(err, "P_FAILED_TO_CALL_KUBERNETES", "")
	}

	if !exists {
		err = policytemplate.ApplyTksPolicyTemplateCR(ctx, primaryClusterId, policyTemplateCR)

		if err != nil {
			errYaml := ""
			if policyCR != nil {
				errYaml, _ = policyTemplateCR.YAML()
			}

			log.Errorf(ctx, "error is :%s(%T), policyTemplateCR='%+v'", err.Error(), err, errYaml)

			return httpErrors.NewInternalServerError(err, "P_FAILED_TO_APPLY_KUBERNETES", "")
		}
	}

	err = policytemplate.ApplyTksPolicyCR(ctx, primaryClusterId, policyCR)

	// policyYaml, _ := policyCR.YAML()
	// policyTemplateYaml, _ := policyTemplateCR.YAML()

	// fmt.Println("-------------------- Policy --------------------")
	// fmt.Printf("%+v\n", policyYaml)
	// fmt.Println("------------------------------------------------")
	// fmt.Println("--------------- Policy Template ----------------")
	// fmt.Printf("%+v\n", policyTemplateYaml)
	// fmt.Println("------------------------------------------------")
	// fmt.Printf("%+v\n", policyTemplateCR.Spec.Targets[0].Rego)
	// fmt.Println("------------------------------------------------")

	if err != nil {
		errYaml := ""
		if policyCR != nil {
			errYaml, _ = policyCR.YAML()
		}

		log.Errorf(ctx, "error is :%s(%T), policyCR='%+v'", err.Error(), err, errYaml)

		return httpErrors.NewInternalServerError(err, "P_FAILED_TO_APPLY_KUBERNETES", "")
	}

	return nil
}

func (u *PolicyUsecase) Update(ctx context.Context, organizationId string, policyId uuid.UUID,
	mandatory *bool, policyName *string, description *string, templateId *uuid.UUID, enforcementAction *string,
	parameters *string, match *domain.Match, matchYaml *string, targetClusterIds *[]string) (err error) {

	user, ok := request.UserFrom(ctx)
	if !ok {
		return httpErrors.NewBadRequestError(fmt.Errorf("invalid token"), "A_INVALID_TOKEN", "")
	}

	policy, err := u.repo.GetByID(ctx, organizationId, policyId)
	if err != nil {
		return httpErrors.NewNotFoundError(err, "P_FAILED_FETCH_POLICY", "")
	}

	updateMap := make(map[string]interface{})

	if mandatory != nil {
		updateMap["mandatory"] = mandatory
	}

	// 정책명을 업데이트하기로 설정하였고
	if policyName != nil && policy.PolicyName != *policyName {
		exists, err := u.repo.ExistByName(ctx, organizationId, *policyName)

		if err != nil && !errors.IsNotFound(err) {
			return err
		}

		// 이름이 같은 정책이 존재하지만 아이디가 서로 다른 경우, 즉 다른 정책이 해당 이름 사용 중임
		if exists {
			return httpErrors.NewBadRequestError(httpErrors.DuplicateResource, "P_INVALID_POLICY_NAME", "policy name already exists")
		}

		// 해당 이름 사용중인 정책이 없으면 업데이트, 있으면 동일 정책이므로 업데이트 안 함
		updateMap["policy_name"] = policyName
	}

	if description != nil {
		updateMap["description"] = description
	}

	schema := policy.PolicyTemplate.ParametersSchema

	if templateId != nil {
		policyTemplate, err := u.templateRepo.GetByID(ctx, *templateId)
		schema = policyTemplate.ParametersSchema

		if err != nil {
			return err
		}

		if !policyTemplate.IsPermittedToOrganization(&organizationId) {
			return httpErrors.NewNotFoundError(fmt.Errorf(
				"policy template not found"),
				"PT_NOT_FOUND_POLICY_TEMPLATE", "")
		}

		updateMap["template_id"] = templateId
	}

	if enforcementAction != nil {
		updateMap["enforcement_action"] = enforcementAction
	}

	if parameters != nil {
		updateMap["parameters"] = parameters

		if err := policytemplate.ValidateJSONusingParamdefs(schema, *parameters); err != nil {
			log.Errorf(ctx, "error is :%s(%T)", err.Error(), err)

			return httpErrors.NewBadRequestError(err, "P_INVALID_POLICY_PARAMETER", "")
		}
	}

	if matchYaml != nil {
		updateMap["match_yaml"] = matchYaml
		updateMap["policy_match"] = nil
	} else if match != nil {
		updateMap["policy_match"] = match.JSON()
		updateMap["match_yaml"] = nil
	}

	var newTargetClusters *[]model.Cluster = nil

	if targetClusterIds != nil {
		targetClusters := make([]model.Cluster, len(*targetClusterIds))

		for i, clusterId := range *targetClusterIds {
			cluster, err := u.clusterRepo.Get(ctx, domain.ClusterId(clusterId))
			if err != nil {
				return httpErrors.NewBadRequestError(fmt.Errorf("invalid clusterId"), "C_INVALID_CLUSTER_ID", "")
			}

			targetClusters[i] = cluster
		}
		newTargetClusters = &targetClusters
	} else if len(updateMap) == 0 {
		// 허용된 조직도 필드 속성도 업데이트되지 않았으므로 아무것도 업데이트할 것이 없음
		return nil
	}

	updatorId := user.GetUserId()
	updateMap["updator_id"] = updatorId

	err = u.repo.Update(ctx, organizationId, policyId, updateMap, newTargetClusters)
	if err != nil {
		return err
	}

	if templateId != nil || enforcementAction != nil ||
		parameters != nil || match != nil || targetClusterIds != nil {

		organization, err := u.organizationRepo.Get(ctx, organizationId)

		if err != nil {
			return httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "C_INVALID_ORGANIZATION_ID", "")
		}

		policy, err := u.repo.GetByID(ctx, organizationId, policyId)
		if err != nil {
			return httpErrors.NewBadRequestError(fmt.Errorf("invalid policyId"), "C_INVALID_POLICY_ID", "")
		}

		policyCR := policytemplate.PolicyToTksPolicyCR(policy)

		err = policytemplate.ApplyTksPolicyCR(ctx, organization.PrimaryClusterId, policyCR)

		if err != nil {
			log.Errorf(ctx, "failed to apply TksPolicyCR: %v", err)
			return err
		}

		return err
	}

	return nil
}

func (u *PolicyUsecase) Delete(ctx context.Context, organizationId string, policyId uuid.UUID) (err error) {
	policy, err := u.repo.GetByID(ctx, organizationId, policyId)
	if err != nil {
		return err
	}

	organization, err := u.organizationRepo.Get(ctx, organizationId)

	if err != nil {
		return httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "C_INVALID_ORGANIZATION_ID", "")
	}

	exists, err := policytemplate.ExistsTksPolicyCR(ctx, organization.PrimaryClusterId, policy.PolicyResourceName)
	if err != nil {
		log.Errorf(ctx, "failed to check TksPolicyCR: %v", err)
		return httpErrors.NewInternalServerError(err, "P_FAILED_TO_APPLY_KUBERNETES", "")
	}

	if exists {
		err = policytemplate.DeleteTksPolicyCR(ctx, organization.PrimaryClusterId, policy.PolicyResourceName)

		if err != nil {
			log.Errorf(ctx, "failed to delete TksPolicyCR: %v", err)
			return httpErrors.NewInternalServerError(err, "P_FAILED_TO_APPLY_KUBERNETES", "")
		}
	}

	return u.repo.Delete(ctx, organizationId, policyId)
}

func (u *PolicyUsecase) Get(ctx context.Context, organizationId string, policyId uuid.UUID) (policy *model.Policy, err error) {
	return u.repo.GetByID(ctx, organizationId, policyId)
}

func (u *PolicyUsecase) GetForEdit(ctx context.Context, organizationId string, policyId uuid.UUID) (policy *model.Policy, err error) {
	policy, err = u.repo.GetByID(ctx, organizationId, policyId)

	if err != nil {
		return nil, err
	}

	policyTemplate, err := u.templateRepo.GetByID(ctx, policy.TemplateId)

	if policyTemplate != nil {
		policy.PolicyTemplate = *policyTemplate
	}

	return policy, err
}

func (u *PolicyUsecase) Fetch(ctx context.Context, organizationId string, pg *pagination.Pagination, filledParameter bool) (*[]model.Policy, error) {
	policies, err := u.repo.Fetch(ctx, organizationId, pg)

	if err != nil {
		return nil, err
	}

	// 단순 Fetch인 경우에는 Policy에 해당하는 Template만 Join해 줌
	// PolicyTemplate의 최신 버전을 조회해서 파라미터 스키마 등을 조회해서 넣어줘야 함
	if filledParameter {
		for i, policy := range *policies {
			policyTemplate, err := u.templateRepo.GetByID(ctx, policy.TemplateId)

			if err == nil && policyTemplate != nil {
				(*policies)[i].PolicyTemplate = *policyTemplate
			}
		}
	}

	return policies, err
}

func (u *PolicyUsecase) IsPolicyNameExist(ctx context.Context, organizationId string, policyName string) (exists bool, err error) {
	return u.repo.ExistByName(ctx, organizationId, policyName)
}

func (u *PolicyUsecase) IsPolicyResourceNameExist(ctx context.Context, organizationId string, policyResoName string) (exists bool, err error) {
	return u.repo.ExistByResourceName(ctx, organizationId, policyResoName)
}

func (u *PolicyUsecase) IsPolicyIdExist(ctx context.Context, organizationId string, policyId uuid.UUID) (exists bool, err error) {
	return u.repo.ExistByID(ctx, organizationId, policyId)
}

func (u *PolicyUsecase) UpdatePolicyTargetClusters(ctx context.Context, organizationId string, policyId uuid.UUID, currentClusterIds []string, targetClusterIds []string) (err error) {
	targetClusters := make([]model.Cluster, len(targetClusterIds))

	for i, clusterId := range targetClusterIds {
		cluster, err := u.clusterRepo.Get(ctx, domain.ClusterId(clusterId))
		if err != nil {
			return httpErrors.NewBadRequestError(fmt.Errorf("invalid clusterId"), "C_INVALID_CLUSTER_ID", "")
		}

		targetClusters[i] = cluster
	}

	organization, err := u.organizationRepo.Get(ctx, organizationId)

	if err != nil {
		log.Errorf(ctx, "error is :%s(%T)", err.Error(), err)

		return httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "C_INVALID_ORGANIZATION_ID", "")
	}

	err = u.repo.UpdatePolicyTargetClusters(ctx, organizationId, policyId, currentClusterIds, targetClusters)
	if err != nil {
		return err
	}

	policy, err := u.repo.GetByID(ctx, organizationId, policyId)
	if err != nil {
		return httpErrors.NewBadRequestError(fmt.Errorf("invalid policyId"), "C_INVALID_POLICY_ID", "")
	}

	policyCR := policytemplate.PolicyToTksPolicyCR(policy)

	err = policytemplate.ApplyTksPolicyCR(ctx, organization.PrimaryClusterId, policyCR)

	if err != nil {
		log.Errorf(ctx, "failed to apply TksPolicyCR: %v", err)
		return err
	}

	return nil
}

func (u *PolicyUsecase) SetMandatoryPolicies(ctx context.Context, organizationId string, mandatoryPolicyIds []uuid.UUID, nonMandatoryPolicyIds []uuid.UUID) (err error) {
	return u.repo.SetMandatoryPolicies(ctx, organizationId, mandatoryPolicyIds, nonMandatoryPolicyIds)
}

func (u *PolicyUsecase) GetMandatoryPolicies(ctx context.Context, organizationId string) (response *domain.GetMandatoryPoliciesResponse, err error) {

	var out domain.GetMandatoryPoliciesResponse

	policyTemplates, err := u.templateRepo.Fetch(ctx, nil)

	if err != nil {
		return nil, err
	}

	templateMaps := map[string]*domain.MandatoryTemplateInfo{}

	for _, policyTemplate := range policyTemplates {
		templateId := policyTemplate.ID.String()

		if slices.Contains(policyTemplate.PermittedOrganizationIds, organizationId) {
			templateMaps[templateId] = &domain.MandatoryTemplateInfo{
				TemplateName: policyTemplate.TemplateName,
				TemplateId:   templateId,
				Description:  policyTemplate.Description,
				Policies:     []domain.MandatoryPolicyInfo{},
			}
		}
	}

	policies, err := u.repo.Fetch(ctx, organizationId, nil)

	if err != nil {
		return nil, err
	}

	for _, policy := range *policies {
		template, ok := templateMaps[policy.TemplateId.String()]

		if ok {
			template.Policies = append(template.Policies, domain.MandatoryPolicyInfo{
				PolicyName:  policy.PolicyName,
				PolicyId:    policy.ID.String(),
				Description: policy.Description,
				Mandatory:   policy.Mandatory,
			})

			if policy.Mandatory {
				template.Mandatory = true
			}
		}
	}

	for _, template := range templateMaps {
		out.Templates = append(out.Templates, *template)
	}

	return &out, nil
}

func (u *PolicyUsecase) ListStackPolicyStatus(ctx context.Context, clusterId string, pg *pagination.Pagination) (policyStatuses []domain.StackPolicyStatusResponse, err error) {
	policies, err := u.repo.FetchByClusterId(ctx, clusterId, pg)

	if err != nil {
		return nil, err
	}

	// policies가 빈 목록일 수도 있으므로 policy의 organization 정보는 못 가져올 수도 있음
	// 따라서 cluster의 Organization에서 primaryClusterId를 가져옴
	cluster, err := u.clusterRepo.Get(ctx, domain.ClusterId(clusterId))
	if err != nil {
		return nil, err
	}

	primaryClusterId := cluster.Organization.PrimaryClusterId

	// tksClusterCR, err := policytemplate.GetTksClusterCR(ctx, primaryClusterId, clusterId)
	// if err != nil {
	// 	return nil, err
	// }

	result := make([]domain.StackPolicyStatusResponse, len(*policies))

	for i, policy := range *policies {
		if err := serializer.Map(ctx, policy, &result[i]); err != nil {
			continue
		}

		result[i].PolicyId = policy.ID.String()
		result[i].PolicyDescription = policy.Description
		result[i].PolicyMandatory = policy.Mandatory
		latestVersion, _ := u.templateRepo.GetLatestTemplateVersion(ctx, policy.TemplateId)

		tksPolicyTemplateCR, err := policytemplate.GetTksPolicyTemplateCR(ctx, primaryClusterId, policy.PolicyTemplate.ResoureName())
		if err != nil {
			return nil, err
		}

		result[i].TemplateCurrentVersion = tksPolicyTemplateCR.Spec.Version

		// version, ok := tksClusterCR.Status.Templates[policy.PolicyTemplate.Kind]

		// if ok {
		// 	result[i].TemplateCurrentVersion = version
		// }

		result[i].TemplateLatestVersion = latestVersion
		result[i].TemplateDescription = policy.PolicyTemplate.Description
	}

	return result, nil
}

func (u *PolicyUsecase) UpdateStackPolicyTemplateStatus(ctx context.Context, clusterId string, policyTemplateId uuid.UUID,
	currentVersion string, targetVerson string) (err error) {
	if currentVersion == targetVerson {
		// 버전 동일, 할일 없음
		return nil
	}

	latestTemplate, err := u.templateRepo.GetByID(ctx, policyTemplateId)
	if err != nil {
		return err
	}

	if targetVerson != latestTemplate.Version {
		return fmt.Errorf("targetVersion is not a latest version")
	}

	currentTemplate, err := u.templateRepo.GetPolicyTemplateVersion(ctx, policyTemplateId, currentVersion)
	if err != nil {
		return err
	}

	// 파라미터 호환성 검증, 파라미터 스키마가 동일하거나 추가된 필드만 있어야 하고 기존 필드는 이름, 타입이 유지되어야 함
	_, err = extractNewTemplateParameter(currentTemplate.ParametersSchema, latestTemplate.ParametersSchema)
	if err != nil {
		return err
	}

	cluster, err := u.clusterRepo.Get(ctx, domain.ClusterId(clusterId))
	if err != nil {
		return err
	}

	primaryClusterId := cluster.Organization.PrimaryClusterId

	resourceName := strings.ToLower(latestTemplate.Kind)

	tksPolicyTemplate, err := policytemplate.GetTksPolicyTemplateCR(ctx, primaryClusterId, resourceName)

	if err != nil {
		return err
	}

	if tksPolicyTemplate.Spec.ToLatest == nil {
		tksPolicyTemplate.Spec.ToLatest = []string{}
	}

	latestTemplateCR := policytemplate.PolicyTemplateToTksPolicyTemplateCR(latestTemplate)
	tksPolicyTemplate.Spec = latestTemplateCR.Spec

	if !slices.Contains(tksPolicyTemplate.Spec.ToLatest, clusterId) {
		tksPolicyTemplate.Spec.ToLatest = append(tksPolicyTemplate.Spec.ToLatest, clusterId)
	}

	return policytemplate.UpdateTksPolicyTemplateCR(ctx, primaryClusterId, tksPolicyTemplate)
}

func (u *PolicyUsecase) GetStackPolicyTemplateStatus(ctx context.Context, clusterId string, policyTemplateId uuid.UUID) (stackPolicyTemplateStatusResponse *domain.GetStackPolicyTemplateStatusResponse, err error) {
	policies, err := u.repo.FetchByClusterIdAndTemplaeId(ctx, clusterId, policyTemplateId)

	if err != nil {
		return nil, err
	}

	latestTemplate, err := u.templateRepo.GetByID(ctx, policyTemplateId)
	if err != nil {
		return nil, err
	}

	if latestTemplate == nil {
		return nil, httpErrors.NewBadRequestError(err, "P_FAILED_FETCH_TEMPLATE", "")
	}

	// policies가 빈 목록일 수도 있으므로 policy의 organization 정보는 못 가져올 수도 있음
	// 따라서 cluster의 Organization에서 primaryClusterId를 가져옴
	cluster, err := u.clusterRepo.Get(ctx, domain.ClusterId(clusterId))
	if err != nil {
		return nil, err
	}

	primaryClusterId := cluster.Organization.PrimaryClusterId

	// tksClusterCR, err := policytemplate.GetTksClusterCR(ctx, primaryClusterId, clusterId)
	// if err != nil {
	// 	return nil, err
	// }

	// version, ok := tksClusterCR.Status.Templates[latestTemplate.Kind]
	//
	// if !ok {
	// 	return nil, fmt.Errorf("version not found in CR")
	// }

	var version string

	tksPolicyTemplateCR, err := policytemplate.GetTksPolicyTemplateCR(ctx, primaryClusterId, latestTemplate.ResoureName())
	if err != nil {
		if errors.IsNotFound(err) {
			// CR이 배포되지 않았으므로 최신 버전이 현재 버전임(앞으로 설치될 모든 정책은 이 버전으로 설치되므로)
			version = latestTemplate.Version
		} else {
			return nil, err
		}
	} else {
		version = tksPolicyTemplateCR.Spec.Version
	}

	currentTemplate, err := u.templateRepo.GetPolicyTemplateVersion(ctx, policyTemplateId, version)
	if err != nil {
		return nil, err
	}

	updatedPolicyParameters, err := extractNewTemplateParameter(currentTemplate.ParametersSchema, latestTemplate.ParametersSchema)

	if err != nil {
		return nil, err
	}

	affectedPolicies := make([]domain.PolicyStatus, len(*policies))

	for i, policy := range *policies {
		affectedPolicies[i].PolicyId = policy.ID.String()
		affectedPolicies[i].PolicyName = policy.PolicyName

		parsed, err := parseParameter(currentTemplate.ParametersSchema, latestTemplate.ParametersSchema, policy.Parameters)

		if err != nil {
			return nil, err
		}
		affectedPolicies[i].PolicyParameters = parsed
	}

	result := domain.GetStackPolicyTemplateStatusResponse{
		TemplateName:                     currentTemplate.TemplateName,
		TemplateId:                       policyTemplateId.String(),
		TemplateDescription:              currentTemplate.Description,
		TemplateLatestVersion:            latestTemplate.Version,
		TemplateCurrentVersion:           currentTemplate.Version,
		TemplateLatestVersionReleaseDate: latestTemplate.CreatedAt,
		AffectedPolicies:                 affectedPolicies,
		UpdatedPolicyParameters:          updatedPolicyParameters,
	}

	return &result, nil
}

func (u *PolicyUsecase) GetPolicyStatistics(ctx context.Context, organizationId string) (response *domain.PolicyStatisticsResponse, err error) {
	result := domain.PolicyStatisticsResponse{}

	orgTemplateCount, err := u.templateRepo.CountOrganizationTemplate(ctx, organizationId)
	if err != nil {
		return nil, err
	}

	tksTemplateCount, err := u.templateRepo.CountTksTemplateByOrganization(ctx, organizationId)
	if err != nil {
		return nil, err
	}

	policyStatistics, err := u.repo.CountPolicyByEnforcementAction(ctx, organizationId)
	if err != nil {
		return nil, err
	}

	var policyTotal int64
	var deny int64
	var dryrun int64
	var warn int64

	for _, stat := range policyStatistics {
		switch stat.EnforcementAction {
		case "deny":
			deny = stat.Count
			policyTotal += deny
		case "dryrun":
			dryrun = stat.Count
			policyTotal += dryrun
		case "warn":
			warn = stat.Count
			policyTotal += warn
		}
	}

	result.Template = domain.TemplateCount{
		TksTemplate:          tksTemplateCount,
		OrganizationTemplate: orgTemplateCount,
		Total:                tksTemplateCount + orgTemplateCount,
	}

	policyFromOrgTemplates, err := u.templateRepo.CountPolicyFromOrganizationTemplate(ctx, organizationId)
	if err != nil {
		return nil, err
	}

	result.Policy = domain.PolicyCount{
		Deny:            deny,
		Warn:            warn,
		Dryrun:          dryrun,
		FromOrgTemplate: policyFromOrgTemplates,
		FromTksTemplate: policyTotal - policyFromOrgTemplates,
		Total:           policyTotal,
	}

	return &result, nil
}

func (u *PolicyUsecase) AddPoliciesForClusterID(ctx context.Context, organizationId string, clusterId domain.ClusterId, policyIds []uuid.UUID) (err error) {
	organization, err := u.organizationRepo.Get(ctx, organizationId)

	primaryClusterId := organization.PrimaryClusterId

	if err != nil {
		log.Errorf(ctx, "error is :%s(%T)", err.Error(), err)

		return httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "C_INVALID_ORGANIZATION_ID", "")
	}

	tkpolicies, err := policytemplate.GetTksPolicyCRs(ctx, primaryClusterId)
	if err != nil {
		log.Errorf(ctx, "error is :%s(%T)", err.Error(), err)

		return err
	}

	ids := make([]string, len(policyIds))
	for i, policyId := range policyIds {
		ids[i] = policyId.String()
	}

	// mapset.NewSet([]string{})

	for _, tkspolicy := range tkpolicies {
		policyId := tkspolicy.GetPolicyID()

		// 추가 대상
		if slices.Contains(ids, policyId) {
			// 현재 클러스터가 안 추가되어 있으면
			if !slices.Contains(tkspolicy.Spec.Clusters, string(clusterId)) {
				tkspolicy.Spec.Clusters = append(tkspolicy.Spec.Clusters, string(clusterId))

				err := policytemplate.ApplyTksPolicyCR(ctx, primaryClusterId, &tkspolicy)
				if err != nil {
					log.Errorf(ctx, "error is :%s(%T)", err.Error(), err)
				}
			}

		}
	}

	policies := []model.Policy{}

	for _, policyId := range policyIds {
		policy, err := u.repo.GetByID(ctx, organizationId, policyId)

		if err != nil {
			log.Errorf(ctx, "error is :%s(%T)", err.Error(), err)

			continue
		}

		policies = append(policies, *policy)
	}

	return u.repo.AddPoliciesForClusterID(ctx, organizationId, clusterId, policies)
}

func (u *PolicyUsecase) UpdatePoliciesForClusterID(ctx context.Context, organizationId string, clusterId domain.ClusterId, policyIds []uuid.UUID) (err error) {
	organization, err := u.organizationRepo.Get(ctx, organizationId)

	primaryClusterId := organization.PrimaryClusterId

	if err != nil {
		log.Errorf(ctx, "error is :%s(%T)", err.Error(), err)

		return httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "C_INVALID_ORGANIZATION_ID", "")
	}

	tkpolicies, err := policytemplate.GetTksPolicyCRs(ctx, primaryClusterId)
	if err != nil {
		log.Errorf(ctx, "error is :%s(%T)", err.Error(), err)

		return err
	}

	ids := make([]string, len(policyIds))
	for i, policyId := range policyIds {
		ids[i] = policyId.String()
	}

	// mapset.NewSet([]string{})

	for _, tkspolicy := range tkpolicies {
		policyId := tkspolicy.GetPolicyID()

		// 세팅 대상
		if slices.Contains(ids, policyId) {
			// 현재 클러스터가 안 추가되어 있으면
			if !slices.Contains(tkspolicy.Spec.Clusters, string(clusterId)) {
				tkspolicy.Spec.Clusters = append(tkspolicy.Spec.Clusters, string(clusterId))

				err := policytemplate.ApplyTksPolicyCR(ctx, primaryClusterId, &tkspolicy)
				if err != nil {
					log.Errorf(ctx, "error is :%s(%T)", err.Error(), err)
				}
			}

		} else { // 클리어 대상
			// 현재 클러스터가 추가되어 있으면
			if slices.Contains(tkspolicy.Spec.Clusters, string(clusterId)) {
				newClusters := []string{}

				tkspolicy.Spec.Clusters = slices.Filter(newClusters, tkspolicy.Spec.Clusters,
					func(s string) bool { return s != string(clusterId) })

				err := policytemplate.ApplyTksPolicyCR(ctx, primaryClusterId, &tkspolicy)
				if err != nil {
					log.Errorf(ctx, "error is :%s(%T)", err.Error(), err)
				}
			}
		}
	}

	policies := []model.Policy{}

	for _, policyId := range policyIds {
		policy, err := u.repo.GetByID(ctx, organizationId, policyId)

		if err != nil {
			log.Errorf(ctx, "error is :%s(%T)", err.Error(), err)

			continue
		}

		policies = append(policies, *policy)
	}

	return u.repo.UpdatePoliciesForClusterID(ctx, organizationId, clusterId, policies)
}

func (u *PolicyUsecase) DeletePoliciesForClusterID(ctx context.Context, organizationId string, clusterId domain.ClusterId, policyIds []uuid.UUID) (err error) {
	organization, err := u.organizationRepo.Get(ctx, organizationId)

	primaryClusterId := organization.PrimaryClusterId

	if err != nil {
		log.Errorf(ctx, "error is :%s(%T)", err.Error(), err)

		return httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "C_INVALID_ORGANIZATION_ID", "")
	}

	tkpolicies, err := policytemplate.GetTksPolicyCRs(ctx, primaryClusterId)
	if err != nil {
		log.Errorf(ctx, "error is :%s(%T)", err.Error(), err)

		return err
	}

	ids := make([]string, len(policyIds))
	for i, policyId := range policyIds {
		ids[i] = policyId.String()
	}

	for _, tkspolicy := range tkpolicies {
		policyId := tkspolicy.GetPolicyID()

		if slices.Contains(ids, policyId) {
			// 현재 클러스터가 추가되어 있으면 제거
			if slices.Contains(tkspolicy.Spec.Clusters, string(clusterId)) {
				newClusters := []string{}

				tkspolicy.Spec.Clusters = slices.Filter(newClusters, tkspolicy.Spec.Clusters,
					func(s string) bool { return s != string(clusterId) })
				err := policytemplate.ApplyTksPolicyCR(ctx, primaryClusterId, &tkspolicy)

				if err != nil {
					log.Errorf(ctx, "error is :%s(%T)", err.Error(), err)
				}
			}
		}
	}

	return u.repo.DeletePoliciesForClusterID(ctx, organizationId, clusterId, policyIds)
}

func (u *PolicyUsecase) GetStackPolicyStatistics(ctx context.Context, organizationId string, clusterId domain.ClusterId) (statistics *domain.StackPolicyStatistics, err error) {
	organization, err := u.organizationRepo.Get(ctx, organizationId)

	if err != nil {
		log.Errorf(ctx, "error is :%s(%T)", err.Error(), err)

		return nil, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "C_INVALID_ORGANIZATION_ID", "")
	}

	primaryClusterId := organization.PrimaryClusterId

	templateList, err := policytemplate.GetTksPolicyTemplateCRs(ctx, primaryClusterId)
	if err != nil {
		log.Errorf(ctx, "error is :%s(%T)", err.Error(), err)

		return nil, httpErrors.NewInternalServerError(fmt.Errorf("fail to get template list from kubernetes"), "P_FAILED_TO_CALL_KUBERNETES", "")
	}

	totalTemplateCount := len(templateList)
	outdatedTemplateIds := []string{}

	for _, template := range templateList {
		templateId := template.GetId()

		id, err := uuid.Parse(templateId)

		if err != nil {
			log.Errorf(ctx, "error is :%s(%T)", err.Error(), err)
			continue
		}

		version, err := u.templateRepo.GetLatestTemplateVersion(ctx, id)
		if err != nil {
			log.Errorf(ctx, "error is :%s(%T)", err.Error(), err)
			continue
		}

		if version != template.Spec.Version {
			outdatedTemplateIds = append(outdatedTemplateIds, templateId)
		}
	}

	outdatedTemplateCount := len(outdatedTemplateIds)
	uptodateTemplateCount := totalTemplateCount - outdatedTemplateCount

	policyList, err := policytemplate.GetTksPolicyCRs(ctx, primaryClusterId)
	if err != nil {
		log.Errorf(ctx, "error is :%s(%T)", err.Error(), err)

		return nil, httpErrors.NewInternalServerError(fmt.Errorf("fail to get policy list from kubernetes"), "P_FAILED_TO_CALL_KUBERNETES", "")
	}

	outdatedPolicyCount := 0

	for _, policy := range policyList {
		templateId := policy.GetTemplateID()

		if slices.Contains(outdatedTemplateIds, templateId) {
			outdatedPolicyCount++
		}
	}

	tototalPolicyCount := len(policyList)

	uptodatePolicyCount := tototalPolicyCount - outdatedPolicyCount

	result := domain.StackPolicyStatistics{
		TotalTemplateCount:     totalTemplateCount,
		UptodateTemplateCount:  uptodateTemplateCount,
		OutofdateTemplateCount: outdatedTemplateCount,
		TotalPolicyCount:       tototalPolicyCount,
		UptodatePolicyCount:    uptodatePolicyCount,
		OutofdatePolicyCount:   outdatedPolicyCount,
	}

	return &result, nil
}

func (u *PolicyUsecase) GetPolicyIDsByClusterID(ctx context.Context, clusterId domain.ClusterId) (out *[]uuid.UUID, err error) {
	return u.repo.GetPolicyIDsByClusterID(ctx, clusterId)
}

func extractNewTemplateParameter(paramdefs []*domain.ParameterDef, newParamDefs []*domain.ParameterDef) (policyParameters []domain.UpdatedPolicyTemplateParameter, err error) {
	diffParamDef, err := policytemplate.GetNewParamDefs(paramdefs, newParamDefs)

	if err != nil {
		return nil, err
	}

	results := []domain.UpdatedPolicyTemplateParameter{}

	// 새버전에 추가된 파라미터
	for _, paramdef := range diffParamDef {
		result := domain.UpdatedPolicyTemplateParameter{
			Name:  paramdef.Key,
			Type:  paramdef.Type,
			Value: paramdef.DefaultValue,
		}

		results = append(results, result)
	}

	return results, nil
}

func parseParameter(paramdefs []*domain.ParameterDef, newParamDefs []*domain.ParameterDef, parameter string) (policyParameters []domain.PolicyParameter, err error) {
	diffParamDef, err := policytemplate.GetNewParamDefs(paramdefs, newParamDefs)

	if err != nil {
		return nil, err
	}

	paramMap, err := parameterToValueMap(parameter)
	if err != nil {
		return nil, err
	}

	// 기존 파라미터
	results := []domain.PolicyParameter{}
	for _, paramdef := range paramdefs {
		result := domain.PolicyParameter{
			Name:      paramdef.Key,
			Type:      paramdef.Type,
			Value:     paramdef.DefaultValue,
			Updatable: false,
		}

		if val, ok := paramMap[paramdef.Key]; ok {
			result.Value = val
		}

		results = append(results, result)
	}

	// 새버전에 추가된 파라미터
	for _, paramdef := range diffParamDef {
		result := domain.PolicyParameter{
			Name:      paramdef.Key,
			Type:      paramdef.Type,
			Value:     paramdef.DefaultValue,
			Updatable: true,
		}

		results = append(results, result)
	}

	return results, nil
}

func parameterToValueMap(parameters string) (parameterMap map[string]string, err error) {
	var obj map[string]interface{}
	result := map[string]string{}

	err = json.Unmarshal([]byte(parameters), &obj)

	for k, v := range obj {
		value, err := json.Marshal(v)

		if err != nil {
			return nil, err
		}

		result[k] = string(value)
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}
