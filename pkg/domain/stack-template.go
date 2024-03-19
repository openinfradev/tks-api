package domain

import (
	"time"
)

const STACK_TEMPLATE_TYPE_STANDARD = "STANDARD"
const STACK_TEMPLATE_TYPE_MSA = "MSA"

type StackTemplateServiceApplicationResponse struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

type StackTemplateServiceResponse struct {
	Type         string                                    `json:"type"`
	Applications []StackTemplateServiceApplicationResponse `json:"applications"`
}

type StackTemplateResponse struct {
	ID            string                         `json:"id"`
	Name          string                         `json:"name"`
	Description   string                         `json:"description"`
	Template      string                         `json:"template"`
	TemplateType  string                         `json:"templateType"`
	CloudService  string                         `json:"cloudService"`
	Version       string                         `json:"version"`
	Platform      string                         `json:"platform"`
	KubeVersion   string                         `json:"kubeVersion"`
	KubeType      string                         `json:"kubeType"`
	Organizations []SimpleOrganizationResponse   `json:"organizations"`
	Services      []StackTemplateServiceResponse `json:"services"`
	Creator       SimpleUserResponse             `json:"creator"`
	Updator       SimpleUserResponse             `json:"updator"`
	CreatedAt     time.Time                      `json:"createdAt"`
	UpdatedAt     time.Time                      `json:"updatedAt"`
}

type SimpleStackTemplateServiceResponse struct {
	Type string `json:"type"`
}

type SimpleStackTemplateResponse struct {
	ID           string                               `json:"id"`
	Name         string                               `json:"name"`
	Description  string                               `json:"description"`
	Template     string                               `json:"template"`
	CloudService string                               `json:"cloudService"`
	KubeVersion  string                               `json:"kubeVersion"`
	KubeType     string                               `json:"kubeType"`
	Services     []SimpleStackTemplateServiceResponse `json:"services"`
}

type GetStackTemplatesResponse struct {
	StackTemplates []StackTemplateResponse `json:"stackTemplates"`
	Pagination     PaginationResponse      `json:"pagination"`
}

type GetStackTemplateResponse struct {
	StackTemplate StackTemplateResponse `json:"stackTemplate"`
}

type CreateStackTemplateRequest struct {
	Name        string `json:"name" validate:"required,name"`
	Description string `json:"description"`
	Version     string `json:"version" validate:"required"`

	CloudService string `json:"cloudService" validate:"oneof=AWS AZZURE GCP"`
	Platform     string `json:"platform" validate:"required"`
	TemplateType string `json:"templateType" validate:"oneof=STANDARD MSA"`
	Template     string `json:"template" validate:"required"`
	KubeVersion  string `json:"kubeVersion" validate:"required"`
	KubeType     string `json:"kubeType" validate:"required"`

	OrganizationIds []string `json:"organizationIds" validate:"required"`
	ServiceIds      []string `json:"serviceIds" validate:"required"`
}

type CreateStackTemplateResponse struct {
	ID string `json:"id"`
}

type UpdateStackTemplateRequest struct {
	Description  string   `json:"description"`
	Template     string   `json:"template"`
	TemplateType string   `json:"templateType"`
	CloudService string   `json:"cloudService"`
	Version      string   `json:"version"`
	Platform     string   `json:"platform"`
	KubeVersion  string   `json:"kubeVersion"`
	KubeType     string   `json:"kubeType"`
	ServiceIds   []string `json:"serviceIds" validate:"required"`
}

type GetStackTemplateServicesResponse struct {
	Services []StackTemplateServiceResponse `json:"services"`
}

type UpdateStackTemplateOrganizationsRequest struct {
	OrganizationIds []string `json:"organizationIds" validate:"required"`
}

type CheckStackTemplateNameResponse struct {
	Existed bool `json:"existed"`
}
