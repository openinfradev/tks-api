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

type AlertTemplateHandler struct {
	usecase usecase.IAlertTemplateUsecase
}

func NewAlertTemplateHandler(h usecase.Usecase) *AlertTemplateHandler {
	return &AlertTemplateHandler{
		usecase: h.AlertTemplate,
	}
}

// CreateAlertTemplate godoc
//
//	@Tags			AlertTemplates
//	@Summary		Create alert template. ADMIN ONLY
//	@Description	Create alert template. ADMIN ONLY
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	domain_admin.CreateAlertTemplateResponse
//	@Router			/admin/alert-templates [post]
//	@Security		JWT
func (h *AlertTemplateHandler) CreateAlertTemplate(w http.ResponseWriter, r *http.Request) {
	input := domain_admin.CreateAlertTemplateRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var dto model.AlertTemplate
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

	out := domain_admin.CreateAlertTemplateResponse{
		ID: id.String(),
	}
	ResponseJSON(w, r, http.StatusOK, out)
}

// GetAlertTemplate godoc
//
//	@Tags			AlertTemplates
//	@Summary		Get AlertTemplates
//	@Description	Get AlertTemplates
//	@Accept			json
//	@Produce		json
//	@Param			limit		query		string		false	"pageSize"
//	@Param			page		query		string		false	"pageNumber"
//	@Param			soertColumn	query		string		false	"sortColumn"
//	@Param			sortOrder	query		string		false	"sortOrder"
//	@Param			filters		query		[]string	false	"filters"
//	@Success		200			{object}	domain_admin.GetAlertTemplatesResponse
//	@Router			/admin/alert-templates [get]
//	@Security		JWT
func (h *AlertTemplateHandler) GetAlertTemplates(w http.ResponseWriter, r *http.Request) {
	urlParams := r.URL.Query()
	pg := pagination.NewPagination(&urlParams)
	alertTemplates, err := h.usecase.Fetch(r.Context(), pg)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain_admin.GetAlertTemplatesResponse
	out.AlertTemplates = make([]domain_admin.AlertTemplateResponse, len(alertTemplates))
	for i, alertTemplate := range alertTemplates {
		if err := serializer.Map(r.Context(), alertTemplate, &out.AlertTemplates[i]); err != nil {
			log.Info(r.Context(), err)
		}

		out.AlertTemplates[i].Organizations = make([]domain.SimpleOrganizationResponse, len(alertTemplate.Organizations))
		for j, organization := range alertTemplate.Organizations {
			if err := serializer.Map(r.Context(), organization, &out.AlertTemplates[i].Organizations[j]); err != nil {
				log.Info(r.Context(), err)
			}
		}
	}

	if out.Pagination, err = pg.Response(r.Context()); err != nil {
		log.Info(r.Context(), err)
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// GetAlertTemplate godoc
//
//	@Tags			AlertTemplates
//	@Summary		Get AlertTemplate
//	@Description	Get AlertTemplate
//	@Accept			json
//	@Produce		json
//	@Param			alertTemplateId	path		string	true	"alertTemplateId"
//	@Success		200				{object}	domain_admin.GetAlertTemplateResponse
//	@Router			/admin/alert-templates/{alertTemplateId} [get]
//	@Security		JWT
func (h *AlertTemplateHandler) GetAlertTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strId, ok := vars["alertTemplateId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid alertTemplateId"), "C_INVALID_ALERT_TEMPLATE_ID", ""))
		return
	}

	alertTemplateId, err := uuid.Parse(strId)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(errors.Wrap(err, "Failed to parse uuid %s"), "C_INVALID_ALERT_TEMPLATE_ID", ""))
		return
	}

	alertTemplate, err := h.usecase.Get(r.Context(), alertTemplateId)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	var out domain_admin.GetAlertTemplateResponse
	if err := serializer.Map(r.Context(), alertTemplate, &out.AlertTemplate); err != nil {
		log.Info(r.Context(), err)
	}

	out.AlertTemplate.Organizations = make([]domain.SimpleOrganizationResponse, len(alertTemplate.Organizations))
	for i, organization := range alertTemplate.Organizations {
		if err := serializer.Map(r.Context(), organization, &out.AlertTemplate.Organizations[i]); err != nil {
			log.Info(r.Context(), err)
			continue
		}
	}

	out.AlertTemplate.MetricParameters = make([]domain_admin.MetricParameterResponse, len(alertTemplate.MetricParameters))
	for i, metricParameters := range alertTemplate.MetricParameters {
		if err := serializer.Map(r.Context(), metricParameters, &out.AlertTemplate.MetricParameters[i]); err != nil {
			log.Info(r.Context(), err)
			continue
		}
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

// UpdateAlertTemplate godoc
//
//	@Tags			AlertTemplates
//	@Summary		Update AlertTemplate
//	@Description	Update AlertTemplate
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain_admin.UpdateAlertTemplateRequest	true	"Update alert template request"
//	@Success		200		{object}	nil
//	@Router			/admin/alert-templates/{alertTemplateId} [put]
//	@Security		JWT
func (h *AlertTemplateHandler) UpdateAlertTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strId, ok := vars["alertTemplateId"]
	if !ok {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid alertTemplateId"), "C_INVALID_ALERT_TEMPLATE_ID", ""))
		return
	}
	alertTemplateId, err := uuid.Parse(strId)
	if err != nil {
		ErrorJSON(w, r, httpErrors.NewBadRequestError(errors.Wrap(err, "Failed to parse uuid %s"), "C_INVALID_ALERT_TEMPLATE_ID", ""))
		return
	}

	input := domain_admin.UpdateAlertTemplateRequest{}
	err = UnmarshalRequestInput(r, &input)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}
	var dto model.AlertTemplate
	if err = serializer.Map(r.Context(), input, &dto); err != nil {
		log.Info(r.Context(), err)
	}
	dto.ID = alertTemplateId
	dto.MetricParameters = make([]model.MetricParameter, len(input.MetricParameters))
	for i, metricParameter := range input.MetricParameters {
		if err := serializer.Map(r.Context(), metricParameter, &dto.MetricParameters[i]); err != nil {
			log.Info(r.Context(), err)
		}
		dto.MetricParameters[i].AlertTemplateId = alertTemplateId
	}

	err = h.usecase.Update(r.Context(), dto)
	if err != nil {
		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusOK, nil)
}
