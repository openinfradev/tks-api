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

type SystemNotificationTemplateHandler struct {
	usecase usecase.ISystemNotificationTemplateUsecase
}

func NewSystemNotificationTemplateHandler(h usecase.Usecase) *SystemNotificationTemplateHandler {
	return &SystemNotificationTemplateHandler{
		usecase: h.SystemNotificationTemplate,
	}
}

// CreateSystemNotificationTemplate godoc
//
//	@Tags			SystemNotificationTemplates
//	@Summary		Create alert template. ADMIN ONLY
//	@Description	Create alert template. ADMIN ONLY
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	domain.CreateSystemNotificationTemplateRequest
//	@Router			/admin/system-notification-templates [post]
//	@Security		JWT
func (h *SystemNotificationTemplateHandler) CreateSystemNotificationTemplate(w http.ResponseWriter, r *http.Request) {
	input := domain.CreateSystemNotificationTemplateRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var dto model.SystemNotificationTemplate
	if err = serializer.Map(r.Context(), input, &dto); err != nil {
		log.Info(r.Context(), err)
	}
	dto.MetricParameters = make([]model.SystemNotificationMetricParameter, len(input.MetricParameters))
	for i, metricParameter := range input.MetricParameters {
		if err := serializer.Map(r.Context(), metricParameter, &dto.MetricParameters[i]); err != nil {
			log.Info(r.Context(), err)
		}
	}

	id, err := h.usecase.Create(r.Context(), dto)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	out := domain.CreateSystemNotificationTemplateResponse{
		ID: id.String(),
	}
	ResponseJSON(w, r, http.StatusOK, out)
}

