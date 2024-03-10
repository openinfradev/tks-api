package http

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/serializer"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"

	"github.com/Masterminds/semver/v3"
)

type PolicyTemplateHandler struct {
	usecase usecase.IPolicyTemplateUsecase
}

type IPolicyTemplateHandler interface {
	CreatePolicyTemplate(w http.ResponseWriter, r *http.Request)
	UpdatePolicyTemplate(w http.ResponseWriter, r *http.Request)
	DeletePolicyTemplate(w http.ResponseWriter, r *http.Request)
	GetPolicyTemplate(w http.ResponseWriter, r *http.Request)
	ListPolicyTemplate(w http.ResponseWriter, r *http.Request)
	ExistsPolicyTemplateName(w http.ResponseWriter, r *http.Request)
	ExistsPolicyTemplateKind(w http.ResponseWriter, r *http.Request)
	ListPolicyTemplateStatistics(w http.ResponseWriter, r *http.Request)
	GetPolicyTemplateDeploy(w http.ResponseWriter, r *http.Request)
	CreatePolicyTemplateVersion(w http.ResponseWriter, r *http.Request)
	GetPolicyTemplateVersion(w http.ResponseWriter, r *http.Request)
	DeletePolicyTemplateVersion(w http.ResponseWriter, r *http.Request)
	ListPolicyTemplateVersions(w http.ResponseWriter, r *http.Request)
}

func NewPolicyTemplateHandler(u usecase.Usecase) IPolicyTemplateHandler {
	return &PolicyTemplateHandler{
		usecase: u.PolicyTemplate,
	}
}

// CreatePolicyTemplate godoc
//
//	@Tags			PolicyTemplate
//	@Summary		[CreatePolicyTemplate] 정책 템플릿 신규 생성
//	@Description	정책 템플릿을 신규 생성(v1.0.0을 생성)한다.
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.CreatePolicyTemplateRequest	true	"create policy template request"
//	@Success		200		{object}	domain.CreatePolicyTemplateReponse
//	@Router			/api/1.0/admin/policytemplates [post]
//	@Security		JWT
func (h *PolicyTemplateHandler) CreatePolicyTemplate(w http.ResponseWriter, r *http.Request) {
	input := domain.CreatePolicyTemplateRequest{}

	err := UnmarshalRequestInput(r, &input)

	if err != nil {
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, r, err)
		return
	}

	var dto domain.PolicyTemplate
	if err = serializer.Map(input, &dto); err != nil {
		log.InfoWithContext(r.Context(), err)
	}

	policyTemplateId, err := h.usecase.Create(r.Context(), dto)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.CreatePolicyTemplateReponse
	out.ID = domain.PolicyTemplateId(policyTemplateId)

	ResponseJSON(w, r, http.StatusOK, out)
}

// UpdatePolicyTemplate godoc
//
//	@Tags			PolicyTemplate
//	@Summary		[UpdatePolicyTemplate] 정책 템플릿 업데이트
//	@Description	정책 템플릿의 업데이트 가능한 필드들을 업데이트한다.
//	@Accept			json
//	@Produce		json
//	@Param			policyTemplateId	path		string								true	"정책 템플릿 식별자(uuid)"
//	@Param			body				body		domain.UpdatePolicyTemplateRequest	true	"update  policy template request"
//	@Success		200					{object}	nil
//	@Router			/api/1.0/admin/policytemplates/{policyTemplateId} [patch]
//	@Security		JWT
func (h *PolicyTemplateHandler) UpdatePolicyTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	policyTemplateId, ok := vars["policyTemplateId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid policyTemplateId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	id, err := uuid.Parse(policyTemplateId)
	if err != nil {
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
			return
		}

		ErrorJSON(w, r, err)
		return
	}

	input := domain.UpdatePolicyTemplateRequest{}

	err = UnmarshalRequestInput(r, &input)

	if err != nil {
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, r, err)
		return
	}

	err = h.usecase.Update(r.Context(), id, input)

	if err != nil {
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
			return
		}

		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, "")
}

