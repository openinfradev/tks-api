package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	policytemplate "github.com/openinfradev/tks-api/internal/policy-template"
	"github.com/openinfradev/tks-api/internal/serializer"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/util/validation"
)

type PolicyHandler struct {
	usecase usecase.IPolicyUsecase
}

type IPolicyHandler interface {
	CreatePolicy(w http.ResponseWriter, r *http.Request)
	UpdatePolicy(w http.ResponseWriter, r *http.Request)
	DeletePolicy(w http.ResponseWriter, r *http.Request)
	GetPolicy(w http.ResponseWriter, r *http.Request)
	ListPolicy(w http.ResponseWriter, r *http.Request)
	UpdatePolicyTargetClusters(w http.ResponseWriter, r *http.Request)
	GetMandatoryPolicies(w http.ResponseWriter, r *http.Request)
	SetMandatoryPolicies(w http.ResponseWriter, r *http.Request)
	ExistsPolicyName(w http.ResponseWriter, r *http.Request)
	ExistsPolicyResourceName(w http.ResponseWriter, r *http.Request)
	ListStackPolicyStatus(w http.ResponseWriter, r *http.Request)
	GetStackPolicyTemplateStatus(w http.ResponseWriter, r *http.Request)
	UpdateStackPolicyTemplateStatus(w http.ResponseWriter, r *http.Request)
	GetPolicyEdit(w http.ResponseWriter, r *http.Request)
	GetPolicyStatistics(w http.ResponseWriter, r *http.Request)
	AddPoliciesForStack(w http.ResponseWriter, r *http.Request)
	UpdatePoliciesForStack(w http.ResponseWriter, r *http.Request)
	DeletePoliciesForStack(w http.ResponseWriter, r *http.Request)
	StackPolicyStatistics(w http.ResponseWriter, r *http.Request)
}

func NewPolicyHandler(u usecase.Usecase) IPolicyHandler {
	return &PolicyHandler{
		usecase: u.Policy,
	}
}

// CreatePolicy godoc
//
//	@Tags			Policy
//	@Summary		[CreatePolicy] 정책 생성
//	@Description	새로운 정책을 생성한다. targetClusterIds가 명시되지 않으면 정책은 활성화되지 않은 상태로 생성된다. 다른 클러스터에 동일한 정책이 존재한다면 정책 생성이 아닌 정책 업데이트를 통해 targetClusterIds를 수정해야 한다.
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string						true	"조직 식별자(o로 시작)"
//	@Param			body			body		domain.CreatePolicyRequest	true	"create  policy request"
//	@Success		200				{object}	domain.CreatePolicyResponse
//	@Router			/organizations/{organizationId}/policies [post]
//	@Security		JWT
func (h *PolicyHandler) CreatePolicy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	input := domain.CreatePolicyRequest{}

	err := UnmarshalRequestInput(r, &input)

	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	// if input.Match != nil && input.MatchYaml != nil {
	// 	ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("both match and match yaml specified"), "P_INVALID_MATCH", ""))
	// 	return
	// }

	// if input.MatchYaml != nil {
	// 	var match domain.Match

	// 	err := yaml.Unmarshal([]byte(*input.MatchYaml), &match)

	// 	if err != nil {
	// 		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("match yaml error: %s", err), "P_INVALID_MATCH", ""))
	// 		return
	// 	}

	// 	if err := ValidateDomainObject(match); err != nil {
	// 		ErrorJSON(w, r, err)
	// 		return
	// 	}
	// } else if input.Match != nil {
	// 	normaized, err := policytemplate.CheckAndNormalizeKinds(input.Match.Kinds)

	// 	if err != nil {
	// 		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("match error: %s", err), "P_INVALID_MATCH", ""))
	// 		return
	// 	}

	// 	input.Match.Kinds = normaized
	// }

	var dto model.Policy
	if err = serializer.Map(r.Context(), input, &dto); err != nil {
		log.Info(r.Context(), err)
	}

	if input.Target != nil {
		match, matchYaml, err := ValidateAndGetMatch(input.Target)
		if err != nil {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "P_INVALID_MATCH", ""))
			return
		}
		dto.Match = match
		dto.MatchYaml = matchYaml
	}

	if len(input.PolicyResourceName) > 0 {
		errMsgs := validation.IsDNS1123Subdomain(input.PolicyResourceName)

		if len(errMsgs) > 0 {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid k8s resource name for policy: %v", errMsgs), "P_INVALID_RESURCE_NAME", ""))
			return
		}
	}

	policyId, err := h.usecase.Create(r.Context(), organizationId, dto)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.CreatePolicyResponse
	out.ID = policyId.String()

	ResponseJSON(w, r, http.StatusOK, out)
}

