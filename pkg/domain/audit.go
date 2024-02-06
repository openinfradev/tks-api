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
	UserId         *uuid.UUID
	User           User
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
