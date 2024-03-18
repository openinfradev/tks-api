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
	"github.com/pkg/errors"
)

type SystemNotificationRuleHandler struct {
	usecase usecase.ISystemNotificationRuleUsecase
}

func NewSystemNotificationRuleHandler(h usecase.Usecase) *SystemNotificationRuleHandler {
	return &SystemNotificationRuleHandler{
		usecase: h.SystemNotificationRule,
	}
}

// CreateSystemNotificationRule godoc
//
//	@Tags			SystemNotificationRules
//	@Summary		Create SystemNotificationRule
//	@Description	Create SystemNotificationRule
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.CreateSystemNotificationRuleRequest	true	"create stack template request"
//	@Success		200		{object}	domain.CreateSystemNotificationRuleResponse
//	@Router			/admin/system-notification-rules [post]
//	@Security		JWT
func (h *SystemNotificationRuleHandler) CreateSystemNotificationRule(w http.ResponseWriter, r *http.Request) {
	input := domain.CreateSystemNotificationRuleRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var dto model.SystemNotificationRule
	if err = serializer.Map(r.Context(), input, &dto); err != nil {
		log.Info(r.Context(), err)
	}

	id, err := h.usecase.Create(r.Context(), dto)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	out := domain.CreateSystemNotificationRuleResponse{
		ID: id.String(),
	}
	ResponseJSON(w, r, http.StatusOK, out)
}

// GetSystemNotificationRule godoc
//
//	@Tags			SystemNotificationRules
//	@Summary		Get SystemNotificationRules
//	@Description	Get SystemNotificationRules
//	@Accept			json
//	@Produce		json
//	@Param			limit		query		string		false	"pageSize"
//	@Param			page		query		string		false	"pageNumber"
//	@Param			soertColumn	query		string		false	"sortColumn"
//	@Param			sortOrder	query		string		false	"sortOrder"
//	@Param			filters		query		[]string	false	"filters"
//	@Success		200			{object}	domain.GetSystemNotificationRulesResponse
//	@Router			/admin/system-notification-rules [get]
//	@Security		JWT
func (h *SystemNotificationRuleHandler) GetSystemNotificationRules(w http.ResponseWriter, r *http.Request) {
	urlParams := r.URL.Query()
	pg := pagination.NewPagination(&urlParams)
	systemNotificationRules, err := h.usecase.Fetch(r.Context(), pg)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetSystemNotificationRulesResponse
	out.SystemNotificationRules = make([]domain.SystemNotificationRuleResponse, len(systemNotificationRules))
	for i, systemNotificationRule := range systemNotificationRules {
		if err := serializer.Map(r.Context(), systemNotificationRule, &out.SystemNotificationRules[i]); err != nil {
			log.Info(r.Context(), err)
		}
	}

	if out.Pagination, err = pg.Response(r.Context()); err != nil {
		log.Info(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetSystemNotificationRule godoc
//
//	@Tags			SystemNotificationRules
//	@Summary		Get SystemNotificationRule
//	@Description	Get SystemNotificationRule
//	@Accept			json
//	@Produce		json
//	@Param			systemNotificationRuleId	path		string	true	"systemNotificationRuleId"
//	@Success		200				{object}	domain.GetSystemNotificationRuleResponse
//	@Router			/admin/system-notification-rules/{systemNotificationRuleId} [get]
//	@Security		JWT
func (h *SystemNotificationRuleHandler) GetSystemNotificationRule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strId, ok := vars["systemNotificationRuleId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid systemNotificationRuleId"), "C_INVALID_STACK_TEMPLATE_ID", ""))
		return
	}

	systemNotificationRuleId, err := uuid.Parse(strId)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(errors.Wrap(err, "Failed to parse uuid %s"), "C_INVALID_STACK_TEMPLATE_ID", ""))
		return
	}

	systemNotificationRule, err := h.usecase.Get(r.Context(), systemNotificationRuleId)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetSystemNotificationRuleResponse
	if err := serializer.Map(r.Context(), systemNotificationRule, &out.SystemNotificationRule); err != nil {
		log.Info(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// UpdateSystemNotificationRule godoc
//
//	@Tags			SystemNotificationRules
//	@Summary		Update SystemNotificationRule
//	@Description	Update SystemNotificationRule
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.UpdateSystemNotificationRuleRequest	true	"Update systemNotificationRule request"
//	@Success		200		{object}	nil
//	@Router			/admin/system-notification-rules/{systemNotificationRuleId} [put]
//	@Security		JWT
func (h *SystemNotificationRuleHandler) UpdateSystemNotificationRule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strId, ok := vars["systemNotificationRuleId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid systemNotificationRuleId"), "C_INVALID_STACK_TEMPLATE_ID", ""))
		return
	}

	systemNotificationRuleId, err := uuid.Parse(strId)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(errors.Wrap(err, "Failed to parse uuid %s"), "C_INVALID_STACK_TEMPLATE_ID", ""))
		return
	}

	var dto model.SystemNotificationRule
	if err := serializer.Map(r.Context(), r, &dto); err != nil {
		log.Info(r.Context(), err)
	}
	dto.ID = systemNotificationRuleId

	err = h.usecase.Update(r.Context(), dto)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}
	ResponseJSON(w, r, http.StatusOK, nil)
}

