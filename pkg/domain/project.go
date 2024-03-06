package domain

import "time"

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

type ProjectRoleResponse struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"` // project-leader, project-member, project-viewer
	Description string     `json:"description,omitempty"`
	CreatedAt   time.Time  `json:"createdAt" `
	UpdatedAt   *time.Time `json:"updatedAt"`
	DeletedAt   *time.Time `json:"deletedAt"`
}

type GetProjectRoleResponse struct {
	ProjectRole ProjectRoleResponse `json:"projectRole"`
}

type GetProjectRolesResponse struct {
	ProjectRoles []ProjectRoleResponse `json:"projectRoles"`
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

type ProjectNamespaceResourcesUsage struct {
	CPU    int `json:"cpu"`
	Memory int `json:"memory"`
	PV     int `json:"pv"`
}

type GetProjectNamespaceResourcesUsageResponse struct {
	ResourcesUsage ProjectNamespaceResourcesUsage `json:"resourcesUsage"`
}
