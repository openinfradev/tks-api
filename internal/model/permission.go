package model

import (
	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/delivery/api"
	"github.com/openinfradev/tks-api/internal/helper"
	"gorm.io/gorm"
)

type PermissionKind string

const (
	DashBoardPermission         PermissionKind = "대시보드"
	StackPermission             PermissionKind = "스택 관리"
	SecurityPolicyPermission    PermissionKind = "보안/정책 관리"
	ProjectManagementPermission PermissionKind = "프로젝트 관리"
	NotificationPermission      PermissionKind = "알림"
	ConfigurationPermission     PermissionKind = "설정"
)

type Permission struct {
	gorm.Model

	ID   uuid.UUID `gorm:"primarykey;type:uuid;" json:"ID"`
	Name string    `json:"name"`

	IsAllowed *bool       `gorm:"type:boolean;" json:"is_allowed,omitempty"`
	RoleID    *string     `json:"role_id,omitempty"`
	Role      *Role       `gorm:"foreignKey:RoleID;references:ID;" json:"role,omitempty"`
	Endpoints []*Endpoint `gorm:"many2many:permission_endpoints;" json:"endpoints,omitempty"`
	// omit empty

	ParentID *uuid.UUID    `json:"parent_id,omitempty"`
	Parent   *Permission   `gorm:"foreignKey:ParentID;references:ID;" json:"parent,omitempty"`
	Children []*Permission `gorm:"foreignKey:ParentID;references:ID;" json:"children,omitempty"`
}

type PermissionSet struct {
	Dashboard         *Permission `gorm:"-:all" json:"dashboard,omitempty"`
	Stack             *Permission `gorm:"-:all" json:"stack,omitempty"`
	SecurityPolicy    *Permission `gorm:"-:all" json:"security_policy,omitempty"`
	ProjectManagement *Permission `gorm:"-:all" json:"project_management,omitempty"`
	Notification      *Permission `gorm:"-:all" json:"notification,omitempty"`
	Configuration     *Permission `gorm:"-:all" json:"configuration,omitempty"`
}

func NewDefaultPermissionSet() *PermissionSet {
	return &PermissionSet{
		Dashboard:         newDashboard(),
		Stack:             newStack(),
		SecurityPolicy:    newSecurityPolicy(),
		ProjectManagement: newProjectManagement(),
		Notification:      newNotification(),
		Configuration:     newConfiguration(),
	}
}

func GetEdgePermission(root *Permission, edgePermissions []*Permission, f *func(permission Permission) bool) []*Permission {
	if root.Children == nil {
		return append(edgePermissions, root)
	}

	for _, child := range root.Children {
		if f != nil && !(*f)(*child) {
			continue
		}
		edgePermissions = GetEdgePermission(child, edgePermissions, f)
	}

	return edgePermissions
}

func endpointObjects(eps ...api.Endpoint) []*Endpoint {
	var result []*Endpoint
	for _, ep := range eps {
		result = append(result, &Endpoint{
			Name:  api.ApiMap[ep].Name,
			Group: api.ApiMap[ep].Group,
		})
	}
	return result
}

func newDashboard() *Permission {
	dashboard := &Permission{
		ID:   uuid.New(),
		Name: string(DashBoardPermission),
		Children: []*Permission{
			{
				ID:   uuid.New(),
				Name: "대시보드",
				Children: []*Permission{
					{
						ID:        uuid.New(),
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.GetChartsDashboard,
							api.GetChartDashboard,
							api.GetStacksDashboard,
							api.GetResourcesDashboard,
						),
					},
				},
			},
			{
				ID:   uuid.New(),
				Name: "대시보드 설정",
				Children: []*Permission{
					{
						ID:        uuid.New(),
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "생성",
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "수정",
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "삭제",
						IsAllowed: helper.BoolP(false),
					},
				},
			},
		},
	}

	return dashboard
}

func newStack() *Permission {
	stack := &Permission{
		ID:   uuid.New(),
		Name: string(StackPermission),
		Children: []*Permission{
			{
				ID:        uuid.New(),
				Name:      "조회",
				IsAllowed: helper.BoolP(false),
				Endpoints: endpointObjects(
					api.GetStacks,
					api.GetStack,
					api.CheckStackName,
					api.GetStackStatus,
					api.GetStackKubeConfig,

					api.SetFavoriteStack,
					api.DeleteFavoriteStack,
				),
			},
			{
				ID:        uuid.New(),
				Name:      "생성",
				IsAllowed: helper.BoolP(false),
				Endpoints: endpointObjects(
					api.CreateStack,
					api.InstallStack,
				),
			},
			{
				ID:        uuid.New(),
				Name:      "수정",
				IsAllowed: helper.BoolP(false),
				Endpoints: endpointObjects(
					api.UpdateStack,
				),
			},
			{
				ID:        uuid.New(),
				Name:      "삭제",
				IsAllowed: helper.BoolP(false),
				Endpoints: endpointObjects(
					api.DeleteStack,
				),
			},
		},
	}

	return stack
}

