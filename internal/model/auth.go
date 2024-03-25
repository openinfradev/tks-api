package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

// Models
type ExpiredTokenTime struct {
	gorm.Model

	OrganizationId string `gorm:"index:idx_org_id_subject_id,unique;not null"`
	SubjectId      string `gorm:"index:idx_org_id_subject_id,unique;not null"`
	ExpiredTime    time.Time
}

type CacheEmailCode struct {
	gorm.Model

	UserId uuid.UUID `gorm:"not null"`
	Code   string    `gorm:"type:varchar(6);not null"`
}
