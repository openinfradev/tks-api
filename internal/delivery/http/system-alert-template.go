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
	domain "github.com/openinfradev/tks-api/pkg/domain"
	domain_admin "github.com/openinfradev/tks-api/pkg/domain/admin"
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
//	@Success		200	{object}	domain_admin.CreateSystemNotificationTemplateResponse
//	@Router			/admin/system-notification-templates [post]
//	@Security		JWT
func (h *SystemNotificationTemplateHandler) CreateSystemNotificationTemplate(w http.ResponseWriter, r *http.Request) {
	input := domain_admin.CreateSystemNotificationTemplateRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var dto model.SystemNotificationTemplate
	if err = serializer.Map(r.Context(), input, &dto); err != nil {
		log.Info(r.Context(), err)
	}
	dto.MetricParameters = make([]model.MetricParameter, len(input.MetricParameters))
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

	out := domain_admin.CreateSystemNotificationTemplateResponse{
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
//	@Param			limit		query		string		false	"pageSize"
//	@Param			page		query		string		false	"pageNumber"
//	@Param			soertColumn	query		string		false	"sortColumn"
//	@Param			sortOrder	query		string		false	"sortOrder"
//	@Param			filters		query		[]string	false	"filters"
//	@Success		200			{object}	domain_admin.GetSystemNotificationTemplatesResponse
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

	var out domain_admin.GetSystemNotificationTemplatesResponse
	out.SystemNotificationTemplates = make([]domain_admin.SystemNotificationTemplateResponse, len(systemNotificationTemplates))
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
//	@Success		200				{object}	domain_admin.GetSystemNotificationTemplateResponse
//	@Router			/admin/system-notification-templates/{systemNotificationTemplateId} [get]
//	@Security		JWT
func (h *SystemNotificationTemplateHandler) GetSystemNotificationTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strId, ok := vars["systemNotificationTemplateId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid systemNotificationTemplateId"), "C_INVALID_ALERT_TEMPLATE_ID", ""))
		return
	}

	systemNotificationTemplateId, err := uuid.Parse(strId)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(errors.Wrap(err, "Failed to parse uuid %s"), "C_INVALID_ALERT_TEMPLATE_ID", ""))
		return
	}

	systemNotificationTemplate, err := h.usecase.Get(r.Context(), systemNotificationTemplateId)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain_admin.GetSystemNotificationTemplateResponse
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

	out.SystemNotificationTemplate.MetricParameters = make([]domain_admin.MetricParameterResponse, len(systemNotificationTemplate.MetricParameters))
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
//	@Param			body	body		domain_admin.UpdateSystemNotificationTemplateRequest	true	"Update alert template request"
//	@Success		200		{object}	nil
//	@Router			/admin/system-notification-templates/{systemNotificationTemplateId} [put]
//	@Security		JWT
func (h *SystemNotificationTemplateHandler) UpdateSystemNotificationTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strId, ok := vars["systemNotificationTemplateId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid systemNotificationTemplateId"), "C_INVALID_ALERT_TEMPLATE_ID", ""))
		return
	}
	systemNotificationTemplateId, err := uuid.Parse(strId)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(errors.Wrap(err, "Failed to parse uuid %s"), "C_INVALID_ALERT_TEMPLATE_ID", ""))
		return
	}

	input := domain_admin.UpdateSystemNotificationTemplateRequest{}
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
	dto.MetricParameters = make([]model.MetricParameter, len(input.MetricParameters))
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
