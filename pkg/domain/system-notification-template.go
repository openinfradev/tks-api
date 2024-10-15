package domain

import (
	"time"
)

const SN_TYPE_NODE_CPU_HIGH_LOAD = "node-cpu-high-load"
const SN_TYPE_NODE_MEMORY_HIGH_UTILIZATION = "node-memory-high-utilization"
const SN_TYPE_NODE_DISK_FULL = "node-disk-full"
const SN_TYPE_PVC_FULL = "pvc-full"
const SN_TYPE_POD_RESTART_FREQUENTLY = "pod-restart-frequently"
const SN_TYPE_POLICY_WARNING = "policy-warning"
const SN_TYPE_POLICY_BLOCKED = "policy-blocked"

const (
	NT_SYSTEM_NOTIFICATION = "SYSTEM_NOTIFICATION"
	NT_POLICY              = "POLICY"
)

type SystemNotificationTemplateResponse struct {
	ID               string                                      `json:"id"`
	Name             string                                      `json:"name"`
	Description      string                                      `json:"description"`
	NotificationType string                                      `json:"notificationType"`
	MetricQuery      string                                      `json:"metricQuery" validate:"required"`
	MetricParameters []SystemNotificationMetricParameterResponse `json:"metricParameters,omitempty"`
	Organizations    []SimpleOrganizationResponse                `json:"organizations,omitempty"`
	IsSystem         bool                                        `json:"isSystem"`
	Creator          SimpleUserResponse                          `json:"creator"`
	Updator          SimpleUserResponse                          `json:"updator"`
	CreatedAt        time.Time                                   `json:"createdAt"`
	UpdatedAt        time.Time                                   `json:"updatedAt"`
}

type SimpleSystemNotificationTemplateResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type SystemNotificationMetricParameterResponse struct {
	Order int    `json:"order" validate:"required"`
	Key   string `json:"key" validate:"required,name"`
	Value string `json:"value" validate:"required"`
}

type CreateSystemNotificationTemplateRequest struct {
	Name             string                                      `json:"name" validate:"required,name"`
	Description      string                                      `json:"description"`
	OrganizationIds  []string                                    `json:"organizationIds" validate:"required"`
	MetricQuery      string                                      `json:"metricQuery" validate:"required"`
	MetricParameters []SystemNotificationMetricParameterResponse `json:"metricParameters"`
}

type CreateSystemNotificationTemplateResponse struct {
	ID string `json:"id"`
}

type UpdateSystemNotificationTemplateRequest struct {
	Name             string                                      `json:"name" validate:"required,name"`
	Description      string                                      `json:"description"`
	OrganizationIds  []string                                    `json:"organizationIds" validate:"required"`
	MetricQuery      string                                      `json:"metricQuery" validate:"required"`
	MetricParameters []SystemNotificationMetricParameterResponse `json:"metricParameters"`
}

type GetSystemNotificationTemplatesResponse struct {
	SystemNotificationTemplates []SystemNotificationTemplateResponse `json:"systemNotificationTemplates"`
	Pagination                  PaginationResponse                   `json:"pagination"`
}

type GetSystemNotificationTemplateResponse struct {
	SystemNotificationTemplate SystemNotificationTemplateResponse `json:"systemNotificationTemplate"`
}

type AddOrganizationSystemNotificationTemplatesRequest struct {
	SystemNotificationTemplateIds []string `json:"systemNotificationTemplateIds" validate:"required"`
}

type RemoveOrganizationSystemNotificationTemplatesRequest struct {
	SystemNotificationTemplateIds []string `json:"systemNotificationTemplateIds" validate:"required"`
}

type CheckSystemNotificaionTemplateNameResponse struct {
	Existed bool `json:"existed"`
}
