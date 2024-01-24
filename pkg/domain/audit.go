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
	Type           string
	Message        string
	ClientIP       string
	CreatorId      *uuid.UUID
	Creator        User
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
