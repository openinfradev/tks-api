package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (a *Project) BeforeCreate(*gorm.DB) (err error) {
	a.ID = uuid.New().String()
	return nil
}

func (t *ProjectRole) BeforeCreate(*gorm.DB) (err error) {
	t.ID = uuid.New().String()
	return nil
}

func (t *ProjectMember) BeforeCreate(*gorm.DB) (err error) {
	t.ID = uuid.New().String()
	return nil
}

//func (t *ProjectNamespace) BeforeCreate(*gorm.DB) (err error) {
//	t.ID = uuid.New().String()
//	return nil
//}

func (t *ProjectNamespace) BeforeCreate(*gorm.DB) (err error) {
	return nil
}

type Project struct {
	ID                string             `gorm:"primarykey" json:"id"`
	OrganizationId    string             `json:"organizationId"`
	Name              string             `gorm:"index" json:"name"`
	Description       string             `json:"description,omitempty"`
	CreatedAt         time.Time          `gorm:"autoCreateTime:false" json:"createdAt"`
	UpdatedAt         *time.Time         `gorm:"autoUpdateTime:false" json:"updatedAt"`
	DeletedAt         *time.Time         `json:"deletedAt"`
	ProjectMembers    []ProjectMember    `gorm:"foreignKey:ProjectId;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT" json:"projectMembers,omitempty"`
	ProjectNamespaces []ProjectNamespace `gorm:"foreignKey:ProjectId;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT" json:"projectNamespaces,omitempty"`
}

type ProjectRole struct {
	ID          string     `gorm:"primarykey" json:"id"`
	Name        string     `json:"name"` // project-leader, project-member, project-viewer
	Description string     `json:"description,omitempty"`
	CreatedAt   time.Time  `gorm:"autoCreateTime:false" json:"createdAt" `
	UpdatedAt   *time.Time `gorm:"autoUpdateTime:false" json:"updatedAt"`
	DeletedAt   *time.Time `json:"deletedAt"`
}

type ProjectUser struct {
	ID          uuid.UUID `gorm:"primarykey;type:uuid" json:"id"`
	AccountId   string    `json:"accountId"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Department  string    `json:"department"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (ProjectUser) TableName() string {
	return "users"
}

type ProjectMember struct {
	ID              string       `gorm:"primarykey" json:"id"`
	ProjectId       string       `gorm:"not null" json:"projectId"`
	ProjectUserId   uuid.UUID    `json:"projectUserId"`
	ProjectUser     *ProjectUser `gorm:"foreignKey:ProjectUserId;references:ID;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT" json:"projectUser"`
	ProjectRoleId   string       `json:"projectRoleId"`
	ProjectRole     *ProjectRole `gorm:"foreignKey:ProjectRoleId;references:ID;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT" json:"projectRole"`
	IsProjectLeader bool         `gorm:"default:false" json:"projectLeader"`
	CreatedAt       time.Time    `gorm:"autoCreateTime:false" json:"createdAt"`
	UpdatedAt       *time.Time   `gorm:"autoUpdateTime:false" json:"updatedAt"`
	DeletedAt       *time.Time   `json:"deletedAt"`
}

type ProjectStack struct {
	ID             string `gorm:"primarykey" json:"id"`
	OrganizationId string `json:"organizationId"`
	Name           string `json:"name"`
}

func (ProjectStack) TableName() string {
	return "clusters"
}

type ProjectNamespace struct {
	StackId     string        `gorm:"primarykey" json:"stackId"`
	Namespace   string        `gorm:"primarykey" json:"namespace"`
	Stack       *ProjectStack `gorm:"foreignKey:StackId;references:ID;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT" json:"stack"`
	ProjectId   string        `gorm:"not null" json:"projectId"`
	Description string        `json:"description,omitempty"`
	Status      string        `json:"status,omitempty"`
	CreatedAt   time.Time     `gorm:"autoCreateTime:false" json:"createdAt"`
	UpdatedAt   *time.Time    `gorm:"autoUpdateTime:false" json:"updatedAt"`
	DeletedAt   *time.Time    `json:"deletedAt"`
}
