package domain

import (
	"time"

	"github.com/google/uuid"
)

type PolicyTemplateResponse struct {
	ID        string             `json:"id" example:"d98ef5f1-4a68-4047-a446-2207787ce3ff"`
	Type      string             `json:"type" enums:"tks,organization" example:"tks"`
	Creator   SimpleUserResponse `json:"creator"`
	Updator   SimpleUserResponse `json:"updator"`
	CreatedAt time.Time          `json:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt"`

	TemplateName     string         `json:"templateName" example:"필수 Label 검사"`
	Kind             string         `json:"kind" example:"K8sRequiredLabels"`
	Severity         string         `json:"severity" enums:"low,medium,high" example:"medium"`
	Deprecated       bool           `json:"deprecated" example:"false"`
	Version          string         `json:"version,omitempty" example:"v1.0.1"`
	Description      string         `json:"description,omitempty"  example:"이 정책은 ..."`
	ParametersSchema []ParameterDef `json:"parametersSchema,omitempty"`
	Rego             string         `json:"rego" example:"rego 코드"`
	Libs             []string       `json:"libs" example:"rego 코드"`
}

type SimplePolicyTemplateResponse struct {
	ID          string `json:"id" example:"d98ef5f1-4a68-4047-a446-2207787ce3ff"`
	Type        string `json:"type" enums:"tks,organization" example:"tks"`
	Name        string `json:"templateName" example:"필수 Label 검사"`
	Version     string `json:"version,omitempty" example:"v1.0.1"`
	Description string `json:"description,omitempty"  example:"이 정책은 ..."`
}

type CreatePolicyTemplateRequest struct {
	TemplateName     string         `json:"templateName" example:"필수 Label 검사" validate:"name"`
	Kind             string         `json:"kind" example:"K8sRequiredLabels" validate:"required"`
	Severity         string         `json:"severity" enums:"low,medium,high" example:"medium"`
	Deprecated       bool           `json:"deprecated" example:"false"`
	Description      string         `json:"description,omitempty"  example:"이 정책은 ..."`
	ParametersSchema []ParameterDef `json:"parametersSchema,omitempty"`
	// "type: object\nproperties:  message:\n    type: string\n  labels:\n    type: array\n    items:\n      type: object\n      properties:\n        key:\n          type: string\n        allowedRegex:\n          type: string"

	Rego string   `json:"rego" example:"rego 코드" validate:"required"`
	Libs []string `json:"libs" example:"rego 코드"`

	PermittedOrganizationIds []string `json:"permittedOrganizationIds"`
}

type CreatePolicyTemplateReponse struct {
	ID string `json:"id" example:"d98ef5f1-4a68-4047-a446-2207787ce3ff"`
}

type UpdateCommmonPolicyTemplateRequest struct {
	TemplateName string `json:"templateName" example:"필수 Label 검사"`
	Description  string `json:"description,omitempty"`
	Severity     string `json:"severity" enums:"low,medium,high" example:"medium"`
	Deprecated   bool   `json:"deprecated" example:"false"`
	// Tags         []string `json:"tags,omitempty"`
}

type UpdatePolicyTemplateUpdate struct {
	ID                       uuid.UUID
	Type                     string
	UpdatorId                uuid.UUID
	TemplateName             *string
	Description              *string
	Severity                 *string
	Deprecated               *bool
	PermittedOrganizationIds *[]string
}

func (dto *UpdatePolicyTemplateUpdate) IsNothingToUpdate() bool {
	return dto.TemplateName == nil &&
		dto.Description == nil &&
		dto.Severity == nil &&
		dto.Deprecated == nil &&
		dto.PermittedOrganizationIds == nil
}

type UpdatePolicyTemplateRequest struct {
	TemplateName             *string   `json:"templateName,omitempty" example:"필수 Label 검사"`
	Description              *string   `json:"description,omitempty"`
	Severity                 *string   `json:"severity,omitempty" enums:"low,medium,high" example:"medium"`
	Deprecated               *bool     `json:"deprecated,omitempty" example:"false"`
	PermittedOrganizationIds *[]string `json:"permittedOrganizationIds,omitempty"`
}

type GetPolicyTemplateDeployResponse struct {
	DeployVersion map[string]string `json:"deployVersion"`
}

type ListPolicyTemplateVersionsResponse struct {
	Versions []string `json:"versions" example:"v1.1.0,v1.0.1,v1.0.0"`
}

type GetPolicyTemplateVersionResponse struct {
	PolicyTemplate PolicyTemplateResponse `json:"policyTemplate"`
}

type CreatePolicyTemplateVersionRequest struct {
	VersionUpType   string `json:"versionUpType" enums:"major,minor,patch" example:"minor" validate:"required"`
	CurrentVersion  string `json:"currentVersion" example:"v1.0.0" validate:"required"`
	ExpectedVersion string `json:"expectedVersion" example:"v1.1.0" validate:"required"`

	ParametersSchema []ParameterDef `json:"parametersSchema,omitempty"`
	// "type: object\nproperties:  message:\n    type: string\n  labels:\n    type: array\n    items:\n      type: object\n      properties:\n        key:\n          type: string\n        allowedRegex:\n          type: string"

	Rego string   `json:"rego" example:"rego 코드" validate:"required"`
	Libs []string `json:"libs" example:"rego 코드"`
}

type CreatePolicyTemplateVersionResponse struct {
	Version string `json:"version" example:"v1.1.1"`
}

type GetPolicyTemplateResponse struct {
	PolicyTemplate PolicyTemplateResponse `json:"policyTemplate"`
}

type ListPolicyTemplateResponse struct {
	PolicyTemplates []PolicyTemplateResponse `json:"policyTemplates"`
	Pagination      PaginationResponse       `json:"pagination"`
}

type PolicyTemplateStatistics struct {
	OrganizationId   string `json:"organizationId"`
	OrganizationName string `json:"organizationName"`
	UsageCount       int    `json:"usageCount"`
}

type ListPolicyTemplateStatisticsResponse struct {
	PolicyTemplateStatistics []PolicyTemplateStatistics `json:"policyTemplateStatistics"`
}

type ExistsPolicyTemplateNameResponse struct {
	Existed bool `json:"existed"`
}

type ExistsPolicyTemplateKindResponse struct {
	Existed bool `json:"existed"`
}

type ParameterDef struct {
	Key          string          `json:"key" exmaples:"repos"`
	Type         string          `json:"type" examples:"string[]"`
	DefaultValue string          `json:"defaultValue"`
	Children     []*ParameterDef `json:"children"`
	IsArray      bool            `json:"isArray"  examples:"true"`
}

type RegoCompileRequest struct {
	Rego string   `json:"rego" example:"Rego 코드" validate:"required"`
	Libs []string `json:"libs,omitempty"`
}

type RegoCompieError struct {
	Status  int    `json:"status" example:"400"`
	Code    string `json:"code" example:"P_INVALID_REGO_SYNTAX"`
	Message string `json:"message" example:"Invalid rego syntax"`
	Text    string `json:"text" example:"Rego 문법 에러입니다. 라인:2 컬럼:1 에러메시지: var testnum is not safe"`
}

type RegoCompileResponse struct {
	ParametersSchema []*ParameterDef   `json:"parametersSchema,omitempty"`
	Errors           []RegoCompieError `json:"errors,omitempty"`
}