// UpdatePolicy godoc
//
//	@Tags			Policy
//	@Summary		[UpdatePolicy] 정책을 업데이트
//	@Description	정책의 내용을 업데이트 한다. 업데이트할 필드만 명시하면 된다.
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string						true	"조직 식별자(o로 시작)"
//	@Param			policyId		path		string						true	"정책 식별자(uuid)"
//	@Param			body			body		domain.UpdatePolicyRequest	true	"update policy set request"
//	@Success		200				{object}	nil
//	@Router			/organizations/{organizationId}/policies/{policyId} [patch]
//	@Security		JWT
func (h *PolicyHandler) UpdatePolicy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	policyId, ok := vars["policyId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid policyId"), "C_INVALID_POLICY_ID", ""))
		return
	}

	id, err := uuid.Parse(policyId)
	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid policyId"), "C_INVALID_POLICY_ID", ""))
		return
	}

	input := domain.UpdatePolicyRequest{}

	err = UnmarshalRequestInput(r, &input)

	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, r, err)
		return
	}

	// if input.Match != nil && input.MatchYaml != nil {
	// 	ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("both match and match yaml specified"), "P_INVALID_MATCH", ""))
	// 	return
	// }

	// if input.MatchYaml != nil {
	// 	var match domain.Match

	// 	err := yaml.Unmarshal([]byte(*input.MatchYaml), &match)

	// 	if err != nil {
	// 		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("match yaml error: %s", err), "P_INVALID_MATCH", ""))
	// 		return
	// 	}

	// 	if err := ValidateDomainObject(match); err != nil {
	// 		ErrorJSON(w, r, err)
	// 		return
	// 	}
	// } else if input.Match != nil {
	// 	normaized, err := policytemplate.CheckAndNormalizeKinds(input.Match.Kinds)
	// 	if err != nil {
	// 		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("match error: %s", err), "P_INVALID_MATCH", ""))
	// 		return
	// 	}

	// 	input.Match.Kinds = normaized
	// }

	var match *domain.Match
	var matchYaml *string
	if input.Target != nil {
		match, matchYaml, err = ValidateAndGetMatch(input.Target)
		if err != nil {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "P_INVALID_MATCH", ""))
			return
		}

	}

	var templateId *uuid.UUID = nil

	if input.TemplateId != nil {
		tuuid, err := uuid.Parse(*input.TemplateId)
		if err != nil {
			log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
			ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid policyTemplateId"), "C_INVALID_POLICY_TEMPLATE_ID", ""))
			return
		}
		templateId = &tuuid
	}

	err = h.usecase.Update(r.Context(), organizationId, id,
		input.Mandatory, input.PolicyName, input.Description, templateId, input.EnforcementAction,
		input.Parameters, match, matchYaml, input.TargetClusterIds)

	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
			return
		}

		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, nil)
}

// DeletePolicy godoc
//
//	@Tags			Policy
//	@Summary		[DeletePolicy] 정책 삭제
//	@Description	정첵을 삭제한다. 정책이 적용된 클러스터가 있으면 삭제되지 않으므로 삭제 전 적용된 클러스터가 비어있어야 한다.
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"조직 식별자(o로 시작)"
//	@Param			policyId		path		string	true	"정책 식별자(uuid)"
//	@Success		200				{object}	nil
//	@Router			/organizations/{organizationId}/policies/{policyId} [delete]
//	@Security		JWT
func (h *PolicyHandler) DeletePolicy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	policyId, ok := vars["policyId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid policyId"), "C_INVALID_POLICY_ID", ""))
		return
	}

	id, err := uuid.Parse(policyId)
	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid policyId"), "C_INVALID_POLICY_ID", ""))
		return
	}

	err = h.usecase.Delete(r.Context(), organizationId, id)

	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
			return
		}

		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, "")
}

