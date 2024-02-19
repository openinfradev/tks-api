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

type CreateAuditRequest struct {
}
type CreateAuditResponse struct {
}
type GetAuditsResponse struct {
}
type GetAuditResponse struct {
}