// DeletePolicyTemplate godoc
//
//	@Tags			PolicyTemplate
//	@Summary		[DeletePolicyTemplate] 정책 템플릿 삭제
//	@Description	정책 템플릿을 삭제한다.
//	@Accept			json
//	@Produce		json
//	@Param			policyTemplateId	path		string	true	"정책 템플릿 식별자(uuid)"
//	@Success		200					{object}	nil
//	@Router			/api/1.0/admin/policytemplates/{policyTemplateId} [delete]
//	@Security		JWT
func (h *PolicyTemplateHandler) DeletePolicyTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	policyTemplateId, ok := vars["policyTemplateId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid policyTemplateId"), "C_INVALID_POLICY_TEMPLATE_ID", ""))
		return
	}

	id, err := uuid.Parse(policyTemplateId)
	if err != nil {
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
			return
		}

		ErrorJSON(w, r, err)
		return
	}

	err = h.usecase.Delete(r.Context(), id)

	if err != nil {
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
			return
		}

		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, "")
}

// GetPolicyTemplate godoc
//
//	@Tags			PolicyTemplate
//	@Summary		[GetPolicyTemplate] 정책 템플릿 조회(최신 버전)
//	@Description	해당 식별자를 가진 정책 템플릿의 최신 버전을 조회한다.
//	@Accept			json
//	@Produce		json
//	@Param			policyTemplateId	path		string	true	"정책 템플릿 식별자(uuid)"
//	@Success		200					{object}	domain.GetPolicyTemplateResponse
//	@Router			/api/1.0/admin/policytemplates/{policyTemplateId} [get]
//	@Security		JWT
func (h *PolicyTemplateHandler) GetPolicyTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	policyTemplateId, ok := vars["policyTemplateId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid policyTemplateId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	id, err := uuid.Parse(policyTemplateId)
	if err != nil {
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
			return
		}

		ErrorJSON(w, r, err)
		return
	}

	policyTemplate, err := h.usecase.Get(r.Context(), id)
	if err != nil {
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
			return
		}

		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetPolicyTemplateResponse
	if err = serializer.Map(*policyTemplate, &out.PolicyTemplate); err != nil {
		log.ErrorWithContext(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// ListPolicyTemplate godoc
//
//	@Tags			PolicyTemplate
//	@Summary		[ListPolicyTemplate] 정책 템플릿 목록 조회
//	@Description	정책 템플릿 목록을 조회한다. 정책 템플릿 목록 조회 결과는 최신 템플릿 버전 목록만 조회된다.
//	@Accept			json
//	@Produce		json
//	@Param			limit		query		string		false	"pageSize"
//	@Param			page		query		string		false	"pageNumber"
//	@Param			soertColumn	query		string		false	"sortColumn"
//	@Param			sortOrder	query		string		false	"sortOrder"
//	@Param			filters		query		[]string	false	"filters"
//	@Success		200			{object}	domain.ListPolicyTemplateResponse
//	@Router			/api/1.0/admin/policytemplates [get]
//	@Security		JWT
func (h *PolicyTemplateHandler) ListPolicyTemplate(w http.ResponseWriter, r *http.Request) {
	urlParams := r.URL.Query()

	pg := pagination.NewPagination(&urlParams)

	policyTemplates, err := h.usecase.Fetch(r.Context(), pg)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.ListPolicyTemplateResponse
	out.PolicyTemplates = make([]domain.PolicyTemplateResponse, len(policyTemplates))
	for i, policyTemplate := range policyTemplates {
		if err := serializer.Map(policyTemplate, &out.PolicyTemplates[i]); err != nil {
			log.InfoWithContext(r.Context(), err)
			continue
		}
	}

	if out.Pagination, err = pg.Response(); err != nil {
		log.InfoWithContext(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// ListPolicyTemplateVersions godoc
//
//	@Tags			PolicyTemplate
//	@Summary		[ListPolicyTemplateVersions] 정책 템플릿 버전목록 조회
//	@Description	해당 식별자를 가진 정책 템플릿의 최신 버전을 조회한다.
//	@Accept			json
//	@Produce		json
//	@Param			policyTemplateId	path		string	true	"정책 템플릿 식별자(uuid)"
//	@Success		200					{object}	domain.ListPolicyTemplateVersionsResponse
//	@Router			/api/1.0/admin/policytemplates/{policyTemplateId}/versions [get]
//	@Security		JWT
func (h *PolicyTemplateHandler) ListPolicyTemplateVersions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	policyTemplateId, ok := vars["policyTemplateId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid policyTemplateId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	id, err := uuid.Parse(policyTemplateId)
	if err != nil {
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
			return
		}

		ErrorJSON(w, r, err)
		return
	}

	policyTemplateVersions, err := h.usecase.ListPolicyTemplateVersions(r.Context(), id)

	if err != nil {
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
			return
		}

		ErrorJSON(w, r, err)
		return
	}

	var out domain.ListPolicyTemplateVersionsResponse
	if err = serializer.Map(*policyTemplateVersions, &out); err != nil {
		log.ErrorWithContext(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// ListPolicyTemplateStatistics godoc
//
//	@Tags			PolicyTemplate
//	@Summary		[ListPolicyTemplateStatistics] 정책 템플릿 사용 카운트 조회
//	@Description	해당 식별자를 가진 정책 템플릿의 최신 버전을 조회한다. 전체 조직의 통계를 조회하려면 organizationId를 tks로 설정한다.
//	@Accept			json
//	@Produce		json
//	@Param			policyTemplateId	path		string	true	"정책 템플릿 식별자(uuid)"
//	@Success		200					{object}	domain.ListPolicyTemplateStatisticsResponse
//	@Router			/api/1.0/admin/policytemplates/{policyTemplateId}/statistics [get]
//	@Security		JWT
func (h *PolicyTemplateHandler) ListPolicyTemplateStatistics(w http.ResponseWriter, r *http.Request) {
	// result := domain.ListPolicyTemplateStatisticsResponse{
	// 	PolicyTemplateStatistics: []domain.PolicyTemplateStatistics{
	// 		{
	// 			OrganizationId:   util.UUIDGen(),
	// 			OrganizationName: "개발팀",
	// 			UsageCount:       10,
	// 		},
	// 		{
	// 			OrganizationId:   util.UUIDGen(),
	// 			OrganizationName: "운영팀",
	// 			UsageCount:       5,
	// 		},
	// 	},
	// }
	// util.JsonResponse(w, result)
}

// GetPolicyTemplateDeploy godoc
//
//	@Tags			PolicyTemplate
//	@Summary		[GetPolicyTemplateDeploy] 정책 템플릿 클러스터 별 설치 버전 조회
//	@Description	해당 식별자를 가진 정책 템플릿의 정책 템플릿 클러스터 별 설치 버전을 조회한다.
//	@Accept			json
//	@Produce		json
//	@Param			policyTemplateId	path		string	true	"정책 템플릿 식별자(uuid)"
//	@Success		200					{object}	domain.GetPolicyTemplateDeployResponse
//	@Router			/api/1.0/admin/policytemplates/{policyTemplateId}/deploy [get]
//	@Security		JWT
func (h *PolicyTemplateHandler) GetPolicyTemplateDeploy(w http.ResponseWriter, r *http.Request) {
	// c1 := util.UUIDGen()
	// c2 := util.UUIDGen()
	// c3 := util.UUIDGen()

	// result := domain.GetPolicyTemplateDeployResponse{
	// 	DeployVersion: map[string]string{
	// 		c1: "v1.0.1",
	// 		c2: "v1.1.0",
	// 		c3: "v1.1.0",
	// 	},
	// }
	// util.JsonResponse(w, result)
}

// GetPolicyTemplateVersion godoc
//
//	@Tags			PolicyTemplate
//	@Summary		[GetPolicyTemplateVersion] 정책 템플릿 특정 버전 조회
//	@Description	해당 식별자를 가진 정책 템플릿의 특정 버전을 조회한다.
//	@Accept			json
//	@Produce		json
//	@Param			policyTemplateId	path		string	true	"정책 템플릿 식별자(uuid)"
//	@Param			version				path		string	true	"조회할 버전(v0.0.0 형식)"
//	@Success		200					{object}	domain.GetPolicyTemplateVersionResponse
//	@Router			/api/1.0/admin/policytemplates/{policyTemplateId}/versions/{version} [get]
//	@Security		JWT
func (h *PolicyTemplateHandler) GetPolicyTemplateVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	policyTemplateId, ok := vars["policyTemplateId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid policyTemplateId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	version, ok := vars["version"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid version"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	id, err := uuid.Parse(policyTemplateId)
	if err != nil {
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
			return
		}

		ErrorJSON(w, r, err)
		return
	}

	policyTemplate, err := h.usecase.GetPolicyTemplateVersion(r.Context(), id, version)
	if err != nil {
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
			return
		}

		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetPolicyTemplateVersionResponse
	if err = serializer.Map(*policyTemplate, &out.PolicyTemplate); err != nil {
		log.ErrorWithContext(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// CreatePolicyTemplateVersion godoc
//
//	@Tags			PolicyTemplate
//	@Summary		[CreatePolicyTemplateVersion] 정책 템플릿 특정 버전 저장
//	@Description	해당 식별자를 가진 정책 템플릿의 특정 버전을 저장한다.
//	@Accept			json
//	@Produce		json
//	@Param			policyTemplateId	path		string										true	"정책 템플릿 식별자(uuid)"
//	@Param			body				body		domain.CreatePolicyTemplateVersionRequest	true	"create policy template version request"
//	@Success		200					{object}	domain.CreatePolicyTemplateVersionResponse
//	@Router			/api/1.0/admin/policytemplates/{policyTemplateId}/versions [post]
//	@Security		JWT
func (h *PolicyTemplateHandler) CreatePolicyTemplateVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	policyTemplateId, ok := vars["policyTemplateId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid policyTemplateId"), "C_INVALID_POLICY_TEMPLATE_ID", ""))
		return
	}

	id, err := uuid.Parse(policyTemplateId)
	if err != nil {
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
			return
		}

		ErrorJSON(w, r, err)
		return
	}

	input := domain.CreatePolicyTemplateVersionRequest{}

	err = UnmarshalRequestInput(r, &input)

	if err != nil {
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)
		ErrorJSON(w, r, err)
		return
	}

	currentVer, err := semver.NewVersion(input.CurrentVersion)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid currentVersion"), "C_INVALID_CURRENT_VERSION", fmt.Sprintf("invalid version format '%s'", input.CurrentVersion)))
		return
	}

	versionUpType := strings.ToLower(input.VersionUpType)

	var expectedVersion string

	switch versionUpType {
	case "major":
		newVersion := currentVer.IncMajor()
		expectedVersion = newVersion.Original()
	case "minor":
		newVersion := currentVer.IncMinor()
		expectedVersion = newVersion.Original()
	case "patch":
		newVersion := currentVer.IncPatch()
		expectedVersion = newVersion.Original()
	}

	if expectedVersion != input.ExpectedVersion {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid expectedVersion"), "C_INVALID_EXPECTED_VERSION", fmt.Sprintf("expected version mismatch '%s' != '%s'",
			input.ExpectedVersion, expectedVersion)))
	}

	createdVersion, err := h.usecase.CreatePolicyTemplateVersion(r.Context(), id, expectedVersion, input.ParametersSchema, input.Rego, input.Libs)

	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.CreatePolicyTemplateVersionResponse
	out.Version = createdVersion

	ResponseJSON(w, r, http.StatusOK, out)
}