// DeleteSystemNotificationRule godoc
//
//	@Tags			SystemNotificationRules
//	@Summary		Delete SystemNotificationRule
//	@Description	Delete SystemNotificationRule
//	@Accept			json
//	@Produce		json
//	@Param			systemNotificationRuleId	path		string	true	"systemNotificationRuleId"
//	@Success		200				{object}	nil
//	@Router			/admin/system-notification-rules/{systemNotificationRuleId} [delete]
//	@Security		JWT
func (h *SystemNotificationRuleHandler) DeleteSystemNotificationRule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	_, ok := vars["systemNotificationRuleId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid systemNotificationRuleId"), "C_INVALID_STACK_TEMPLATE_ID", ""))
		return
	}

	ErrorJSON(w, r, fmt.Errorf("need implementation"))
}

// GetOrganizationSystemNotificationRules godoc
//
//	@Tags			SystemNotificationRules
//	@Summary		Get Organization SystemNotificationRules
//	@Description	Get Organization SystemNotificationRules
//	@Accept			json
//	@Produce		json
//	@Param			limit		query		string		false	"pageSize"
//	@Param			page		query		string		false	"pageNumber"
//	@Param			soertColumn	query		string		false	"sortColumn"
//	@Param			sortOrder	query		string		false	"sortOrder"
//	@Param			filters		query		[]string	false	"filters"
//	@Success		200			{object}	domain.GetSystemNotificationRulesResponse
//	@Router			/organizations/{organizationId}/system-notification-rules [get]
//	@Security		JWT
func (h *SystemNotificationRuleHandler) GetOrganizationSystemNotificationRules(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	urlParams := r.URL.Query()
	pg := pagination.NewPagination(&urlParams)
	systemNotificationRules, err := h.usecase.FetchWithOrganization(r.Context(), organizationId, pg)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetSystemNotificationRulesResponse
	out.SystemNotificationRules = make([]domain.SystemNotificationRuleResponse, len(systemNotificationRules))
	for i, systemNotificationRule := range systemNotificationRules {
		if err := serializer.Map(r.Context(), systemNotificationRule, &out.SystemNotificationRules[i]); err != nil {
			log.Info(r.Context(), err)
		}
	}

	if out.Pagination, err = pg.Response(r.Context()); err != nil {
		log.Info(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// CheckSystemNotificationRuleName godoc
//
//	@Tags			SystemNotificationRules
//	@Summary		Check name for systemNotificationRule
//	@Description	Check name for systemNotificationRule
//	@Accept			json
//	@Produce		json
//	@Param			name	path		string	true	"name"
//	@Success		200		{object}	domain.CheckSystemNotificationRuleNameResponse
//	@Router			/admin/system-notification-rules/name/{name}/existence [GET]
//	@Security		JWT
func (h *SystemNotificationRuleHandler) CheckSystemNotificationRuleName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name, ok := vars["name"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid name"), "SNR_INVALID_STACK_TEMAPLTE_NAME", ""))
		return
	}

	exist := true
	_, err := h.usecase.GetByName(r.Context(), name)
	if err != nil {
		if _, code := httpErrors.ErrorResponse(err); code == http.StatusNotFound {
			exist = false
		} else {
			ErrorJSON(w, r, err)
			return
		}
	}

	var out domain.CheckSystemNotificationRuleNameResponse
	out.Existed = exist

	ResponseJSON(w, r, http.StatusOK, out)
}
