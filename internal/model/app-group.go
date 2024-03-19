package model

import (
	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/pkg/domain"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Models
type AppGroup struct {
	gorm.Model

	ID           domain.AppGroupId `gorm:"primarykey"`
	AppGroupType domain.AppGroupType
	ClusterId    domain.ClusterId
	Name         string
	Description  string
	WorkflowId   string
	Status       domain.AppGroupStatus
	StatusDesc   string
	CreatorId    *uuid.UUID `gorm:"type:uuid"`
	Creator      User       `gorm:"foreignKey:CreatorId"`
	UpdatorId    *uuid.UUID `gorm:"type:uuid"`
	Updator      User       `gorm:"foreignKey:UpdatorId"`
}

type Application struct {
	gorm.Model

	ID         uuid.UUID `gorm:"primarykey;type:uuid"`
	AppGroupId domain.AppGroupId
	Endpoint   string
	Metadata   datatypes.JSON
	Type       domain.ApplicationType
}
