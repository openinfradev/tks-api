package admin

import (
	"time"

	"github.com/openinfradev/tks-api/pkg/domain"
)

type AlertTemplateResponse struct {
	ID               string                              `json:"id"`
	Name             string                              `json:"name"`
	Description      string                              `json:"description"`
	MetricQuery      string                              `json:"metricQuery" validate:"required"`
	MetricParameters []MetricParameterResponse           `json:"metricParameters,omitempty"`
	Organizations    []domain.SimpleOrganizationResponse `json:"organizations,omitempty"`
	Creator          domain.SimpleUserResponse           `json:"creator"`
	Updator          domain.SimpleUserResponse           `json:"updator"`
	CreatedAt        time.Time                           `json:"createdAt"`
	UpdatedAt        time.Time                           `json:"updatedAt"`
}

type MetricParameterResponse struct {
	Order int    `json:"order" validate:"required"`
	Key   string `json:"key" validate:"required,name"`
	Value string `json:"value" validate:"required"`
}

type CreateAlertTemplateRequest struct {
	Name             string                    `json:"name" validate:"required,name"`
	Description      string                    `json:"description"`
	OrganizationIds  []string                  `json:"organizationIds" validate:"required"`
	MetricQuery      string                    `json:"metricQuery" validate:"required"`
	MetricParameters []MetricParameterResponse `json:"metricParameters"`
}

type CreateAlertTemplateResponse struct {
	ID string `json:"id"`
}

type UpdateAlertTemplateRequest struct {
	Name             string                    `json:"name" validate:"required,name"`
	Description      string                    `json:"description"`
	OrganizationIds  []string                  `json:"organizationIds" validate:"required"`
	MetricQuery      string                    `json:"metricQuery" validate:"required"`
	MetricParameters []MetricParameterResponse `json:"metricParameters"`
}

type GetAlertTemplatesResponse struct {
	AlertTemplates []AlertTemplateResponse   `json:"alertTemplates"`
	Pagination     domain.PaginationResponse `json:"pagination"`
}

type GetAlertTemplateResponse struct {
	AlertTemplate AlertTemplateResponse `json:"alertTemplate"`
}
