package domain

import (
	"time"
)

// enum
type OrganizationStatus int32

const (
	OrganizationStatus_PENDING OrganizationStatus = iota
	OrganizationStatus_CREATE
	OrganizationStatus_CREATING
	OrganizationStatus_CREATED
	OrganizationStatus_DELETE
	OrganizationStatus_DELETING
	OrganizationStatus_DELETED
	OrganizationStatus_ERROR
)

var organizationStatus = [...]string{
	"PENDING",
	"CREATE",
	"CREATING",
	"CREATED",
	"DELETE",
	"DELETING",
	"DELETED",
	"ERROR",
}

var organizationStatusMap = map[string]OrganizationStatus{
	"PENDING":  OrganizationStatus_PENDING,
	"CREATE":   OrganizationStatus_CREATE,
	"CREATING": OrganizationStatus_CREATING,
	"CREATED":  OrganizationStatus_CREATED,
	"DELETE":   OrganizationStatus_DELETE,
	"DELETING": OrganizationStatus_DELETING,
	"DELETED":  OrganizationStatus_DELETED,
	"ERROR":    OrganizationStatus_ERROR,
}

func (m OrganizationStatus) String() string { return organizationStatus[(m)] }
func (m OrganizationStatus) FromString(s string) OrganizationStatus {
	if v, ok := organizationStatusMap[s]; ok {
		return v
	}
	return OrganizationStatus_ERROR
}

type OrganizationResponse struct {
	ID                          string                                     `json:"id"`
	Name                        string                                     `json:"name"`
	Description                 string                                     `json:"description"`
	PrimaryClusterId            string                                     `json:"primaryClusterId"`
	Status                      string                                     `json:"status"`
	StatusDesc                  string                                     `json:"statusDesc"`
	StackTemplates              []SimpleStackTemplateResponse              `json:"stackTemplates"`
	PolicyTemplates             []SimplePolicyTemplateResponse             `json:"policyTemplates"`
	SystemNotificationTemplates []SimpleSystemNotificationTemplateResponse `json:"systemNotificationTemplates"`
	Admin                       SimpleUserResponse                         `json:"admin"`
	ClusterCount                int                                        `json:"stackCount"`
	CreatedAt                   time.Time                                  `json:"createdAt"`
	UpdatedAt                   time.Time                                  `json:"updatedAt"`
}

type SimpleOrganizationResponse = struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreateOrganizationRequest struct {
	Name           string `json:"name" validate:"required,name"`
	Description    string `json:"description" validate:"omitempty,min=0,max=100"`
	AdminAccountId string `json:"adminAccountId" validate:"required"`
	AdminName      string `json:"adminName" validate:"name"`
	AdminEmail     string `json:"adminEmail" validate:"required"`
}

type CreateOrganizationResponse struct {
	ID string `json:"id"`
}

type GetOrganizationResponse struct {
	Organization OrganizationResponse `json:"organization"`
}
type ListOrganizationResponse struct {
	Organizations []OrganizationResponse `json:"organizations"`
	Pagination    PaginationResponse     `json:"pagination"`
}

type UpdateOrganizationRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=30"`
	Description string `json:"description" validate:"omitempty,min=0,max=100"`
}

type UpdateOrganizationResponse struct {
	ID string `json:"id"`
}

type UpdatePrimaryClusterRequest struct {
	PrimaryClusterId string `json:"primaryClusterId"`
}

type UpdateOrganizationTemplatesRequest struct {
	StackTemplateIds              *[]string `json:"stackTemplateIds,omitempty"`
	PolicyTemplateIds             *[]string `json:"policyTemplateIds,omitempty"`
	SystemNotificationTemplateIds *[]string `json:"systemNotificationTemplateIds,omitempty"`
}

type DeleteOrganizationResponse struct {
	ID string `json:"id"`
}
