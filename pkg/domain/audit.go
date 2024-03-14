package domain

import (
	"time"
)

type AuditResponse struct {
	ID             string                     `json:"id"`
	OrganizationId string                     `json:"organizationId"`
	Organization   SimpleOrganizationResponse `json:"organization"`
	Description    string                     `json:"description"`
	Group          string                     `json:"group"`
	Message        string                     `json:"message"`
	ClientIP       string                     `json:"clientIP"`
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