func newSecurityPolicy() *Permission {
	security_policy := &Permission{
		ID:   uuid.New(),
		Name: string(SecurityPolicyPermission),
		Children: []*Permission{
			{
				ID:   uuid.New(),
				Name: "보안/정책",
				Children: []*Permission{
					{
						ID:        uuid.New(),
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "생성",
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "수정",
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "삭제",
						IsAllowed: helper.BoolP(false),
					},
				},
			},
		},
	}

	return security_policy
}

func newProjectManagement() *Permission {
	projectManagement := &Permission{
		ID:   uuid.New(),
		Name: string(ProjectManagementPermission),
		Children: []*Permission{
			{
				ID:   uuid.New(),
				Name: "프로젝트",
				Children: []*Permission{
					{
						ID:        uuid.New(),
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.GetProjects,
							api.GetProject,
						),
					},
					{
						ID:        uuid.New(),
						Name:      "생성",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.CreateProject,
						),
					},
				},
			},
			{
				ID:   uuid.New(),
				Name: "앱 서빙",
				Children: []*Permission{
					{
						ID:        uuid.New(),
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.GetAppServeApps,
							api.GetAppServeApp,
							api.GetNumOfAppsOnStack,
							api.GetAppServeAppLatestTask,
							api.IsAppServeAppExist,
							api.IsAppServeAppNameExist,
						),
					},
					{
						ID:        uuid.New(),
						Name:      "빌드",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.CreateAppServeApp,
							api.IsAppServeAppExist,
							api.IsAppServeAppNameExist,
							api.UpdateAppServeApp,
							api.UpdateAppServeAppEndpoint,
							api.UpdateAppServeAppStatus,
							api.RollbackAppServeApp,
						),
					},
					{
						ID:        uuid.New(),
						Name:      "배포",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.CreateAppServeApp,
							api.IsAppServeAppExist,
							api.IsAppServeAppNameExist,
							api.UpdateAppServeApp,
							api.UpdateAppServeAppEndpoint,
							api.UpdateAppServeAppStatus,
							api.RollbackAppServeApp,
						),
					},
					{
						ID:        uuid.New(),
						Name:      "삭제",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.DeleteAppServeApp,
						),
					},
				},
			},
			{
				ID:   uuid.New(),
				Name: "설정-일반",
				Children: []*Permission{
					{
						ID:        uuid.New(),
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.GetProjects,
							api.GetProject,

							api.GetProjectRoles,
							api.GetProjectRole,
						),
					},
					{
						ID:        uuid.New(),
						Name:      "수정",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.UpdateProject,
						),
					},
					{
						ID:        uuid.New(),
						Name:      "삭제",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.DeleteProject,
						),
					},
				},
			},
			{
				ID:   uuid.New(),
				Name: "설정-멤버",
				Children: []*Permission{
					{
						ID:        uuid.New(),
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.GetProjectMembers,
							api.GetProjectMember,
							api.GetProjectRoles,
							api.GetProjectRole,
						),
					},
					{
						ID:        uuid.New(),
						Name:      "생성",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.AddProjectMember,
						),
					},
					{
						ID:        uuid.New(),
						Name:      "수정",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.UpdateProjectMemberRole,
						),
					},
					{
						ID:        uuid.New(),
						Name:      "삭제",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.RemoveProjectMember,
						),
					},
				},
			},
			{
				ID:   uuid.New(),
				Name: "설정-네임스페이스",
				Children: []*Permission{
					{
						ID:        uuid.New(),
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.GetProjectNamespaces,
							api.GetProjectNamespace,
						),
					},
					{
						ID:        uuid.New(),
						Name:      "생성",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.CreateProjectNamespace,
						),
					},
					{
						ID:        uuid.New(),
						Name:      "수정",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(),
					},
					{
						ID:        uuid.New(),
						Name:      "삭제",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.DeleteProjectNamespace,
						),
					},
				},
			},
		},
	}

	return projectManagement
}

func newNotification() *Permission {
	notification := &Permission{
		ID:   uuid.New(),
		Name: string(NotificationPermission),
		Children: []*Permission{
			{
				ID:   uuid.New(),
				Name: "시스템 경고",
				Children: []*Permission{
					{
						ID:        uuid.New(),
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
					},
				},
			},
			{
				ID:   uuid.New(),
				Name: "보안/정책 감사로그",
				Children: []*Permission{
					{
						ID:        uuid.New(),
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
					},
				},
			},
		},
	}

	return notification
}