// DeletePolicyTemplateVersion godoc
//
//	@Tags			PolicyTemplate
//	@Summary		[DeletePolicyTemplateVersion] 정책 템플릿 특정 버전 삭제
//	@Description	해당 식별자를 가진 정책 템플릿의 특정 버전을 삭제한다.
//	@Accept			json
//	@Produce		json
//	@Param			policyTemplateId	path		string	true	"정책 템플릿 식별자(uuid)"
//	@Param			version				path		string	true	"삭제할 버전(v0.0.0 형식)"
//	@Success		200					{object}	nil
//	@Router			/api/1.0/admin/policytemplates/{policyTemplateId}/versions/{version} [delete]
//	@Security		JWT
func (h *PolicyTemplateHandler) DeletePolicyTemplateVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	policyTemplateId, ok := vars["policyTemplateId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid policyTemplateId"), "C_INVALID_POLICY_TEMPLATE_ID", ""))
		return
	}

	version, ok := vars["version"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid version"), "C_INVALID_VERSION", ""))
		return
	}

	id, err := uuid.Parse(policyTemplateId)
	if err != nil {
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
			return
		}

		ErrorJSON(w, r, err)
		return
	}

	err = h.usecase.DeletePolicyTemplateVersion(r.Context(), id, version)

	if err != nil {
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)
		if _, status := httpErrors.ErrorResponse(err); status == http.StatusNotFound {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(err, "", ""))
			return
		}

		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, "")
}

