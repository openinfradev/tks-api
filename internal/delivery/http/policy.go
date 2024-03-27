package http

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/serializer"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
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

	var dto model.Policy
	if err = serializer.Map(r.Context(), input, &dto); err != nil {
		log.Info(r.Context(), err)
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
		input.Mandatory, input.PolicyName, &input.Description, templateId, input.EnforcementAction,
		input.Parameters, input.Match, input.TargetClusterIds)

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
//	@Param			limit			query		string		false	"pageSize"
//	@Param			page			query		string		false	"pageNumber"
//	@Param			soertColumn		query		string		false	"sortColumn"
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

	policies, err := h.usecase.Fetch(r.Context(), organizationId, pg)
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
//	@Summary		[ExistsPolicyName] 정책 아름 존재 여부 확인
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
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("policyTemplateName not found in path"),
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