// GetPolicy godoc
//
//	@Tags			Policy
//	@Summary		[GetPolicy] 정책 조회
//	@Description	정책 정보를 조회한다.
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"조직 식별자(o로 시작)"
//	@Param			policyId		path		string	true	"정책 식별자(uuid)"
//	@Success		200				{object}	domain.GetPolicyResponse
//	@Router			/organizations/{organizationId}/policies/{policyId} [get]
//	@Security		JWT
func (h *PolicyHandler) GetPolicy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	policyId, ok := vars["policyId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid policyId"), "C_INVALID_POLICY_ID", ""))
		return
	}

	id, err := uuid.Parse(policyId)
	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "C_INVALID_POLICY_ID", ""))
			return
		}

		ErrorJSON(w, r, err)
		return
	}

	policy, err := h.usecase.Get(r.Context(), organizationId, id)
	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "P_NOT_FOUND_POLICY", ""))
			return
		}

		ErrorJSON(w, r, err)
		return
	}

	if policy == nil {
		ResponseJSON(w, r, http.StatusNotFound, nil)
		return
	}

	var out domain.GetPolicyResponse
	if err = serializer.Map(r.Context(), *policy, &out.Policy); err != nil {
		log.Error(r.Context(), err)
	}

	out.Policy.TargetClusters = make([]domain.SimpleClusterResponse, len(policy.TargetClusters))

	for i, targetCluster := range policy.TargetClusters {
		if err = serializer.Map(r.Context(), targetCluster, &out.Policy.TargetClusters[i]); err != nil {
			log.Error(r.Context(), err)
		}
	}

	out.Policy.Target = MatchAndMatchYamlToTarget(policy.Match, policy.MatchYaml)

	ResponseJSON(w, r, http.StatusOK, out)
}

// ListPolicy godoc
//
//	@Tags			Policy
//	@Summary		[ListPolicy] 정책 목록 조회
//	@Description	정책 목록을 조회한다.
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string		true	"조직 식별자(o로 시작)"
//	@Param			filledParameter	query		string		false	"filledParameter"
//	@Param			pageSize		query		string		false	"pageSize"
//	@Param			pageNumber		query		string		false	"pageNumber"
//	@Param			sortColumn		query		string		false	"sortColumn"
//	@Param			sortOrder		query		string		false	"sortOrder"
//	@Param			filters			query		[]string	false	"filters"
//	@Success		200				{object}	domain.ListPolicyResponse
//	@Router			/organizations/{organizationId}/policies [get]
//	@Security		JWT
func (h *PolicyHandler) ListPolicy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	urlParams := r.URL.Query()

	pg := pagination.NewPagination(&urlParams)

	filledParameter := false

	parse := urlParams.Get("filledParameter")
	if len(parse) > 0 {
		parsedBool, err := strconv.ParseBool(parse)
		if err != nil {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid fillParameter: '%s'", parse), "PT_INVALID_FILLPARAMETER", ""))
			return
		}
		filledParameter = parsedBool
	}

	policies, err := h.usecase.Fetch(r.Context(), organizationId, pg, filledParameter)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.ListPolicyResponse
	out.Policies = make([]domain.PolicyResponse, len(*policies))
	for i, policy := range *policies {
		if err := serializer.Map(r.Context(), policy, &out.Policies[i]); err != nil {
			log.Info(r.Context(), err)
			continue
		}

		out.Policies[i].TargetClusters = make([]domain.SimpleClusterResponse, len(policy.TargetClusters))

		for j, targetCluster := range policy.TargetClusters {
			if err = serializer.Map(r.Context(), targetCluster, &out.Policies[i].TargetClusters[j]); err != nil {
				log.Error(r.Context(), err)
			}
		}

		out.Policies[i].Target = MatchAndMatchYamlToTarget(policy.Match, policy.MatchYaml)

		if filledParameter {
			parameterSchema := policy.PolicyTemplate.ParametersSchema
			parameters := policy.Parameters

			err = policytemplate.FillParamDefFromJsonStr(parameterSchema, parameters)
			if err != nil {
				log.Error(r.Context(), err)
			} else {
				out.Policies[i].FilledParameters = parameterSchema
			}
		}
	}

	if out.Pagination, err = pg.Response(r.Context()); err != nil {
		log.Info(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// UpdatePolicyTargetClusters godoc
//
//	@Tags			Policy
//	@Summary		[UpdatePolicyTargetClusters] 정책 적용 대상 클러스터 수정
//	@Description	정책 적용 대상 클러스터를 수정한다. 추가할 클러스터 목록과 제거할 클러스터 목록 중 하나만 명시되어야 한다. 현재 정책이 배포된 클러스터를 확인하지 않고도 특정 클러스터를 추가하거나 제거할 수 있는 편의 API이다.
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string								true	"조직 식별자(o로 시작)"
//	@Param			policyId		path		string								true	"정책 식별자(uuid)"
//	@Param			body			body		domain.UpdatePolicyClustersRequest	true	"update policy set request"
//	@Success		200				{object}	nil
//	@Router			/organizations/{organizationId}/policies/{policyId}/clusters [patch]
//	@Security		JWT
func (h *PolicyHandler) UpdatePolicyTargetClusters(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	policyId, ok := vars["policyId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid policyId"), "C_INVALID_POLICY_ID", ""))
		return
	}

	id, err := uuid.Parse(policyId)
	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "C_INVALID_POLICY_ID", ""))
			return
		}

		ErrorJSON(w, r, err)
		return
	}

	input := domain.UpdatePolicyClustersRequest{}

	err = UnmarshalRequestInput(r, &input)

	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, r, err)
		return
	}

	err = h.usecase.UpdatePolicyTargetClusters(r.Context(), organizationId, id,
		input.CurrentTargetClusterIds, input.NewTargetClusterIds)

	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
			return
		}

		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, nil)
}

