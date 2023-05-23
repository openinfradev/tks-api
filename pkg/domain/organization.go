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

type Organization = struct {
	ID               string             `json:"id"`
	Name             string             `json:"name"`
	Description      string             `json:"description"`
	Phone            string             `json:"phone"`
	PrimaryClusterId string             `json:"primaryClusterId"`
	Status           OrganizationStatus `json:"status"`
	StatusDesc       string             `json:"statusDesc"`
	Creator          string             `json:"creator"`
	CreatedAt        time.Time          `json:"createdAt"`
	UpdatedAt        time.Time          `json:"updatedAt"`
}

type CreateOrganizationRequest struct {
	Name        string `json:"name" validate:"required,name"`
	Description string `json:"description" validate:"omitempty,min=0,max=100"`
	Phone       string `json:"phone"`
	Email       string `json:"Email" validate:"required,email"`
}

type CreateOrganizationResponse struct {
	ID string `json:"id"`
}

type GetOrganizationResponse struct {
	Organization struct {
		ID               string    `json:"id"`
		Name             string    `json:"name"`
		Description      string    `json:"description"`
		Phone            string    `json:"phone"`
		PrimaryClusterId string    `json:"primaryClusterId"`
		Status           string    `json:"status"`
		StatusDesc       string    `json:"statusDesc"`
		Creator          string    `json:"creator"`
		CreatedAt        time.Time `json:"createdAt"`
		UpdatedAt        time.Time `json:"updatedAt"`
	} `json:"organization"`
}
type ListOrganizationResponse struct {
	Organizations []ListOrganizationBody `json:"organizations"`
}
type ListOrganizationBody struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	Phone            string    `json:"phone"`
	PrimaryClusterId string    `json:"primaryClusterId"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type UpdateOrganizationRequest struct {
	PrimaryClusterId string `json:"primaryClusterId"`
	Name             string `json:"name" validate:"required,min=1,max=30"`
	Description      string `json:"description" validate:"omitempty,min=0,max=100"`
	Phone            string `json:"phone"`
}

type UpdateOrganizationResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Phone       string `json:"phone"`
}

type UpdatePrimaryClusterRequest struct {
	PrimaryClusterId string `json:"primaryClusterId"`
}
