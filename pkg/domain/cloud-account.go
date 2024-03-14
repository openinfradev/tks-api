package domain

import (
	"time"

	"github.com/google/uuid"
)

const CLOUD_ACCOUNT_INCLUSTER = "INCLUSTER"

const (
	CloudService_UNDEFINED = "UNDEFINED"
	CloudService_AWS       = "AWS"
	CloudService_AZURE     = "AZZURE"
	CloudService_GCP       = "GCP"
	CloudService_BYOH      = "BYOH"
)

// enum
type CloudAccountStatus int32

const (
	CloudAccountStatus_PENDING CloudAccountStatus = iota
	CloudAccountStatus_CREATING
	CloudAccountStatus_CREATED
	CloudAccountStatus_DELETING
	CloudAccountStatus_DELETED
	CloudAccountStatus_CREATE_ERROR
	CloudAccountStatus_DELETE_ERROR
)

var cloudAccountStatus = [...]string{
	"PENDING",
	"CREATING",
	"CREATED",
	"DELETING",
	"DELETED",
	"CREATE_ERROR",
	"DELETE_ERROR",
}

func (m CloudAccountStatus) String() string { return cloudAccountStatus[(m)] }
func (m CloudAccountStatus) FromString(s string) CloudAccountStatus {
	for i, v := range cloudAccountStatus {
		if v == s {
			return CloudAccountStatus(i)
		}
	}
	return CloudAccountStatus_PENDING
}

// 내부
type CloudAccount struct {
	ID              uuid.UUID
	OrganizationId  string
	Name            string
	Description     string
	CloudService    string
	Resource        string
	Clusters        int
	AwsAccountId    string
	AccessKeyId     string
	SecretAccessKey string
	SessionToken    string
	Status          CloudAccountStatus
	StatusDesc      string
	CreatedIAM      bool
	CreatorId       uuid.UUID
	Creator         User
	UpdatorId       uuid.UUID
	Updator         User
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type ResourceQuotaAttr struct {
	Type     string `json:"type"`
	Usage    int    `json:"usage"`
	Quota    int    `json:"quota"`
	Required int    `json:"required"`
}

type ResourceQuota struct {
	Quotas []ResourceQuotaAttr `json:"quotas"`
}

type CloudAccountResponse struct {
	ID             string             `json:"id"`
	OrganizationId string             `json:"organizationId"`
	Name           string             `json:"name"`
	Description    string             `json:"description"`
	CloudService   string             `json:"cloudService"`
	Resource       string             `json:"resource"`
	Clusters       int                `json:"clusters"`
	Status         string             `json:"status"`
	AwsAccountId   string             `json:"awsAccountId"`
	CreatedIAM     bool               `json:"createdIAM"`
	Creator        SimpleUserResponse `json:"creator"`
	Updator        SimpleUserResponse `json:"updator"`
	CreatedAt      time.Time          `json:"createdAt"`
	UpdatedAt      time.Time          `json:"updatedAt"`
}

type SimpleCloudAccountResponse struct {
	ID             string `json:"id"`
	OrganizationId string `json:"organizationId"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	CloudService   string `json:"cloudService"`
	AwsAccountId   string `json:"awsAccountId"`
	CreatedIAM     bool   `json:"createdIAM"`
	Clusters       int    `json:"clusters"`
}

type GetCloudAccountsResponse struct {
	CloudAccounts []CloudAccountResponse `json:"cloudAccounts"`
	Pagination    PaginationResponse     `json:"pagination"`
}

type GetCloudAccountResponse struct {
	CloudAccount CloudAccountResponse `json:"cloudAccount"`
}

type CreateCloudAccountRequest struct {
	Name            string `json:"name" validate:"required,name"`
	Description     string `json:"description"`
	CloudService    string `json:"cloudService" validate:"oneof=AWS AZZURE GCP"`
	AwsAccountId    string `json:"awsAccountId" validate:"required,min=12,max=12"`
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

type CheckCloudAccountAwsAccountIdResponse struct {
	Existed bool `json:"existed"`
}

type GetCloudAccountResourceQuotaResponse struct {
	Available     bool          `json:"available"`
	ResourceQuota ResourceQuota `json:"resourceQuota"`
}