// GetSystemNotificationTemplate godoc
//
//	@Tags			SystemNotificationTemplates
//	@Summary		Get SystemNotificationTemplates
//	@Description	Get SystemNotificationTemplates
//	@Accept			json
//	@Produce		json
//	@Param			pageSize	query		string		false	"pageSize"
//	@Param			pageNumber	query		string		false	"pageNumber"
//	@Param			soertColumn	query		string		false	"sortColumn"
//	@Param			sortOrder	query		string		false	"sortOrder"
//	@Param			filters		query		[]string	false	"filters"
//	@Success		200			{object}	domain.GetSystemNotificationTemplatesResponse
//	@Router			/admin/system-notification-templates [get]
//	@Security		JWT
func (h *SystemNotificationTemplateHandler) GetSystemNotificationTemplates(w http.ResponseWriter, r *http.Request) {
	urlParams := r.URL.Query()
	pg := pagination.NewPagination(&urlParams)
	systemNotificationTemplates, err := h.usecase.Fetch(r.Context(), pg)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetSystemNotificationTemplatesResponse
	out.SystemNotificationTemplates = make([]domain.SystemNotificationTemplateResponse, len(systemNotificationTemplates))
	for i, systemNotificationTemplate := range systemNotificationTemplates {
		if err := serializer.Map(r.Context(), systemNotificationTemplate, &out.SystemNotificationTemplates[i]); err != nil {
			log.Info(r.Context(), err)
		}

		out.SystemNotificationTemplates[i].Organizations = make([]domain.SimpleOrganizationResponse, len(systemNotificationTemplate.Organizations))
		for j, organization := range systemNotificationTemplate.Organizations {
			if err := serializer.Map(r.Context(), organization, &out.SystemNotificationTemplates[i].Organizations[j]); err != nil {
				log.Info(r.Context(), err)
			}
		}
	}

	if out.Pagination, err = pg.Response(r.Context()); err != nil {
		log.Info(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetSystemNotificationTemplate godoc
//
//	@Tags			SystemNotificationTemplates
//	@Summary		Get SystemNotificationTemplate
//	@Description	Get SystemNotificationTemplate
//	@Accept			json
//	@Produce		json
//	@Param			systemNotificationTemplateId	path		string	true	"systemNotificationTemplateId"
//	@Success		200								{object}	domain.GetSystemNotificationTemplateResponse
//	@Router			/admin/system-notification-templates/{systemNotificationTemplateId} [get]
//	@Security		JWT
func (h *SystemNotificationTemplateHandler) GetSystemNotificationTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strId, ok := vars["systemNotificationTemplateId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid systemNotificationTemplateId"), "C_INVALID_SYSTEM_NOTIFICATION_TEMPLATE_ID", ""))
		return
	}

	systemNotificationTemplateId, err := uuid.Parse(strId)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(errors.Wrap(err, "Failed to parse uuid %s"), "C_INVALID_SYSTEM_NOTIFICATION_TEMPLATE_ID", ""))
		return
	}

	systemNotificationTemplate, err := h.usecase.Get(r.Context(), systemNotificationTemplateId)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetSystemNotificationTemplateResponse
	if err := serializer.Map(r.Context(), systemNotificationTemplate, &out.SystemNotificationTemplate); err != nil {
		log.Info(r.Context(), err)
	}

	out.SystemNotificationTemplate.Organizations = make([]domain.SimpleOrganizationResponse, len(systemNotificationTemplate.Organizations))
	for i, organization := range systemNotificationTemplate.Organizations {
		if err := serializer.Map(r.Context(), organization, &out.SystemNotificationTemplate.Organizations[i]); err != nil {
			log.Info(r.Context(), err)
			continue
		}
	}

	out.SystemNotificationTemplate.MetricParameters = make([]domain.SystemNotificationMetricParameterResponse, len(systemNotificationTemplate.MetricParameters))
	for i, metricParameters := range systemNotificationTemplate.MetricParameters {
		if err := serializer.Map(r.Context(), metricParameters, &out.SystemNotificationTemplate.MetricParameters[i]); err != nil {
			log.Info(r.Context(), err)
			continue
		}
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// UpdateSystemNotificationTemplate godoc
//
//	@Tags			SystemNotificationTemplates
//	@Summary		Update SystemNotificationTemplate
//	@Description	Update SystemNotificationTemplate
//	@Accept			json
//	@Produce		json
//	@Param			systemNotificationTemplateId	path		string											true	"systemNotificationTemplateId"
//	@Param			body							body		domain.UpdateSystemNotificationTemplateRequest	true	"Update alert template request"
//	@Success		200								{object}	nil
//	@Router			/admin/system-notification-templates/{systemNotificationTemplateId} [put]
//	@Security		JWT
func (h *SystemNotificationTemplateHandler) UpdateSystemNotificationTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strId, ok := vars["systemNotificationTemplateId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid systemNotificationTemplateId"), "C_INVALID_SYSTEM_NOTIFICATION_TEMPLATE_ID", ""))
		return
	}
	systemNotificationTemplateId, err := uuid.Parse(strId)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(errors.Wrap(err, "Failed to parse uuid %s"), "C_INVALID_SYSTEM_NOTIFICATION_TEMPLATE_ID", ""))
		return
	}

	input := domain.UpdateSystemNotificationTemplateRequest{}
	err = UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}
	var dto model.SystemNotificationTemplate
	if err = serializer.Map(r.Context(), input, &dto); err != nil {
		log.Info(r.Context(), err)
	}
	dto.ID = systemNotificationTemplateId
	dto.MetricParameters = make([]model.SystemNotificationMetricParameter, len(input.MetricParameters))
	for i, metricParameter := range input.MetricParameters {
		if err := serializer.Map(r.Context(), metricParameter, &dto.MetricParameters[i]); err != nil {
			log.Info(r.Context(), err)
		}
		dto.MetricParameters[i].SystemNotificationTemplateId = systemNotificationTemplateId
	}

	err = h.usecase.Update(r.Context(), dto)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, nil)
}

// DeleteSystemNotificationTemplate godoc
//
//	@Tags			SystemNotificationTemplates
//	@Summary		Delete SystemNotificationTemplate
//	@Description	Delete SystemNotificationTemplate
//	@Accept			json
//	@Produce		json
//	@Param			systemNotificationTemplateId	path		string	true	"systemNotificationTemplateId"
//	@Success		200								{object}	nil
//	@Router			/admin/system-notification-templates/{systemNotificationTemplateId} [delete]
//	@Security		JWT
func (h *SystemNotificationTemplateHandler) DeleteSystemNotificationTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strId, ok := vars["systemNotificationTemplateId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid systemNotificationTemplateId"), "C_INVALID_SYSTEM_NOTIFICATION_TEMPLATE_ID", ""))
		return
	}
	systemNotificationTemplateId, err := uuid.Parse(strId)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(errors.Wrap(err, "Failed to parse uuid %s"), "C_INVALID_SYSTEM_NOTIFICATION_TEMPLATE_ID", ""))
		return
	}

	err = h.usecase.Delete(r.Context(), systemNotificationTemplateId)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, nil)
}

