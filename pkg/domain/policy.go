package domain

import (
	"encoding/json"
	"time"
)

type Kinds struct {
	APIGroups []string `json:"apiGroups,omitempty" protobuf:"bytes,1,rep,name=apiGroups"`
	Kinds     []string `json:"kinds,omitempty"`
}

type Match struct {
	Namespaces         []string `json:"namespaces,omitempty"`
	ExcludedNamespaces []string `json:"excludedNamespaces,omitempty"`
	Kinds              []Kinds  `json:"kinds,omitempty"`
}

func (m *Match) JSON() string {
	jsonBytes, err := json.Marshal(m)

	if err != nil {
		return ""
	}

	return string(jsonBytes)
}

type PolicyResponse struct {
	ID        string             `json:"id" example:"d98ef5f1-4a68-4047-a446-2207787ce3ff"`
	Creator   SimpleUserResponse `json:"creator,omitempty"`
	Updator   SimpleUserResponse `json:"updator,omitempty"`
	CreatedAt time.Time          `json:"createdAt" format:"date-time"`
	UpdatedAt time.Time          `json:"updatedAt" format:"date-time"`

	//	TargetClusterIds []string                `json:"targetClusterIds" example:"83bf8081-f0c5-4b31-826d-23f6f366ec90,83bf8081-f0c5-4b31-826d-23f6f366ec90"`
	TargetClusters []SimpleClusterResponse `json:"targetClusters"`
	Mandatory      bool                    `json:"mandatory"`

	PolicyName         string          `json:"policyName" example:"label 정책"`
	PolicyResourceName string          `json:"policyResourceName,omitempty" example:"labelpolicy"`
	Description        string          `json:"description"`
	TemplateId         string          `json:"templateId" example:"d98ef5f1-4a68-4047-a446-2207787ce3ff"`
	TemplateName       string          `json:"templateName" example:"필수 Label 검사"`
	EnforcementAction  string          `json:"enforcementAction" enum:"warn,deny,dryrun" example:"deny"`
	Parameters         string          `json:"parameters" example:"{\"key\":\"value\"}"`
	FilledParameters   []*ParameterDef `json:"filledParameters"`
	Match              *Match          `json:"match,omitempty"`
	MatchYaml          *string         `json:"matchYaml,omitempty" example:"namespaces:\r\n- testns1"`
	//Tags              []string         `json:"tags,omitempty" example:"k8s,label"`
}

type CreatePolicyRequest struct {
	TargetClusterIds []string `json:"targetClusterIds" example:"83bf8081-f0c5-4b31-826d-23f6f366ec90,83bf8081-f0c5-4b31-826d-23f6f366ec90"`
	Mandatory        bool     `json:"mandatory"`

	PolicyName         string  `json:"policyName" validate:"required,name" example:"label 정책"`
	PolicyResourceName string  `json:"policyResourceName,omitempty" validate:"resourcename" example:"labelpolicy"`
	Description        string  `json:"description"`
	TemplateId         string  `json:"templateId" example:"d98ef5f1-4a68-4047-a446-2207787ce3ff"`
	EnforcementAction  string  `json:"enforcementAction" validate:"required,oneof=deny dryrun warn" enum:"warn,deny,dryrun" example:"deny"`
	Parameters         string  `json:"parameters" example:"{\"key\":\"value\"}"`
	Match              *Match  `json:"match,omitempty"`
	MatchYaml          *string `json:"matchYaml,omitempty" example:"namespaces:\r\n- testns1"`
	//Tags              []string         `json:"tags,omitempty" example:"k8s,label"`
}

type CreatePolicyResponse struct {
	ID string `json:"id"`
}

type UpdatePolicyRequest struct {
	TargetClusterIds *[]string `json:"targetClusterIds,omitempty" example:"83bf8081-f0c5-4b31-826d-23f6f366ec90,83bf8081-f0c5-4b31-826d-23f6f366ec90"`
	Mandatory        *bool     `json:"mandatory,omitempty"`

	PolicyName        *string `json:"policyName,omitempty" validate:"required,name" example:"label 정책"`
	Description       *string `json:"description"`
	TemplateId        *string `json:"templateId,omitempty" example:"d98ef5f1-4a68-4047-a446-2207787ce3ff"`
	EnforcementAction *string `json:"enforcementAction" validate:"required,oneof=deny dryrun warn" enum:"warn,deny,dryrun" example:"deny"`
	Parameters        *string `json:"parameters" example:"{\"labels\":{\"key\":\"owner\",\"allowedRegex\":\"test*\"}"`
	Match             *Match  `json:"match,omitempty"`
	MatchYaml         *string `json:"matchYaml,omitempty"`
	//Tags              []string         `json:"tags,omitempty" example:"k8s,label"`
}

type UpdatePolicyClustersRequest struct {
	CurrentTargetClusterIds []string `json:"currentTargetClusterIds" example:"83bf8081-f0c5-4b31-826d-23f6f366ec90"`
	NewTargetClusterIds     []string `json:"newTargetClusterIds" example:"83bf8081-f0c5-4b31-826d-23f6f366ec90,83bf8081-f0c5-4b31-826d-23f6f366ec90"`
}

type GetPolicyResponse struct {
	Policy PolicyResponse `json:"policy"`
}

type ListPolicyResponse struct {
	Policies   []PolicyResponse   `json:"policies"`
	Pagination PaginationResponse `json:"pagination"`
}

