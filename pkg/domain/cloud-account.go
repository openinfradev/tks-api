package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	CloudService_UNDEFINED = "UNDEFINED"
	CloudService_AWS       = "AWS"
	CloudService_AZURE     = "AZZURE"
	CloudService_GCP       = "GCP"
)

// 내부
type CloudAccount struct {
	ID             uuid.UUID
	OrganizationId string
	Name           string
	Description    string
	CloudService   string
	Resource       string
	Clusters       int
	CreatorId      uuid.UUID
	Creator        User
	UpdatorId      uuid.UUID
	Updator        User
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type CloudAccountResponse struct {
	ID             string             `json:"id"`
	OrganizationId string             `json:"organizationId"`
	Name           string             `json:"name"`
	Description    string             `json:"description"`
	CloudService   string             `json:"cloudService"`
	Resource       string             `json:"resource"`
	Clusters       int                `json:"clusters"`
	Creator        SimpleUserResponse `json:"creator"`
	Updator        SimpleUserResponse `json:"updator"`
	CreatedAt      time.Time          `json:"createdAt"`
	UpdatedAt      time.Time          `json:"updatedAt"`
}

type GetCloudAccountsResponse struct {
	CloudAccounts []CloudAccountResponse `json:"cloudAccounts"`
}

type GetCloudAccountResponse struct {
	CloudAccount CloudAccountResponse `json:"cloudAccount"`
}

type CreateCloudAccountRequest struct {
	Name            string `json:"name" validate:"required"`
	Description     string `json:"description"`
	CloudService    string `json:"cloudService" validate:"oneof=AWS AZZURE GCP"`
	AccessKeyId     string `json:"accessKeyId" validate:"required,min=16,max=128"`
	SecretAccessKey string `json:"secretAccessKey" validate:"required,min=16,max=128"`
	SessionToken    string `json:"sessionToken" validate:"max=2000"`
}

type CreateCloudAccountResponse struct {
	ID string `json:"id"`
}

type UpdateCloudAccountRequest struct {
	Description string `json:"description"`
}

type DeleteCloudAccountRequest struct {
	AccessKeyId     string `json:"accessKeyId" validate:"required,min=16,max=128"`
	SecretAccessKey string `json:"secretAccessKey" validate:"required,min=16,max=128"`
	SessionToken    string `json:"sessionToken" validate:"max=2000"`
}

type CheckCloudAccountNameResponse struct {
	Existed bool `json:"existed"`
}
