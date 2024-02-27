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

type ProjectResponse struct {
	ID              string    `json:"id"`
	OrganizationId  string    `json:"organizationId"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	IsMyProject     string    `json:"isMyProject"`
	ProjectRoleId   string    `json:"projectRoleId"`
	ProjectRoleName string    `json:"projectRoleName"`
	NamespaceCount  int       `json:"namespaceCount"`
	AppCount        int       `json:"appCount"`
	MemberCount     int       `json:"memberCount"`
	CreatedAt       time.Time `json:"createdAt"`
}

type GetProjectsResponse struct {
	Projects   []ProjectResponse  `json:"projects"`
	Pagination PaginationResponse `json:"pagination"`
}

type ProjectDetailResponse struct {
	ID                      string `json:"id"`
	OrganizationId          string `json:"organizationId"`
	Name                    string `json:"name"`
	Description             string `json:"description"`
	ProjectLeaderId         string `json:"projectLeaderId"`
	ProjectLeaderName       string `json:"projectLeaderName"`
	ProjectLeaderAccountId  string `json:"projectLeaderAccountId"`
	ProjectLeaderDepartment string `json:"projectLeaderDepartment"`
	ProjectRoleId           string `json:"projectRoleId"`
	ProjectRoleName         string `json:"projectRoleName"`
	//AppCount                int    `json:"appCount"`
	//NamespaceCount          int        `json:"namespaceCount"`
	//MemberCount             int        `json:"memberCount"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt"`
}

type GetProjectResponse struct {
	Project *ProjectDetailResponse `json:"project"`
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

type CreateProjectRequest struct {
	Name            string `json:"name" validate:"required"`
	Description     string `json:"description"`
	ProjectLeaderId string `json:"projectLeaderId"`
}

type CreateProjectResponse struct {
	ProjectId string `json:"projectId"`
}

type UpdateProjectRequest struct {
	CreateProjectRequest
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

type ProjectMemberResponse struct {
	ID                    string     `json:"id"`
	ProjectId             string     `json:"projectId"`
	ProjectUserId         string     `json:"projectUserId"`
	ProjectUserName       string     `json:"projectUserName"`
	ProjectUserAccountId  string     `json:"projectUserAccountId"`
	ProjectUserEmail      string     `json:"projectUserEmail"`
	ProjectUserDepartment string     `json:"projectUserDepartment"`
	ProjectRoleId         string     `json:"projectRoleId"`
	ProjectRoleName       string     `json:"projectRoleName"`
	CreatedAt             time.Time  `json:"createdAt"`
	UpdatedAt             *time.Time `json:"updatedAt"`
}

type GetProjectMemberResponse struct {
	ProjectMember *ProjectMemberResponse `json:"projectMember"`
}

type GetProjectMembersResponse struct {
	ProjectMembers []ProjectMemberResponse `json:"projectMembers"`
	Pagination     PaginationResponse      `json:"pagination"`
}

type GetProjectMemberCountResponse struct {
	ProjectMemberAllCount int `json:"projectMemberAllCount"`
	ProjectLeaderCount    int `json:"projectLeaderCount"`
	ProjectMemberCount    int `json:"projectMemberCount"`
	ProjectViewerCount    int `json:"projectViewerCount"`
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

type UpdateProjectMembersRoleRequest struct {
	ProjectMemberRoleRequests []struct {
		ProjectMemberId string `json:"projectMemberId"`
		ProjectRoleId   string `json:"projectRoleId"`
	} `json:"projectMembers"`
}

type CreateProjectNamespaceRequest struct {
	StackId     string `json:"stackId"`
	Namespace   string `json:"namespace"`
	Description string `json:"description"`
}

type ProjectNamespaceResponse struct {
	StackId     string     `json:"stackId"`
	Namespace   string     `json:"namespace"`
	StackName   string     `json:"stackName"`
	ProjectId   string     `json:"projectId"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	AppCount    int        `json:"appCount"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   *time.Time `json:"updatedAt"`
}

type GetProjectNamespacesResponse struct {
	ProjectNamespaces []ProjectNamespaceResponse `json:"projectNamespaces"`
}

type GetProjectNamespaceResponse struct {
	ProjectNamespace *ProjectNamespaceResponse `json:"projectNamespace"`
}

type UpdateProjectNamespaceRequest struct {
	Description string `json:"description"`
}

type GetProjectKubeconfigResponse struct {
	Kubeconfig string `json:"kubeconfig"`
}

type ProjectNamespaceK8sResources struct {
	Pods         int       `json:"pods"`
	Deployments  int       `json:"deployments"`
	Statefulsets int       `json:"statefulsets"`
	Daemonsets   int       `json:"daemonsets"`
	Jobs         int       `json:"jobs"`
	Cronjobs     int       `json:"cronjobs"`
	PVCs         int       `json:"pvcs"`
	Services     int       `json:"services"`
	Ingresses    int       `json:"ingresses"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type GetProjectNamespaceK8sResourcesResponse struct {
	K8sResources ProjectNamespaceK8sResources `json:"k8sResources"`
}
