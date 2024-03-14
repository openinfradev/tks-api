package domain

import (
	"time"

	"github.com/google/uuid"
)

// 내부
type Audit struct {
	ID             uuid.UUID
	OrganizationId string
	Organization   Organization
	Group          string
	Message        string
	Description    string
	ClientIP       string
	UserId         *uuid.UUID
	User           User
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type AuditResponse struct {
	ID             string                     `json:"id"`
	OrganizationId string                     `json:"organizationId"`
	Organization   SimpleOrganizationResponse `json:"organization"`
	Description    string                     `json:"description"`
	Group          string                     `json:"group"`
	Message        string                     `json:"message"`
	ClientIP       string                     `json:"clientIP"`
	UserId         string                     `json:"userId"`
	User           SimpleUserResponse         `json:"user"`
	CreatedAt      time.Time                  `json:"createdAt"`
	UpdatedAt      time.Time                  `json:"updatedAt"`
}

type CreateAuditRequest struct {
}
type CreateAuditResponse struct {
}

type GetAuditResponse struct {
	Audit AuditResponse `json:"audit"`
}
type GetAuditsResponse struct {
	Audits     []AuditResponse    `json:"audits"`
	Pagination PaginationResponse `json:"pagination"`
}
