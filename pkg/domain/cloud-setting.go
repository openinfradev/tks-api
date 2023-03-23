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
)

type CloudSetting = struct {
	ID             string    `json:"id"`
	OrganizationId string    `json:"organizationId"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Type           CloudType `json:"type"`
	Resource       string    `json:"resource"`
	Creator        string    `json:"creator"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type CreateCloudSettingRequest struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        CloudType `json:"type"`
	AccessKey   string    `json:"accessKey"`
	SecretKey   string    `json:"secretKey"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
