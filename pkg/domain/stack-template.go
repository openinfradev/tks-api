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
	Template       string
	CloudService   string
	Version        string
	Platform       string
	KubeVersion    string
	KubeType       string
	Services       []byte
	CreatorId      uuid.UUID
	Creator        User
	UpdatorId      uuid.UUID
	Updator        User
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type StackTemplateServiceApplicationResponse struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

type StackTemplateServiceResponse struct {
	Type         string                                    `json:"type"`
	Applications []StackTemplateServiceApplicationResponse `json:"applications"`
}

type StackTemplateResponse struct {
	ID           string                         `json:"id"`
	Name         string                         `json:"name"`
	Description  string                         `json:"description"`
	Template     string                         `json:"template"`
	CloudService string                         `json:"cloudService"`
	Version      string                         `json:"version"`
	Platform     string                         `json:"platform"`
	KubeVersion  string                         `json:"kubeVersion"`
	KubeType     string                         `json:"kubeType"`
	Services     []StackTemplateServiceResponse `json:"services"`
	Creator      SimpleUserResponse             `json:"creator"`
	Updator      SimpleUserResponse             `json:"updator"`
	CreatedAt    time.Time                      `json:"createdAt"`
	UpdatedAt    time.Time                      `json:"updatedAt"`
}

type SimpleStackTemplateResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Template     string `json:"template"`
	CloudService string `json:"cloudService"`
}

type GetStackTemplatesResponse struct {
	StackTemplates []StackTemplateResponse `json:"stackTemplates"`
}

type GetStackTemplateResponse struct {
	StackTemplate StackTemplateResponse `json:"stackTemplate"`
}

type CreateStackTemplateRequest struct {
	Name         string `json:"name" validate:"required,name"`
	Description  string `json:"description"`
	CloudService string `json:"cloudService" validate:"oneof=AWS AZZURE GCP"`
	Version      string `json:"version" validate:"required"`
	Platform     string `json:"platform" validate:"required"`
	Template     string `json:"template" validate:"required"`
}

type CreateStackTemplateResponse struct {
	ID string `json:"id"`
}

type UpdateStackTemplateRequest struct {
	Description string `json:"description"`
}
