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

// 내부
type CloudSetting struct {
	ID             string
	OrganizationId string
	Name           string
	Description    string
	Type           CloudType
	Resource       string
	Clusters       int
	Creator        string
	Updator        string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type CloudSettingResponse struct {
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

func (c *CloudSettingResponse) From(s CloudSetting) {
	c.ID = s.ID
	c.OrganizationId = s.OrganizationId
	c.Name = s.Name
	c.Description = s.Description
	c.Type = s.Type
	c.Resource = s.Resource
	c.Clusters = s.Clusters
	c.Creator = s.Creator
	c.Updator = s.Updator
	c.CreatedAt = s.CreatedAt
	c.UpdatedAt = s.UpdatedAt
}

func NewCloudSettingResponse(s CloudSetting) CloudSettingResponse {
	var r CloudSettingResponse
	r.From(s)
	return r
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
	Type           CloudType `json:"type" validate:"oneof=AWS AZZURE GCP"`
	SecretKeyId    string    `json:"secretKeyId" validate:"required"`
	SecretKey      string    `json:"secretKey" validate:"required"`
}

type CreateCloudSettingsResponse struct {
	CloudSettingId string `json:"cloudSettingId"`
}

type UpdateCloudSettingRequest struct {
	Description string `json:"description"`
	SecretKeyId string `json:"secretKeyId"`
	SecretKey   string `json:"secretKey"`
}
