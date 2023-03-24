package domain

import (
	"time"
)

// enum
type CloudType string

const (
	CloudType_UNDEFINED = "UNDEFINED"
	CloudType_AWS       = "AWS"
	CloudType_AZURE     = "AZZURE"
	CloudType_GCP       = "GCP"
)

type CloudSetting = struct {
	ID             string    `json:"id"`
	OrganizationId string    `json:"organizationId"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Type           CloudType `json:"type"`
	Resource       string    `json:"resource"`
	Clusters       int       `json:"clusters"`
	Creator        string    `json:"creator"`
	Updator        string    `json:"updator"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type CreateCloudSettingRequest struct {
	Name        string    `json:"name" validate:"required"`
	Description string    `json:"description"`
	Type        CloudType `json:"type" validate:"oneof=AWS AZZURE GCP"`
	SecretKeyId string    `json:"secretKeyId" validate:"required"`
	SecretKey   string    `json:"secretKey" validate:"required"`
}

type UpdateCloudSettingRequest struct {
	Description string `json:"description"`
	SecretKeyId string `json:"secretKeyId"`
	SecretKey   string `json:"secretKey"`
}