// GetMandatoryPolicies godoc
//
//	@Tags			Policy
//	@Summary		[GetMandatoryPolicies] 필수 정책 템플릿, 정책을 조회
//	@Description	템플릿, 정책이 필수 인지 여부를 조회한다.
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"조직 식별자(o로 시작)"
//	@Success		200				{object}	domain.GetMandatoryPoliciesResponse
//	@Router			/organizations/{organizationId}/mandatory-policies [get]
//	@Security		JWT
func (h *PolicyHandler) GetMandatoryPolicies(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	out, err := h.usecase.GetMandatoryPolicies(r.Context(), organizationId)

	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
			return
		}

		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// SetMandatoryPolicies godoc
//
//	@Tags			Policy
//	@Summary		[SetMandatoryPolicies] 필수 정책 템플릿, 정책을 설정
//	@Description	템플릿, 정책이 필수 인지 여부를 설정한다.
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string								true	"조직 식별자(o로 시작)"
//	@Param			body			body		domain.SetMandatoryPoliciesRequest	true	"update mandatory policy/policy template request"
//	@Success		200				{object}	nil
//	@Router			/organizations/{organizationId}/mandatory-policies [patch]
//	@Security		JWT
func (h *PolicyHandler) SetMandatoryPolicies(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	input := domain.SetMandatoryPoliciesRequest{}

	err := UnmarshalRequestInput(r, &input)

	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	mandatoryPolicyIds := []uuid.UUID{}
	nonMandatoryPolicyIds := []uuid.UUID{}
	for _, policy := range input.Policies {
		policyId, err := uuid.Parse(policy.PolicyId)
		if err != nil {
			log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
			if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
				ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "C_INVALID_POLICY_ID", ""))
				return
			}

			ErrorJSON(w, r, err)
			return
		}

		if policy.Mandatory {
			mandatoryPolicyIds = append(mandatoryPolicyIds, policyId)
		} else {
			nonMandatoryPolicyIds = append(nonMandatoryPolicyIds, policyId)
		}
	}

	err = h.usecase.SetMandatoryPolicies(r.Context(), organizationId, mandatoryPolicyIds, nonMandatoryPolicyIds)

	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
			return
		}

		ErrorJSON(w, r, err)
		return
	}
}

