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

func (t *ProjectNamesapce) BeforeCreate(*gorm.DB) (err error) {
	t.ID = uuid.New().String()
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
	ProjectMembers    []ProjectMember    `gorm:"foreignKey:ProjectId" json:"projectMembers,omitempty"`
	ProjectNamesapces []ProjectNamesapce `gorm:"foreignKey:ProjectId" json:"projectNamesapces,omitempty"`
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
	ID          string    `json:"id"`
	AccountId   string    `json:"accountId"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Department  string    `json:"department"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
type ProjectMember struct {
	ID            string      `gorm:"primarykey" json:"id"`
	ProjectId     string      `gorm:"not null" json:"projectId"`
	UserId        string      `json:"userId"`
	User          ProjectUser `gorm:"-:all" json:"user"`
	ProjectRoleId string      `json:"projectRoleId"`
	ProjectRole   ProjectRole `gorm:"foreignKey:ProjectRoleId" json:"projectRole"`
	CreatedAt     time.Time   `gorm:"autoCreateTime:false" json:"createdAt"`
	UpdatedAt     *time.Time  `gorm:"autoUpdateTime:false" json:"updatedAt"`
	DeletedAt     *time.Time  `json:"deletedAt"`
}

type ProjectNamesapce struct {
	ID          string     `gorm:"primarykey" json:"id"`
	ProjectId   string     `gorm:"not null" json:"projectId"`
	StackID     string     `gorm:"uniqueIndex:idx_stackid_namespace" json:"stackId"`
	Namespace   string     `gorm:"uniqueIndex:idx_stackid_namespace" json:"namespace"`
	StackName   string     `gorm:"-:all" json:"stackName,omitempty"`
	Description string     `json:"description,omitempty"`
	Status      string     `json:"status,omitempty"`
	CreatedAt   time.Time  `gorm:"autoCreateTime:false" json:"createdAt"`
	UpdatedAt   *time.Time `gorm:"autoUpdateTime:false" json:"updatedAt"`
	DeletedAt   *time.Time `json:"deletedAt"`
}

type CreateProjectRequest struct {
	Name            string `json:"name" validate:"required"`
	Description     string `json:"description"`
	ProjectLeaderId string `json:"projectLeaderId"`
	ProjectRoleId   string `json:"projectRoleId"`
}

type CreateProjectResponse struct {
	Project Project `json:"project"`
}

type GetProjectRoleResponse struct {
	ProjectRole ProjectRole `json:"projectRole"`
}

type GetProjectRolesResponse struct {
	ProjectRoles []ProjectRole `json:"projectRoles"`
}

type ProjectMemberRequest struct {
	UserId        string `json:"userId" validate:"required"`
	ProjectRoleId string `json:"projectRoleId" validate:"required"`
}
type AddProjectMemberRequest struct {
	ProjectMemberRequests []ProjectMemberRequest `json:"projectMembers"`
}

type AddProjectMemberResponse struct {
	ProjectMembers []ProjectMember `json:"projectMembers"`
}

type GetProjectMemberResponse struct {
	ProjectMember ProjectMember `json:"projectMember"`
}

type GetProjectMembersResponse struct {
	ProjectMembers []ProjectMember `json:"projectMembers"`
}

type RemoveProjectMemberRequest struct {
	ProjectMember []struct {
		ProjectMemberId string `json:"projectMemberId"`
	} `json:"projectMembers"`
}

type CommonProjectResponse struct {
	Result string `json:"result"`
}

type UpdateProjectMemberRoleRequest struct {
	ProjectRoleId string `json:"projectRoleId"`
}