// GetOrganizationSystemNotificationTemplates godoc
//
//	@Tags			SystemNotificationTemplates
//	@Summary		Get Organization SystemNotificationTemplates
//	@Description	Get Organization SystemNotificationTemplates
//	@Accept			json
//	@Produce		json
//	@Param			pageSize	query		string		false	"pageSize"
//	@Param			pageNumber	query		string		false	"pageNumber"
//	@Param			soertColumn	query		string		false	"sortColumn"
//	@Param			sortOrder	query		string		false	"sortOrder"
//	@Param			filters		query		[]string	false	"filters"
//	@Success		200			{object}	domain.GetSystemNotificationTemplatesResponse
//	@Router			/organizations/{organizationId}/system-notification-templates [get]
//	@Security		JWT
func (h *SystemNotificationTemplateHandler) GetOrganizationSystemNotificationTemplates(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	urlParams := r.URL.Query()
	pg := pagination.NewPagination(&urlParams)
	systemNotificationTemplates, err := h.usecase.FetchWithOrganization(r.Context(), organizationId, pg)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain.GetSystemNotificationTemplatesResponse
	out.SystemNotificationTemplates = make([]domain.SystemNotificationTemplateResponse, len(systemNotificationTemplates))
	for i, systemNotificationTemplate := range systemNotificationTemplates {
		if err := serializer.Map(r.Context(), systemNotificationTemplate, &out.SystemNotificationTemplates[i]); err != nil {
			log.Info(r.Context(), err)
		}

		out.SystemNotificationTemplates[i].Organizations = make([]domain.SimpleOrganizationResponse, len(systemNotificationTemplate.Organizations))
		for j, organization := range systemNotificationTemplate.Organizations {
			if err := serializer.Map(r.Context(), organization, &out.SystemNotificationTemplates[i].Organizations[j]); err != nil {
				log.Info(r.Context(), err)
			}
		}
	}

	if out.Pagination, err = pg.Response(r.Context()); err != nil {
		log.Info(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// AddOrganizationSystemNotificationTemplates godoc
//
//	@Tags			SystemNotificationTemplates
//	@Summary		Add organization systemNotificationTemplates
//	@Description	Add organization systemNotificationTemplates
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string														true	"organizationId"
//	@Param			body			body		domain.AddOrganizationSystemNotificationTemplatesRequest	true	"Add organization systemNotification templates request"
//	@Success		200				{object}	nil
//	@Router			/organizations/{organizationId}/system-notification-templates [post]
//	@Security		JWT
func (h *SystemNotificationTemplateHandler) AddOrganizationSystemNotificationTemplates(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	input := domain.AddOrganizationSystemNotificationTemplatesRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	err = h.usecase.AddOrganizationSystemNotificationTemplates(r.Context(), organizationId, input.SystemNotificationTemplateIds)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}
	ResponseJSON(w, r, http.StatusOK, nil)
}

// RemoveOrganizationSystemNotificationTemplates godoc
//
//	@Tags			SystemNotificationTemplates
//	@Summary		Remove organization systemNotificationTemplates
//	@Description	Remove organization systemNotificationTemplates
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string														true	"organizationId"
//	@Param			body			body		domain.RemoveOrganizationSystemNotificationTemplatesRequest	true	"Remove organization systemNotification templates request"
//	@Success		200				{object}	nil
//	@Router			/organizations/{organizationId}/system-notification-templates [put]
//	@Security		JWT
func (h *SystemNotificationTemplateHandler) RemoveOrganizationSystemNotificationTemplates(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	organizationId, ok := vars["organizationId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid organizationId"), "C_INVALID_ORGANIZATION_ID", ""))
		return
	}

	input := domain.RemoveOrganizationSystemNotificationTemplatesRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	err = h.usecase.RemoveOrganizationSystemNotificationTemplates(r.Context(), organizationId, input.SystemNotificationTemplateIds)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}
	ResponseJSON(w, r, http.StatusOK, nil)
}

// CheckSystemNotificationTemplateName godoc
//
//	@Tags			SystemNotificationTemplates
//	@Summary		Check name for systemNotificationTemplate
//	@Description	Check name for systemNotificationTemplate
//	@Accept			json
//	@Produce		json
//	@Param			name	path		string	true	"name"
//	@Success		200		{object}	domain.CheckSystemNotificaionTemplateNameResponse
//	@Router			/admin/system-notification-templates/name/{name}/existence [GET]
//	@Security		JWT
func (h *SystemNotificationTemplateHandler) CheckSystemNotificationTemplateName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name, ok := vars["name"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("Invalid name"), "ST_INVALID_SYSTEM_NOTIFICATION_TEMAPLTE_NAME", ""))
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

	var out domain.CheckSystemNotificaionTemplateNameResponse
	out.Existed = exist

	ResponseJSON(w, r, http.StatusOK, out)
}