// ExistsPolicyName godoc
//
//	@Tags			Policy
//	@Summary		[ExistsPolicyName] 정책 이름 존재 여부 확인
//	@Description	해당 이름을 가진 정책이 이미 존재하는지 확인한다.
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"조직 식별자(o로 시작)"
//	@Param			policyName		path		string	true	"정책 이름"
//	@Success		200				{object}	domain.CheckExistedResponse
//	@Router			/organizations/{organizationId}/policies/name/{policyName}/existence [get]
//	@Security		JWT
func (h *PolicyHandler) ExistsPolicyName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	policyName, ok := vars["policyName"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("policyName not found in path"),
			"P_INVALID_POLICY_NAME", ""))
		return
	}

	exist, err := h.usecase.IsPolicyNameExist(r.Context(), organizationId, policyName)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.CheckExistedResponse
	out.Existed = exist

	ResponseJSON(w, r, http.StatusOK, out)
}

// ExistsPolicyResourceName godoc
//
//	@Tags			Policy
//	@Summary		[ExistsPolicyResourceName] 정책 자원 이름 존재 여부 확인
//	@Description	해당 자원 이름을 가진 정책이 이미 존재하는지 확인한다.
//	@Accept			json
//	@Produce		json
//	@Param			organizationId		path		string	true	"조직 식별자(o로 시작)"
//	@Param			policyResourceName	path		string	true	"정책 자원 이름(쿠버네티스 배포 시 자원 이름)"
//	@Success		200					{object}	domain.CheckExistedResponse
//	@Router			/organizations/{organizationId}/policies/resource-name/{policyResourceName}/existence [get]
//	@Security		JWT
func (h *PolicyHandler) ExistsPolicyResourceName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	policyResourceName, ok := vars["policyResourceName"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("policyResourceName not found in path"),
			"P_INVALID_POLICY_RESOURCE_NAME", ""))
		return
	}

	exist, err := h.usecase.IsPolicyResourceNameExist(r.Context(), organizationId, policyResourceName)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.CheckExistedResponse
	out.Existed = exist

	ResponseJSON(w, r, http.StatusOK, out)
}

// ListStackPolicyStatus godoc
//
//	@Tags			StackPolicyStatus
//	@Summary		[ListStackPolicyStatus] 클러스터의 정책과 정책 템플릿, 버전 조회
//	@Description	클러스터의 정책과 정책 템플릿, 버전 등을 포함한 상태 목록을 조회한다.
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string		true	"조직 식별자(o로 시작)"
//	@Param			stackId			path		string		true	"스택 식별자"
//	@Param			pageSize		query		string		false	"pageSize"
//	@Param			pageNumber		query		string		false	"pageNumber"
//	@Param			sortColumn		query		string		false	"sortColumn"
//	@Param			sortOrder		query		string		false	"sortOrder"
//	@Param			filters			query		[]string	false	"filters"
//	@Success		200				{object}	domain.ListStackPolicyStatusResponse
//	@Router			/organizations/{organizationId}/stacks/{stackId}/policy-status [get]
//	@Security		JWT
func (h *PolicyHandler) ListStackPolicyStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	stackId, ok := vars["stackId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid clusterId"),
			"C_INVALID_STACK_ID", ""))
		return
	}

	urlParams := r.URL.Query()

	pg := pagination.NewPagination(&urlParams)

	policyStatuses, err := h.usecase.ListStackPolicyStatus(r.Context(), stackId, pg)

	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	out := domain.ListStackPolicyStatusResponse{
		Polices: policyStatuses,
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// StackPolicyStatistics godoc
//
//	@Tags			StackPolicyStatus
//	@Summary		[ListStackPolicyStatus] 클러스터의 정책과 정책 템플릿, 버전 조회
//	@Description	클러스터의 정책과 정책 템플릿, 버전 등을 포함한 상태 목록을 조회한다.
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string		true	"조직 식별자(o로 시작)"
//	@Param			stackId			path		string		true	"스택 식별자"
//	@Param			pageSize		query		string		false	"pageSize"
//	@Param			pageNumber		query		string		false	"pageNumber"
//	@Param			sortColumn		query		string		false	"sortColumn"
//	@Param			sortOrder		query		string		false	"sortOrder"
//	@Param			filters			query		[]string	false	"filters"
//	@Success		200				{object}	domain.StackPolicyStatistics
//	@Router			/organizations/{organizationId}/stacks/{stackId}/statistics [get]
//	@Security		JWT
func (h *PolicyHandler) StackPolicyStatistics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid clusterId"),
			"C_INVALID_STACK_ID", ""))
		return
	}

	stackId, ok := vars["stackId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid clusterId"),
			"C_INVALID_STACK_ID", ""))
		return
	}

	stackPolicyStatistics, err := h.usecase.GetStackPolicyStatistics(r.Context(),
		organizationId, domain.ClusterId(stackId))

	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, stackPolicyStatistics)
}

