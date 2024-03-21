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

	TargetClusterIds []string `json:"targetClusterIds" example:"83bf8081-f0c5-4b31-826d-23f6f366ec90,83bf8081-f0c5-4b31-826d-23f6f366ec90"`
	Mandatory        bool     `json:"mandatory"`

	PolicyName        string `json:"policyName" example:"label 정책"`
	Description       string `json:"description"`
	TemplateId        string `json:"templateId" example:"d98ef5f1-4a68-4047-a446-2207787ce3ff"`
	TemplateName      string `json:"templateName" example:"필수 Label 검사"`
	EnforcementAction string `json:"enforcementAction" enum:"warn,deny,dryrun"`
	Parameters        string `json:"parameters" example:"\"labels\":{\"key\":\"owner\",\"allowedRegex:^[a-zA-Z]+.agilebank.demo$}\""`
	Match             *Match `json:"match,omitempty" swaggertype:"object,string" example:"refer:match.Match"`
	//Tags              []string         `json:"tags,omitempty" example:"k8s,label"`
}

type CreatePolicyRequest struct {
	TargetClusterIds []string `json:"targetClusterIds" example:"83bf8081-f0c5-4b31-826d-23f6f366ec90,83bf8081-f0c5-4b31-826d-23f6f366ec90"`
	Mandatory        bool     `json:"mandatory"`

	PolicyName        string `json:"policyName" example:"label 정책"`
	Description       string `json:"description"`
	TemplateId        string `json:"templateId" example:"d98ef5f1-4a68-4047-a446-2207787ce3ff"`
	TemplateName      string `json:"templateName" example:"필수 Label 검사"`
	EnforcementAction string `json:"enforcementAction" enum:"warn,deny,dryrun"`
	Parameters        string `json:"parameters" example:"\"labels\":{\"key\":\"owner\",\"allowedRegex:^[a-zA-Z]+.agilebank.demo$}\""`
	Match             *Match `json:"match,omitempty" swaggertype:"object,string" example:"refer:match.Match"`
	//Tags              []string         `json:"tags,omitempty" example:"k8s,label"`
}

type CreatePolicyResponse struct {
	ID string `json:"id"`
}

type UpdatePolicyRequest struct {
	TargetClusterIds *[]string `json:"targetClusterIds,omitempty" example:"83bf8081-f0c5-4b31-826d-23f6f366ec90,83bf8081-f0c5-4b31-826d-23f6f366ec90"`
	Mandatory        *bool     `json:"mandatory,omitempty"`

	PolicyName        *string `json:"policyName,omitempty" example:"label 정책"`
	Description       string  `json:"description"`
	TemplateId        *string `json:"templateId,omitempty" example:"d98ef5f1-4a68-4047-a446-2207787ce3ff"`
	EnforcementAction *string `json:"enforcementAction,omitempty" enum:"warn,deny,dryrun"`
	Parameters        *string `json:"parameters,omitempty" example:"\"labels\":{\"key\":\"owner\",\"allowedRegex:^[a-zA-Z]+.agilebank.demo$}\""`
	Match             *Match  `json:"match,omitempty" swaggertype:"object,string" example:"refer:match.Match"`
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
