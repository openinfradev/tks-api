package domain

import (
	"time"

	"github.com/google/uuid"
)

// enum
type CloudType string

const (
	CloudType_UNDEFINED = "UNDEFINED"
	CloudType_AWS       = "AWS"
	CloudType_AZURE     = "AZZURE"
	CloudType_GCP       = "GCP"
)

// 내부
type CloudSetting struct {
	ID             uuid.UUID
	OrganizationId string
	Name           string
	Description    string
	Type           CloudType
	Resource       string
	Clusters       int
	CreatorId      uuid.UUID
	Creator        User
	UpdatorId      uuid.UUID
	Updator        User
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type CloudSettingResponse struct {
	ID             string             `json:"id"`
	OrganizationId string             `json:"organizationId"`
	Name           string             `json:"name"`
	Description    string             `json:"description"`
	Type           CloudType          `json:"cloudService"`
	Resource       string             `json:"resource"`
	Clusters       int                `json:"clusters"`
	Creator        SimpleUserResponse `json:"creator"`
	Updator        SimpleUserResponse `json:"updator"`
	CreatedAt      time.Time          `json:"createdAt"`
	UpdatedAt      time.Time          `json:"updatedAt"`
}

type GetCloudSettingsResponse struct {
	CloudSettings []CloudSettingResponse `json:"cloudSettings"`
}

type GetCloudSettingResponse struct {
	CloudSetting CloudSettingResponse `json:"cloudSetting"`
}

type CreateCloudSettingRequest struct {
	OrganizationId string    `json:"organizationId"`
	Name           string    `json:"name" validate:"required"`
	Description    string    `json:"description"`
	Type           CloudType `json:"cloudService" validate:"oneof=AWS AZZURE GCP"`
	SecretKeyId    string    `json:"secretKeyId" validate:"required"`
	SecretKey      string    `json:"secretKey" validate:"required"`
}

type CreateCloudSettingsResponse struct {
	ID string `json:"id"`
}

type UpdateCloudSettingRequest struct {
	Description string `json:"description"`
}

type DeleteCloudSettingRequest struct {
	SecretKeyId string `json:"secretKeyId" validate:"required"`
	SecretKey   string `json:"secretKey" validate:"required"`
}