// GetStackPolicyTemplateStatus godoc
//
//	@Tags			StackPolicyStatus
//	@Summary		[GetStackPolicyTemplateStatus] 클러스터 템플릿 상태 상세 조회
//	@Description	템플릿의 클러스터 버전 등 상태를 조회한다.
//	@Accept			json
//	@Produce		json
//	@Param			organizationId		path		string	true	"조직 식별자(o로 시작)"
//	@Param			stackId				path		string	true	"스택 식별자"
//	@Param			policyTemplateId	path		string	true	"정책 템플릿 식별자(uuid)"
//	@Success		200					{object}	domain.GetStackPolicyTemplateStatusResponse
//	@Router			/organizations/{organizationId}/stacks/{stackId}/policy-templates/{policyTemplateId} [get]
//	@Security		JWT
func (h *PolicyHandler) GetStackPolicyTemplateStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	stackId, ok := vars["stackId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid clusterId"),
			"C_INVALID_STACK_ID", ""))
		return
	}

	policyTemplateId, ok := vars["policyTemplateId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid policyTemplateId"), "C_INVALID_POLICY_TEMPLATE_ID", ""))
		return
	}

	id, err := uuid.Parse(policyTemplateId)
	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid policyId"), "C_INVALID_POLICY_ID", ""))
		return
	}

	out, err := h.usecase.GetStackPolicyTemplateStatus(r.Context(), stackId, id)

	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// UpdateStackPolicyTemplateStatus godoc
//
//	@Tags			StackPolicyStatus
//	@Summary		[UpdateStackPolicyTemplateStatus] 템플릿 버전 업데이트
//	@Description	해당 템플릿의 버전 업데이트 및 연관된 정책의 새 기본값을 설정한다.
//	@Accept			json
//	@Produce		json
//	@Param			organizationId		path		string											true	"조직 식별자(o로 시작)"
//	@Param			stackId				path		string											true	"스택 식별자"
//	@Param			policyTemplateId	path		string											true	"정책 템플릿 식별자(uuid)"
//	@Param			body				body		domain.UpdateStackPolicyTemplateStatusRequest	true	"update stack policy template status request"
//	@Success		200					{object}	nil
//	@Router			/organizations/{organizationId}/stacks/{stackId}/policy-templates/{policyTemplateId} [patch]
//	@Security		JWT
func (h *PolicyHandler) UpdateStackPolicyTemplateStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	stackId, ok := vars["stackId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid clusterId"),
			"C_INVALID_STACK_ID", ""))
		return
	}

	policyTemplateId, ok := vars["policyTemplateId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid policyTemplateId"), "C_INVALID_POLICY_TEMPLATE_ID", ""))
		return
	}

	input := domain.UpdateStackPolicyTemplateStatusRequest{}

	err := UnmarshalRequestInput(r, &input)

	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	id, err := uuid.Parse(policyTemplateId)
	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid policyId"), "C_INVALID_POLICY_ID", ""))
		return
	}

	err = h.usecase.UpdateStackPolicyTemplateStatus(r.Context(), stackId, id,
		input.TemplateCurrentVersion, input.TemplateTargetVerson)

	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
			return
		}

		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, nil)
}

