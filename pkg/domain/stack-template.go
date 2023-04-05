package domain

import (
	"time"

	"github.com/google/uuid"
)

// 내부
type StackTemplate struct {
	ID             uuid.UUID
	OrganizationId string
	Name           string
	Description    string
	Version        string
	CloudService   string
	Platform       string
	Template       string
	CreatorId      uuid.UUID
	Creator        User
	UpdatorId      uuid.UUID
	Updator        User
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type StackTemplateResponse struct {
	ID             string             `json:"id"`
	OrganizationId string             `json:"organizationId"`
	Name           string             `json:"name"`
	Description    string             `json:"description"`
	CloudService   string             `json:"cloudService"`
	Version        string             `json:"version"`
	Platform       string             `json:"platform"`
	Template       string             `json:"template"`
	Creator        SimpleUserResponse `json:"creator"`
	Updator        SimpleUserResponse `json:"updator"`
	CreatedAt      time.Time          `json:"createdAt"`
	UpdatedAt      time.Time          `json:"updatedAt"`
}

type GetStackTemplatesResponse struct {
	StackTemplates []StackTemplateResponse `json:"stackTemplates"`
}

type GetStackTemplateResponse struct {
	StackTemplate StackTemplateResponse `json:"stackTemplate"`
}

type CreateStackTemplateRequest struct {
	OrganizationId string `json:"organizationId" validate:"required"`
	Name           string `json:"name" validate:"required"`
	Description    string `json:"description"`
	CloudService   string `json:"cloudService" validate:"oneof=AWS AZZURE GCP"`
	Version        string `json:"version" validate:"required"`
	Platform       string `json:"platform" validate:"required"`
	Template       string `json:"template" validate:"required"`
}

type CreateStackTemplateResponse struct {
	ID string `json:"id"`
}

type UpdateStackTemplateRequest struct {
	Description string `json:"description"`
}
