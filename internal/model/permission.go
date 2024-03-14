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

func (p *Permission) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
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

func SetRoleIDToPermission(roleID string, permission *Permission) {
	permission.RoleID = helper.StringP(roleID)
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
		Name: string(DashBoardPermission),
		Children: []*Permission{
			{
				Name: "대시보드",
				Children: []*Permission{
					{
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
				Name: "대시보드 설정",
				Children: []*Permission{
					{
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
					},
					{
						Name:      "생성",
						IsAllowed: helper.BoolP(false),
					},
					{
						Name:      "수정",
						IsAllowed: helper.BoolP(false),
					},
					{
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
		Name: string(StackPermission),
		Children: []*Permission{
			{
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
				Name:      "생성",
				IsAllowed: helper.BoolP(false),
				Endpoints: endpointObjects(
					api.CreateStack,
					api.InstallStack,
				),
			},
			{
				Name:      "수정",
				IsAllowed: helper.BoolP(false),
				Endpoints: endpointObjects(
					api.UpdateStack,
				),
			},
			{
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
		Name: string(SecurityPolicyPermission),
		Children: []*Permission{
			{
				Name: "보안/정책",
				Children: []*Permission{
					{
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
					},
					{
						Name:      "생성",
						IsAllowed: helper.BoolP(false),
					},
					{
						Name:      "수정",
						IsAllowed: helper.BoolP(false),
					},
					{
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
		Name: string(ProjectManagementPermission),
		Children: []*Permission{
			{
				Name: "프로젝트",
				Children: []*Permission{
					{
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.GetProjects,
							api.GetProject,
						),
					},
					{
						Name:      "생성",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.CreateProject,
						),
					},
				},
			},
			{
				Name: "앱 서빙",
				Children: []*Permission{
					{
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
						Name:      "삭제",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.DeleteAppServeApp,
						),
					},
				},
			},
			{
				Name: "설정-일반",
				Children: []*Permission{
					{
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
						Name:      "수정",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.UpdateProject,
						),
					},
					{
						Name:      "삭제",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.DeleteProject,
						),
					},
				},
			},
			{
				Name: "설정-멤버",
				Children: []*Permission{
					{
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
						Name:      "생성",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.AddProjectMember,
						),
					},
					{
						Name:      "수정",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.UpdateProjectMemberRole,
						),
					},
					{
						Name:      "삭제",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.RemoveProjectMember,
						),
					},
				},
			},
			{
				Name: "설정-네임스페이스",
				Children: []*Permission{
					{
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.GetProjectNamespaces,
							api.GetProjectNamespace,
						),
					},
					{
						Name:      "생성",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.CreateProjectNamespace,
						),
					},
					{
						Name:      "수정",
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(),
					},
					{
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
		Name: string(NotificationPermission),
		Children: []*Permission{
			{
				Name: "시스템 경고",
				Children: []*Permission{
					{
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
					},
				},
			},
			{
				Name: "보안/정책 감사로그",
				Children: []*Permission{
					{
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
		Name: string(ConfigurationPermission),
		Children: []*Permission{
			{
				Name: "일반",
				Children: []*Permission{
					{
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
					},
					{
						Name:      "수정",
						IsAllowed: helper.BoolP(false),
					},
				},
			},
			{
				Name: "클라우드 계정",
				Children: []*Permission{
					{
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
					},
					{
						Name:      "생성",
						IsAllowed: helper.BoolP(false),
					},
					{
						Name:      "수정",
						IsAllowed: helper.BoolP(false),
					},
					{
						Name:      "삭제",
						IsAllowed: helper.BoolP(false),
					},
				},
			},
			{
				Name: "스택 템플릿",
				Children: []*Permission{
					{
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
					},
				},
			},
			{
				Name: "프로젝트 관리",
				Children: []*Permission{
					{
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
					},
					{
						Name:      "생성",
						IsAllowed: helper.BoolP(false),
					},
					{
						Name:      "수정",
						IsAllowed: helper.BoolP(false),
					},
					{
						Name:      "삭제",
						IsAllowed: helper.BoolP(false),
					},
				},
			},
			{
				Name: "사용자",
				Children: []*Permission{
					{
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
					},
					{
						Name:      "생성",
						IsAllowed: helper.BoolP(false),
					},
					{
						Name:      "수정",
						IsAllowed: helper.BoolP(false),
					},
					{
						Name:      "삭제",
						IsAllowed: helper.BoolP(false),
					},
				},
			},
			{
				Name: "사용자 권한 관리",
				Children: []*Permission{
					{
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
					},
					{
						Name:      "생성",
						IsAllowed: helper.BoolP(false),
					},
					{
						Name:      "수정",
						IsAllowed: helper.BoolP(false),
					},
					{
						Name:      "삭제",
						IsAllowed: helper.BoolP(false),
					},
				},
			},
			{
				Name: "알림 설정",
				Children: []*Permission{
					{
						Name:      "조회",
						IsAllowed: helper.BoolP(false),
					},
					{
						Name:      "생성",
						IsAllowed: helper.BoolP(false),
					},
					{
						Name:      "수정",
						IsAllowed: helper.BoolP(false),
					},
					{
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

	return
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

	return
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
