package domain

import (
	"time"
)

type SystemNotificationTemplateResponse struct {
	ID               string                       `json:"id"`
	Name             string                       `json:"name"`
	Description      string                       `json:"description"`
	MetricQuery      string                       `json:"metricQuery" validate:"required"`
	MetricParameters []MetricParameterResponse    `json:"metricParameters,omitempty"`
	Organizations    []SimpleOrganizationResponse `json:"organizations,omitempty"`
	Creator          SimpleUserResponse           `json:"creator"`
	Updator          SimpleUserResponse           `json:"updator"`
	CreatedAt        time.Time                    `json:"createdAt"`
	UpdatedAt        time.Time                    `json:"updatedAt"`
}

type SimpleSystemNotificationTemplateResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type MetricParameterResponse struct {
	Order int    `json:"order" validate:"required"`
	Key   string `json:"key" validate:"required,name"`
	Value string `json:"value" validate:"required"`
}

type CreateSystemNotificationTemplateRequest struct {
	Name             string                    `json:"name" validate:"required,name"`
	Description      string                    `json:"description"`
	OrganizationIds  []string                  `json:"organizationIds" validate:"required"`
	MetricQuery      string                    `json:"metricQuery" validate:"required"`
	MetricParameters []MetricParameterResponse `json:"metricParameters"`
}

type CreateSystemNotificationTemplateResponse struct {
	ID string `json:"id"`
}

type UpdateSystemNotificationTemplateRequest struct {
	Name             string                    `json:"name" validate:"required,name"`
	Description      string                    `json:"description"`
	OrganizationIds  []string                  `json:"organizationIds" validate:"required"`
	MetricQuery      string                    `json:"metricQuery" validate:"required"`
	MetricParameters []MetricParameterResponse `json:"metricParameters"`
}

type GetSystemNotificationTemplatesResponse struct {
	SystemNotificationTemplates []SystemNotificationTemplateResponse `json:"systemNotificationTemplates"`
	Pagination                  PaginationResponse                   `json:"pagination"`
}

type GetSystemNotificationTemplateResponse struct {
	SystemNotificationTemplate SystemNotificationTemplateResponse `json:"systemNotificationTemplate"`
}