// GetPolicyEdit godoc
//
//	@Tags			Policy
//	@Summary		[GetPolicy] 정책 조회
//	@Description	정책 정보를 조회한다.
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"조직 식별자(o로 시작)"
//	@Param			policyId		path		string	true	"정책 식별자(uuid)"
//	@Success		200				{object}	domain.GetPolicyResponse
//	@Router			/organizations/{organizationId}/policies/{policyId}/edit [get]
//	@Security		JWT
func (h *PolicyHandler) GetPolicyEdit(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	policyId, ok := vars["policyId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid policyId"), "C_INVALID_POLICY_ID", ""))
		return
	}

	id, err := uuid.Parse(policyId)
	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "C_INVALID_POLICY_ID", ""))
			return
		}

		ErrorJSON(w, r, err)
		return
	}

	policy, err := h.usecase.GetForEdit(r.Context(), organizationId, id)
	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "P_NOT_FOUND_POLICY", ""))
			return
		}

		ErrorJSON(w, r, err)
		return
	}

	if policy == nil {
		ResponseJSON(w, r, http.StatusNotFound, nil)
		return
	}

	var out domain.GetPolicyResponse
	if err = serializer.Map(r.Context(), *policy, &out.Policy); err != nil {
		log.Error(r.Context(), err)
	}

	out.Policy.TargetClusters = make([]domain.SimpleClusterResponse, len(policy.TargetClusters))

	for i, targetCluster := range policy.TargetClusters {
		if err = serializer.Map(r.Context(), targetCluster, &out.Policy.TargetClusters[i]); err != nil {
			log.Error(r.Context(), err)
		}
	}

	out.Policy.Target = MatchAndMatchYamlToTarget(policy.Match, policy.MatchYaml)

	parameterSchema := policy.PolicyTemplate.ParametersSchema
	parameters := policy.Parameters

	err = policytemplate.FillParamDefFromJsonStr(parameterSchema, parameters)
	if err != nil {
		log.Error(r.Context(), err)
	} else {
		out.Policy.FilledParameters = parameterSchema
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetPolicyStatistics godoc
//
//	@Tags			Policy
//	@Summary		[GetPolicyStatistics] 정책 템플릿, 정책 통계 조회
//	@Description	템플릿, 정책의 통계를 조회한다.
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"조직 식별자(o로 시작)"
//	@Success		200				{object}	domain.PolicyStatisticsResponse
//	@Router			/organizations/{organizationId}/policy-statistics [get]
//	@Security		JWT
func (h *PolicyHandler) GetPolicyStatistics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid organizationId"),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	out, err := h.usecase.GetPolicyStatistics(r.Context(), organizationId)

	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
			return
		}

		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// AddPoliciesForStack godoc
//
//	@Tags			Policy
//	@Summary		[AddPoliciesForStack] 특정 스택의 정책 목록 추가
//	@Description	특정 스택의 정책 목록을 정책 식별자 리스트로 지정해서 추가한다.
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string								true	"조직 식별자(o로 시작)"
//	@Param			stackId			path		string								true	"스택 식별자"
//	@Param			body			body		domain.AddPoliciesForStackRequest	true	"add policies for stack request"
//	@Success		200				{object}	nil
//	@Router			/organizations/{organizationId}/stacks/{stackId}/policies [post]
//	@Security		JWT
func (h *PolicyHandler) AddPoliciesForStack(w http.ResponseWriter, r *http.Request) {
	// TODO: API 형상 추후 반드시 검토

	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("organizationId not found in path"),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	stackId, ok := vars["stackId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("stackId not found in path"),
			"C_INVALID_STACK_ID", ""))
		return
	}

	input := domain.AddPoliciesForStackRequest{}

	err := UnmarshalRequestInput(r, &input)

	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, r, err)
		return
	}

	ids := make([]uuid.UUID, len(input.PolicyIds))

	for i, policyId := range input.PolicyIds {
		id, err := uuid.Parse(policyId)
		if err != nil {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "C_INVALID_POLICY_ID", ""))
			return
		}
		ids[i] = id
	}

	err = h.usecase.AddPoliciesForClusterID(r.Context(), organizationId, domain.ClusterId(stackId), ids)

	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, nil)
}

func (h *PolicyHandler) UpdatePoliciesForStack(w http.ResponseWriter, r *http.Request) {
	// TODO: API 형상 추후 반드시 검토

	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("organizationId not found in path"),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	stackId, ok := vars["stackId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("stackId not found in path"),
			"C_INVALID_STACK_ID", ""))
		return
	}

	input := domain.UpdatePoliciesForStackRequest{}

	err := UnmarshalRequestInput(r, &input)

	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, r, err)
		return
	}

	ids := make([]uuid.UUID, len(input.PolicyIds))

	for i, policyId := range input.PolicyIds {
		id, err := uuid.Parse(policyId)
		if err != nil {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "C_INVALID_POLICY_ID", ""))
			return
		}
		ids[i] = id
	}

	err = h.usecase.UpdatePoliciesForClusterID(r.Context(), organizationId, domain.ClusterId(stackId), ids)

	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, nil)
}