func newConfiguration() *Permission {
	configuration := &Permission{
		ID:   uuid.New(),
		Name: string(ConfigurationPermission),
		Children: []*Permission{
			{
				ID:   uuid.New(),
				Name: "일반",
				Children: []*Permission{
					{
						ID:        uuid.New(),
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "수정",
						IsAllowed: helper.BoolP(false),
					},
				},
			},
			{
				ID:   uuid.New(),
				Name: "클라우드 계정",
				Children: []*Permission{
					{
						ID:        uuid.New(),
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "생성",
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "수정",
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "삭제",
						IsAllowed: helper.BoolP(false),
					},
				},
			},
			{
				ID:   uuid.New(),
				Name: "스택 템플릿",
				Children: []*Permission{
					{
						ID:        uuid.New(),
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
					},
				},
			},
			{
				ID:   uuid.New(),
				Name: "프로젝트 관리",
				Children: []*Permission{
					{
						ID:        uuid.New(),
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "생성",
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "수정",
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "삭제",
						IsAllowed: helper.BoolP(false),
					},
				},
			},
			{
				ID:   uuid.New(),
				Name: "사용자",
				Children: []*Permission{
					{
						ID:        uuid.New(),
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "생성",
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "수정",
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "삭제",
						IsAllowed: helper.BoolP(false),
					},
				},
			},
			{
				ID:   uuid.New(),
				Name: "사용자 권한 관리",
				Children: []*Permission{
					{
						ID:        uuid.New(),
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "생성",
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "수정",
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "삭제",
						IsAllowed: helper.BoolP(false),
					},
				},
			},
			{
				ID:   uuid.New(),
				Name: "알림 설정",
				Children: []*Permission{
					{
						ID:        uuid.New(),
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "생성",
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "수정",
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "삭제",
						IsAllowed: helper.BoolP(false),
					},
				},
			},
		},
	}

	return configuration
}

func (p *PermissionSet) SetAllowedPermissionSet() {
	edgePermissions := make([]*Permission, 0)
	edgePermissions = append(edgePermissions, GetEdgePermission(p.Dashboard, edgePermissions, nil)...)
	edgePermissions = append(edgePermissions, GetEdgePermission(p.Stack, edgePermissions, nil)...)
	edgePermissions = append(edgePermissions, GetEdgePermission(p.SecurityPolicy, edgePermissions, nil)...)
	edgePermissions = append(edgePermissions, GetEdgePermission(p.ProjectManagement, edgePermissions, nil)...)
	edgePermissions = append(edgePermissions, GetEdgePermission(p.Notification, edgePermissions, nil)...)
	edgePermissions = append(edgePermissions, GetEdgePermission(p.Configuration, edgePermissions, nil)...)

	for _, permission := range edgePermissions {
		permission.IsAllowed = helper.BoolP(true)
	}
}

func (p *PermissionSet) SetUserPermissionSet() {
	f := func(permission Permission) bool {
		return permission.Name == "조회"
	}
	edgePermissions := make([]*Permission, 0)
	edgePermissions = append(edgePermissions, GetEdgePermission(p.Dashboard, edgePermissions, nil)...)
	edgePermissions = append(edgePermissions, GetEdgePermission(p.Stack, edgePermissions, &f)...)
	edgePermissions = append(edgePermissions, GetEdgePermission(p.SecurityPolicy, edgePermissions, &f)...)
	edgePermissions = append(edgePermissions, GetEdgePermission(p.ProjectManagement, edgePermissions, &f)...)
	edgePermissions = append(edgePermissions, GetEdgePermission(p.Notification, edgePermissions, &f)...)
	//edgePermissions = append(edgePermissions, GetEdgePermission(p.Configuration, edgePermissions, &f)...)

	for _, permission := range edgePermissions {
		permission.IsAllowed = helper.BoolP(true)
	}
}

func (p *PermissionSet) SetRoleId(roleId string) {
	setRoleIdToPermission(p.Dashboard, roleId)
	setRoleIdToPermission(p.Stack, roleId)
	setRoleIdToPermission(p.SecurityPolicy, roleId)
	setRoleIdToPermission(p.ProjectManagement, roleId)
	setRoleIdToPermission(p.Notification, roleId)
	setRoleIdToPermission(p.Configuration, roleId)
}

func setRoleIdToPermission(root *Permission, roleId string) {
	root.RoleID = helper.StringP(roleId)

	if root.Children == nil {
		return
	}

	for _, child := range root.Children {
		setRoleIdToPermission(child, roleId)
	}
}
