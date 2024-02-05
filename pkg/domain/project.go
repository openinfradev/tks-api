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

func (t *ProjectNamespace) BeforeCreate(*gorm.DB) (err error) {
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
	ProjectNamespaces []ProjectNamespace `gorm:"foreignKey:ProjectId" json:"projectNamespaces,omitempty"`
}

//type ProjectList struct {
//	ID             string `json:"id"`
//	OrganizationId string `json:"organizationId"`
//	Name           string `json:"name"`
//	Description    string `json:"description"`
//	ProjectMembers []struct {
//		ID              string     `json:"id"`
//		UserId          string     `json:"userId"`
//		AccountId       string     `json:"accountId"`
//		Name            string     `json:"name"`
//		Email           string     `json:"email"`
//		ProjectRoleId   string     `json:"projectId"`
//		ProjectRoleName string     `json:"projectRoleName"`
//		CreatedAt       time.Time  `json:"createdAt"`
//		UpdatedAt       *time.Time `json:"updatedAt"`
//	} `json:"projectMembers"`
//	ProjectNamespaces []struct {
//		ID          string     `json:"id"`
//		StackId     string     `json:"stackId"`
//		StackName   string     `json:"stackName"`
//		Namespace   string     `json:"namespace"`
//		Description string     `json:"description"`
//		Status      string     `json:"status"`
//		CreatedAt   time.Time  `json:"createdAt"`
//		UpdatedAt   *time.Time `json:"updatedAt"`
//	} `json:"projectNamespaces"`
//	CreatedAt time.Time  `json:"createdAt"`
//	UpdatedAt *time.Time `json:"updatedAt"`
//}

type GetProjectsResponse struct {
	Projects []Project `json:"projects"`
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
	ID            string      `gorm:"primarykey" json:"id"`
	ProjectId     string      `gorm:"not null" json:"projectId"`
	ProjectUserId uuid.UUID   `json:"projectUserId"`
	ProjectUser   ProjectUser `gorm:"foreignKey:ProjectUserId;references:ID" json:"projectUser"`
	ProjectRoleId string      `json:"projectRoleId"`
	ProjectRole   ProjectRole `gorm:"foreignKey:ProjectRoleId" json:"projectRole"`
	CreatedAt     time.Time   `gorm:"autoCreateTime:false" json:"createdAt"`
	UpdatedAt     *time.Time  `gorm:"autoUpdateTime:false" json:"updatedAt"`
	DeletedAt     *time.Time  `json:"deletedAt"`
}

type ProjectNamespace struct {
	ID          string     `gorm:"primarykey" json:"id"`
	ProjectId   string     `gorm:"not null" json:"projectId"`
	StackId     string     `gorm:"uniqueIndex:idx_stackid_namespace" json:"stackId"`
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
	ProjectUserId string `json:"projectUserId" validate:"required"`
	ProjectRoleId string `json:"projectRoleId" validate:"required"`
}
type AddProjectMemberRequest struct {
	ProjectMemberRequests []ProjectMemberRequest `json:"projectMembers"`
}

type AddProjectMemberResponse struct {
	ProjectMembers []ProjectMember `json:"projectMembers"`
}

type GetProjectMemberResponse struct {
	ProjectMember *ProjectMember `json:"projectMember"`
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

type CreateProjectNamespaceRequest struct {
	Namespace   string `json:"namespace"`
	Description string `json:"description"`
}

type CreateProjectNamespaceResponse struct {
	ProjectNamesapceId string `json:"projectNamespaceId"`
}

type GetProjectNamespacesResponse struct {
	ProjectNamespaces []ProjectNamespace `json:"projectNamespaces"`
}

type GetProjectNamespaceResponse struct {
	ProjectNamespace *ProjectNamespace `json:"projectNamespace"`
}