// ExistsPolicyTemplateName godoc
//
//	@Tags			PolicyTemplate
//	@Summary		[ExistsPolicyTemplateName] 정책 템플릿 아름 존재 여부 확인
//	@Description	해당 이름을 가진 정책 템플릿이 이미 존재하는지 확인한다.
//	@Accept			json
//	@Produce		json
//	@Param			policyTemplateName	path		string	true	"정책 템플릿 이름"
//	@Success		200					{object}	domain.CheckExistedResponse
//	@Router			/api/1.0/admin/policytemplates/name/{policyTemplateName}/existence [get]
//	@Security		JWT
func (h *PolicyTemplateHandler) ExistsPolicyTemplateName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	policyTemplateName, ok := vars["policyTemplateName"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("policyTemplateName not found in path"),
			"C_INVALID_POLICYTEMPLATE_NAME", ""))
		return
	}

	exist, err := h.usecase.IsPolicyTemplateNameExist(r.Context(), policyTemplateName)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.CheckExistedResponse
	out.Existed = exist

	ResponseJSON(w, r, http.StatusOK, out)
}

// ExistsPolicyTemplateKind godoc
//
//	@Tags			PolicyTemplate
//	@Summary		[ExistsPolicyTemplateKind] 정책 템플릿 유형 존재 여부 확인
//	@Description	해당 유형을 가진 정책 템플릿이 이미 존재하는지 확인한다.
//	@Accept			json
//	@Produce		json
//	@Param			policyTemplateKind	path		string	true	"정책 템플릿 이름"
//	@Success		200					{object}	domain.ExistsPolicyTemplateKindResponse
//	@Router			/api/1.0/admin/policytemplates/kind/{policyTemplateKind}/existence [get]
//	@Security		JWT
func (h *PolicyTemplateHandler) ExistsPolicyTemplateKind(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	policyTemplateKind, ok := vars["policyTemplateKind"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("policyTemplateKind not found in path"),
			"C_INVALID_POLICY_TEMPLATE_KIND", ""))
		return
	}

	exist, err := h.usecase.IsPolicyTemplateKindExist(r.Context(), policyTemplateKind)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.CheckExistedResponse
	out.Existed = exist

	ResponseJSON(w, r, http.StatusOK, out)
}