// DeletePoliciesForStack godoc
//
//	@Tags			Policy
//	@Summary		[DeletePoliciesForStack] 특정 스택의 정책 제거
//	@Description	특정 스택에서 정책 식별자로 지정된 정책을 제거한다.
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path	string									true	"조직 식별자(o로 시작)"
//	@Param			stackId			path	string									true	"스택 식별자"
//	@Param			body			body	domain.DeletePoliciesForStackRequest	true	"delete policies for stack request"
//	@Router			/organizations/{organizationId}/stacks/{stackId}/policies [put]
//	@Security		JWT
func (h *PolicyHandler) DeletePoliciesForStack(w http.ResponseWriter, r *http.Request) {
	// TODO: API 형상 추후 반드시 검토

	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("organizationId not found in path"),
			"C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	stackId, ok := vars["stackId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("stackId not found in path"),
			"C_INVALID_STACK_ID", ""))
		return
	}

	input := domain.DeletePoliciesForStackRequest{}

	err := UnmarshalRequestInput(r, &input)

	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, r, err)
		return
	}

	ids := make([]uuid.UUID, len(input.PolicyIds))

	for i, policyId := range input.PolicyIds {
		id, err := uuid.Parse(policyId)
		if err != nil {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "C_INVALID_POLICY_ID", ""))
			return
		}
		ids[i] = id
	}

	err = h.usecase.DeletePoliciesForClusterID(r.Context(), organizationId, domain.ClusterId(stackId), ids)

	if err != nil {
		log.Errorf(r.Context(), "error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, nil)
}

func ValidateAndGetMatch(t *domain.Target) (match *domain.Match, matchYaml *string, err error) {
	if t.Type != "simple" && t.Type != "yaml" {
		return nil, nil, fmt.Errorf("invalid target type '%s'", t.Type)
	}

	if t.Type == "simple" {
		var simpleMatch domain.SimpleMatch

		jsonbytes := []byte(t.Value)
		dec := json.NewDecoder(bytes.NewReader(jsonbytes))
		dec.DisallowUnknownFields()
		err = dec.Decode(&simpleMatch)

		if err != nil {
			return nil, nil, err
		}

		kinds, err := policytemplate.CheckAndNormalizeKinds(simpleMatch.Kinds)

		if err != nil {
			return nil, nil, err
		}

		return &domain.Match{
			Namespaces:         simpleMatch.Namespaces,
			ExcludedNamespaces: simpleMatch.ExcludedNamespaces,
			Kinds:              kinds,
		}, nil, nil
	} else {
		var match domain.Match

		stripped := strings.ReplaceAll(t.Value, "\r", "")
		yamlbytes := []byte(stripped)

		dec := yaml.NewDecoder(bytes.NewReader(yamlbytes))
		dec.KnownFields(true)
		err = dec.Decode(&match)

		if err != nil {
			return nil, nil, err
		}

		return nil, &t.Value, nil
	}
}

func MatchAndMatchYamlToTarget(match *domain.Match, matchYaml *string) *domain.Target {
	if match != nil {
		flattenKinds := []string{}

		for _, kind := range match.Kinds {
			for _, kindkind := range kind.Kinds {
				if !slices.Contains(flattenKinds, kindkind) {
					flattenKinds = append(flattenKinds, kindkind)
				}
			}
		}

		simpleMatch := domain.SimpleMatch{
			Namespaces:         match.Namespaces,
			ExcludedNamespaces: match.ExcludedNamespaces,
			Kinds:              flattenKinds,
		}

		var value string
		bytes, err := json.Marshal(simpleMatch)

		if err == nil {
			value = string(bytes)
		} else {
			value = ""
		}

		return &domain.Target{
			Type:  "simple",
			Value: value,
		}
	}

	if matchYaml != nil {
		return &domain.Target{
			Type:  "yaml",
			Value: *matchYaml,
		}
	}

	return nil
}
