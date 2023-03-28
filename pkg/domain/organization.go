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

func (m OrganizationStatus) String() string { return organizationStatus[(m)] }
func (m OrganizationStatus) FromString(s string) OrganizationStatus {
	for i, v := range organizationStatus {
		if v == s {
			return OrganizationStatus(i)
		}
	}
	return OrganizationStatus_ERROR
}

type Organization = struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	Phone             string    `json:"phone"`
	Status            string    `json:"status"`
	StatusDescription string    `json:"statusDescription"`
	Creator           string    `json:"creator"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type CreateOrganizationRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Phone       string `json:"phone"`
}

func (r *CreateOrganizationRequest) ToOrganization() *Organization {
	return &Organization{
		ID:                "",
		Name:              r.Name,
		Description:       r.Description,
		Phone:             r.Phone,
		Status:            OrganizationStatus_CREATED.String(),
		StatusDescription: "",
		Creator:           "",
		CreatedAt:         time.Time{},
		UpdatedAt:         time.Time{},
	}
}

type UpdateOrganizationRequest struct {
	Description string `json:"description"`
	Phone       string `json:"phone"`
}

func (r *UpdateOrganizationRequest) ToOrganization() *Organization {
	return &Organization{
		ID:                "",
		Name:              "",
		Description:       r.Description,
		Phone:             r.Phone,
		Status:            "",
		StatusDescription: "",
		Creator:           "",
		CreatedAt:         time.Time{},
		UpdatedAt:         time.Time{},
	}
}