type MandatoryPolicyInfo struct {
	PolicyName  string `json:"policyName" example:"org 레이블 요구"`
	PolicyId    string `json:"policyId" example:"0091fe9b-e44b-423d-9562-ac2b73089593"`
	Description string `json:"description" example:"org 레이블 설정 여부 검사"`
	Mandatory   bool   `json:"mandatory"`
}

type MandatoryTemplateInfo struct {
	TemplateName string                `json:"templateName" example:"레이블 요구"`
	TemplateId   string                `json:"templateId" example:"708d1e5b-4e6f-40e9-87a3-329e2fd051a5"`
	Description  string                `json:"description" example:"파라미터로 설정된 레이블 검사"`
	Mandatory    bool                  `json:"mandatory"`
	Policies     []MandatoryPolicyInfo `json:"policies"`
}

type GetMandatoryPoliciesResponse struct {
	Templates []MandatoryTemplateInfo `json:"templates"`
}

type MandatoryPolicyPatchInfo struct {
	PolicyId  string `json:"policyId" example:"0091fe9b-e44b-423d-9562-ac2b73089593"`
	Mandatory bool   `json:"mandatory"`
}

type SetMandatoryPoliciesRequest struct {
	Policies []MandatoryPolicyPatchInfo `json:"policies"`
}

type StackPolicyStatusResponse struct {
	PolicyName             string `json:"policyName" example:"org 레이블 요구"`
	PolicyId               string `json:"policyId" example:"0091fe9b-e44b-423d-9562-ac2b73089593"`
	PolicyDescription      string `json:"policyDescription" example:"org 레이블 설정 여부 검사"`
	PolicyMandatory        bool   `json:"policyMandatory"`
	TemplateName           string `json:"templateName" example:"레이블 요구"`
	TemplateId             string `json:"templateId" example:"708d1e5b-4e6f-40e9-87a3-329e2fd051a5"`
	TemplateDescription    string `json:"templateDescription" example:"파라미터로 설정된 레이블 검사"`
	TemplateCurrentVersion string `json:"templateCurrentVersion" example:"v1.0.1"`
	TemplateLatestVerson   string `json:"templateLatestVerson" example:"v1.0.3"`
}

type ListStackPolicyStatusResponse struct {
	Polices []StackPolicyStatusResponse `json:"polices"`
}

type GetStackPolicyTemplateStatusResponse struct {
	TemplateName                    string                           `json:"templateName" example:"레이블 요구"`
	TemplateId                      string                           `json:"templateId" example:"708d1e5b-4e6f-40e9-87a3-329e2fd051a5"`
	TemplateDescription             string                           `json:"templateDescription" example:"파라미터로 설정된 레이블 검사"`
	TemplateMandatory               bool                             `json:"templateMandatory"`
	TemplateCurrentVersion          string                           `json:"templateCurrentVersion" example:"v1.0.1"`
	TemplateLatestVerson            string                           `json:"templateLatestVerson" example:"v1.0.3"`
	TemplateLatestVersonReleaseDate time.Time                        `json:"templateLatestVersonReleaseDate" format:"date-time"`
	UpdatedPolicyParameters         []UpdatedPolicyTemplateParameter `json:"updatedPolicyParameters"`
	AffectedPolicies                []PolicyStatus                   `json:"affectedPolicies"`
}

type PolicyStatus struct {
	PolicyId         string            `json:"policyId" example:"0091fe9b-e44b-423d-9562-ac2b73089593"`
	PolicyName       string            `json:"policyName"`
	PolicyParameters []PolicyParameter `json:"policyPolicyParameters"`
}

type UpdatedPolicyTemplateParameter struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

type PolicyParameter struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Value     string `json:"value"`
	Updatable bool   `json:"updatable"`
}

type UpdatedPolicyParameters struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type PolicyUpdate struct {
	PolicyId                string                    `json:"policyId" example:"0091fe9b-e44b-423d-9562-ac2b73089593"`
	UpdatedPolicyParameters []UpdatedPolicyParameters `json:"updatedPolicyParameters"`
}

type UpdateStackPolicyTemplateStatusRequest struct {
	TemplateCurrentVersion string `json:"templateCurrentVersion" validate:"version" example:"v1.0.1"`
	TemplateTargetVerson   string `json:"templateTargetVerson" validate:"version" example:"v1.0.3"`
	// PolicyUpdate           []PolicyUpdate `json:"policyUpdate"`
}

type TemplateCount struct {
	TksTemplate          int64 `json:"tksTemplate"`
	OrganizationTemplate int64 `json:"organizationTemplate"`
	Total                int64 `json:"total"`
}

type PolicyCount struct {
	Deny            int64 `json:"deny"`
	Warn            int64 `json:"warn"`
	Dryrun          int64 `json:"dryrun"`
	FromTksTemplate int64 `json:"fromTksTemplate"`
	FromOrgTemplate int64 `json:"fromOrgTemplate"`
	Total           int64 `json:"total"`
}

type PolicyStatisticsResponse struct {
	Template TemplateCount `json:"templateCount"`
	Policy   PolicyCount   `json:"policyCount"`
}

type StackPolicyStatistics struct {
	TotalTemplateCount     int `json:"totalTemplateCount"`
	UptodateTemplateCount  int `json:"uptodateTemplateCount"`
	OutofdateTemplateCount int `json:"outofdateTemplateCount"`
	TotalPolicyCount       int `json:"totalPolicyCount"`
	UptodatePolicyCount    int `json:"uptodatePolicyCount"`
	OutofdatePolicyCount   int `json:"outofdatePolicyCount"`
}
