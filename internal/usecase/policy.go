package usecase

import (
	"context"
	"fmt"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
)

type IPolicyUsecase interface {
	Create(ctx context.Context, organizationId string, dto model.Policy) (policyId uuid.UUID, err error)
	Update(ctx context.Context, organizationId string, policyId uuid.UUID,
		mandatory *bool, policyName *string, description *string, templateId *uuid.UUID, enforcementAction *string,
		parameters *string, match *domain.Match, targetClusterIds *[]string) (err error)
	Delete(ctx context.Context, organizationId string, policyId uuid.UUID) (err error)
	Get(ctx context.Context, organizationId string, policyId uuid.UUID) (policy *model.Policy, err error)
	Fetch(ctx context.Context, organizationId string, pg *pagination.Pagination) (*[]model.Policy, error)
	IsPolicyIdExist(ctx context.Context, organizationId string, policyId uuid.UUID) (exists bool, err error)
	IsPolicyNameExist(ctx context.Context, organizationId string, policyName string) (exists bool, err error)
	UpdatePolicyTargetClusters(ctx context.Context, organizationId string, policyId uuid.UUID, currentClusterIds []string, targetClusterIds []string) (err error)
	SetMandatoryPolicies(ctx context.Context, organizationId string, mandatoryPolicyIds []uuid.UUID, nonMandatoryPolicyIds []uuid.UUID) (err error)
	GetMandatoryPolicies(ctx context.Context, organizationId string) (response *domain.GetMandatoryPoliciesResponse, err error)
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

func (u *PolicyUsecase) Create(ctx context.Context, organizationId string, dto model.Policy) (policyId uuid.UUID, err error) {
	dto.OrganizationId = organizationId

	user, ok := request.UserFrom(ctx)
	if !ok {
		return uuid.Nil, httpErrors.NewUnauthorizedError(fmt.Errorf("invalid token"), "A_INVALID_TOKEN", "")
	}

	exists, err := u.repo.ExistByName(ctx, dto.OrganizationId, dto.TemplateName)
	if err != nil {
		return uuid.Nil, err
	}

	if exists {
		return uuid.Nil, httpErrors.NewBadRequestError(httpErrors.DuplicateResource, "PT_CREATE_ALREADY_EXISTED_NAME", "policy template name already exists")
	}

	dto.TargetClusters = make([]model.Cluster, len(dto.TargetClusterIds))
	for i, clusterId := range dto.TargetClusterIds {

		cluster, err := u.clusterRepo.Get(ctx, domain.ClusterId(clusterId))
		if err != nil {
			return uuid.Nil, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"), "C_INVALID_ORGANIZATION_ID", "")
		}
		dto.TargetClusters[i] = cluster
	}

	userId := user.GetUserId()
	dto.CreatorId = &userId

	id, err := u.repo.Create(ctx, dto)

	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func (u *PolicyUsecase) Update(ctx context.Context, organizationId string, policyId uuid.UUID,
	mandatory *bool, policyName *string, description *string, templateId *uuid.UUID, enforcementAction *string,
	parameters *string, match *domain.Match, targetClusterIds *[]string) (err error) {

	user, ok := request.UserFrom(ctx)
	if !ok {
		return httpErrors.NewBadRequestError(fmt.Errorf("invalid token"), "A_INVALID_TOKEN", "")
	}

	_, err = u.repo.GetByID(ctx, organizationId, policyId)
	if err != nil {
		return httpErrors.NewNotFoundError(err, "P_FAILED_FETCH_POLICY_TEMPLATE", "")
	}

	updateMap := make(map[string]interface{})

	if mandatory != nil {
		updateMap["mandatory"] = mandatory
	}

	if policyName != nil {
		exists, err := u.repo.ExistByName(ctx, organizationId, *policyName)
		if err == nil && exists {
			return httpErrors.NewBadRequestError(httpErrors.DuplicateResource, "P_INVALID_POLICY__NAME", "policy template name already exists")
		}
		updateMap["policy_name"] = policyName
	}

	if description != nil {
		updateMap["description"] = description
	}

	if templateId != nil {
		updateMap["template_id"] = templateId
	}

	if enforcementAction != nil {
		updateMap["enforcement_action"] = enforcementAction
	}

	if parameters != nil {
		updateMap["parameters"] = parameters
	}

	if parameters != nil {
		updateMap["policy_match"] = match.JSON()
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

	return nil
}

func (u *PolicyUsecase) Delete(ctx context.Context, organizationId string, policyId uuid.UUID) (err error) {
	return u.repo.Delete(ctx, organizationId, policyId)
}

func (u *PolicyUsecase) Get(ctx context.Context, organizationId string, policyId uuid.UUID) (policy *model.Policy, err error) {
	return u.repo.GetByID(ctx, organizationId, policyId)
}

func (u *PolicyUsecase) Fetch(ctx context.Context, organizationId string, pg *pagination.Pagination) (*[]model.Policy, error) {
	return u.repo.Fetch(ctx, organizationId, pg)
}

func (u *PolicyUsecase) IsPolicyNameExist(ctx context.Context, organizationId string, policyName string) (exists bool, err error) {
	return u.repo.ExistByName(ctx, organizationId, policyName)
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

	return u.repo.UpdatePolicyTargetClusters(ctx, organizationId, policyId, currentClusterIds, targetClusters)
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

		if len(policyTemplate.PermittedOrganizationIds) == 0 ||
			mapset.NewSet(policyTemplate.PermittedOrganizationIds...).Contains(organizationId) {
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
