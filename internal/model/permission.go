package model

import (
	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/helper"
	"gorm.io/gorm"
)

type PermissionKind string

var SortOrder = map[string]int{
	OperationRead:     0,
	OperationCreate:   1,
	OperationUpdate:   2,
	OperationDelete:   3,
	OperationDownload: 4,
}

const (
	DashBoardPermission     PermissionKind = "대시보드"
	StackPermission         PermissionKind = "스택"
	PolicyPermission        PermissionKind = "정책"
	ProjectPermission       PermissionKind = "프로젝트"
	NotificationPermission  PermissionKind = "알림"
	ConfigurationPermission PermissionKind = "설정"

	OperationRead     = "READ"
	OperationCreate   = "CREATE"
	OperationUpdate   = "UPDATE"
	OperationDelete   = "DELETE"
	OperationDownload = "DOWNLOAD"

	// Key
	TopDashboardKey                          = "DASHBOARD"
	MiddleDashboardKey                       = "DASHBOARD-DASHBOARD"
	TopStackKey                              = "STACK"
	MiddleStackKey                           = "STACK-STACK"
	TopPolicyKey                             = "POLICY"
	MiddlePolicyKey                          = "POLICY-POLICY"
	TopNotificationKey                       = "NOTIFICATION"
	MiddleNotificationKey                    = "NOTIFICATION-SYSTEM_NOTIFICATION"
	MiddlePolicyNotificationKey              = "NOTIFICATION-POLICY_NOTIFICATION"
	TopProjectKey                            = "PROJECT"
	MiddleProjectKey                         = "PROJECT-PROJECT_LIST"
	MiddleProjectCommonConfigurationKey      = "PROJECT-PROJECT_COMMON_CONFIGURATION"
	MiddleProjectMemberConfigurationKey      = "PROJECT-PROJECT_MEMBER_CONFIGURATION"
	MiddleProjectNamespaceKey                = "PROJECT-PROJECT_NAMESPACE"
	MiddleProjectAppServeKey                 = "PROJECT-PROJECT_APP_SERVE"
	TopConfigurationKey                      = "CONFIGURATION"
	MiddleConfigurationKey                   = "CONFIGURATION-CONFIGURATION"
	MiddleConfigurationCloudAccountKey       = "CONFIGURATION-CLOUD_ACCOUNT"
	MiddleConfigurationProjectKey            = "CONFIGURATION-PROJECT"
	MiddleConfigurationUserKey               = "CONFIGURATION-USER"
	MiddleConfigurationRoleKey               = "CONFIGURATION-ROLE"
	MiddleConfigurationSystemNotificationKey = "CONFIGURATION-SYSTEM_NOTIFICATION"
	CommonKey                                = "COMMON"
)

type Permission struct {
	gorm.Model

	ID      uuid.UUID `gorm:"primarykey;type:uuid;" json:"ID"`
	Name    string    `json:"name"`
	Key     string    `gorm:"type:text;" json:"key,omitempty"`
	EdgeKey *string   `gorm:"type:text;"`

	IsAllowed *bool       `gorm:"type:boolean;" json:"is_allowed,omitempty"`
	RoleID    *string     `json:"role_id,omitempty"`
	Role      *Role       `gorm:"foreignKey:RoleID;references:ID;" json:"role,omitempty"`
	Endpoints []*Endpoint `gorm:"many2many:permission_endpoints;joinForeignKey:EdgeKey;joinReferences:EndpointName;" json:"endpoints,omitempty"`

	ParentID *uuid.UUID    `json:"parent_id,omitempty"`
	Parent   *Permission   `gorm:"foreignKey:ParentID;references:ID;" json:"parent,omitempty"`
	Children []*Permission `gorm:"foreignKey:ParentID;references:ID;" json:"children,omitempty"`
}

type PermissionSet struct {
	Dashboard         *Permission `gorm:"-:all" json:"dashboard,omitempty"`
	Stack             *Permission `gorm:"-:all" json:"stack,omitempty"`
	Policy            *Permission `gorm:"-:all" json:"policy,omitempty"`
	ProjectManagement *Permission `gorm:"-:all" json:"project_management,omitempty"`
	Notification      *Permission `gorm:"-:all" json:"notification,omitempty"`
	Configuration     *Permission `gorm:"-:all" json:"configuration,omitempty"`
	Common            *Permission `gorm:"-:all" json:"common,omitempty"`
	// ToDo: Need to consider  whether to use Admin Permission
	//Admin             *Permission `gorm:"-:all" json:"admin,omitempty"`
}

func NewDefaultPermissionSet() *PermissionSet {
	return &PermissionSet{
		Dashboard:         newDashboard(),
		Stack:             newStack(),
		Policy:            newPolicy(),
		ProjectManagement: newProject(),
		Notification:      newNotification(),
		Configuration:     newConfiguration(),
		Common:            newCommon(),
		//Admin:             nil,
	}
}

func NewAdminPermissionSet() *PermissionSet {
	return &PermissionSet{
		//Admin:             newAdmin(),
		Dashboard:         newDashboard(),
		Stack:             newStack(),
		Policy:            newPolicy(),
		ProjectManagement: newProject(),
		Notification:      newNotification(),
		Configuration:     newConfiguration(),
		Common:            newCommon(),
	}
}

func GetEdgePermission(root *Permission, edgePermissions []*Permission, f *func(permission Permission) bool) []*Permission {
	if root.Children == nil || len(root.Children) == 0 {
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

func newDashboard() *Permission {
	dashboard := &Permission{
		ID:   uuid.New(),
		Name: string(DashBoardPermission),
		Key:  TopDashboardKey,
		Children: []*Permission{
			{
				ID:   uuid.New(),
				Name: "대시보드",
				Key:  MiddleDashboardKey,
				Children: []*Permission{
					{
						ID:        uuid.New(),
						Name:      "조회",
						Key:       OperationRead,
						EdgeKey:   helper.StringP(TopDashboardKey + "-" + MiddleDashboardKey + "-" + OperationRead),
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "수정",
						Key:       OperationUpdate,
						EdgeKey:   helper.StringP(TopDashboardKey + "-" + MiddleDashboardKey + "-" + OperationUpdate),
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
		Key:  TopStackKey,
		Children: []*Permission{
			{
				ID:   uuid.New(),
				Name: "스택",
				Key:  MiddleStackKey,
				Children: []*Permission{
					{
						ID:        uuid.New(),
						Name:      "조회",
						Key:       OperationRead,
						EdgeKey:   helper.StringP(TopStackKey + "-" + MiddleStackKey + "-" + OperationRead),
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "생성",
						Key:       OperationCreate,
						EdgeKey:   helper.StringP(TopStackKey + "-" + MiddleStackKey + "-" + OperationCreate),
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "수정",
						Key:       OperationUpdate,
						EdgeKey:   helper.StringP(TopStackKey + "-" + MiddleStackKey + "-" + OperationUpdate),
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "삭제",
						Key:       OperationDelete,
						EdgeKey:   helper.StringP(TopStackKey + "-" + MiddleStackKey + "-" + OperationDelete),
						IsAllowed: helper.BoolP(false),
					},
				},
			},
		},
	}

	return stack
}

func newPolicy() *Permission {
	policy := &Permission{
		ID:   uuid.New(),
		Name: string(PolicyPermission),
		Key:  TopPolicyKey,
		Children: []*Permission{
			{
				ID:   uuid.New(),
				Name: "정책",
				Key:  MiddlePolicyKey,
				Children: []*Permission{
					{
						ID:        uuid.New(),
						Name:      "조회",
						Key:       OperationRead,
						EdgeKey:   helper.StringP(TopPolicyKey + "-" + MiddlePolicyKey + "-" + OperationRead),
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "생성",
						Key:       OperationCreate,
						EdgeKey:   helper.StringP(TopPolicyKey + "-" + MiddlePolicyKey + "-" + OperationCreate),
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "수정",
						Key:       OperationUpdate,
						EdgeKey:   helper.StringP(TopPolicyKey + "-" + MiddlePolicyKey + "-" + OperationUpdate),
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "삭제",
						Key:       OperationDelete,
						EdgeKey:   helper.StringP(TopPolicyKey + "-" + MiddlePolicyKey + "-" + OperationDelete),
						IsAllowed: helper.BoolP(false),
					},
				},
			},
		},
	}

	return policy
}

func newNotification() *Permission {
	notification := &Permission{
		ID:   uuid.New(),
		Name: string(NotificationPermission),
		Key:  TopNotificationKey,
		Children: []*Permission{
			{
				ID:   uuid.New(),
				Name: "시스템 알림",
				Key:  MiddleNotificationKey,
				Children: []*Permission{
					{
						ID:        uuid.New(),
						Name:      "조회",
						Key:       OperationRead,
						EdgeKey:   helper.StringP(TopNotificationKey + "-" + MiddleNotificationKey + "-" + OperationRead),
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "수정",
						Key:       OperationUpdate,
						EdgeKey:   helper.StringP(TopNotificationKey + "-" + MiddleNotificationKey + "-" + OperationUpdate),
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "다운로드",
						Key:       OperationDownload,
						EdgeKey:   helper.StringP(TopNotificationKey + "-" + MiddleNotificationKey + "-" + OperationDownload),
						IsAllowed: helper.BoolP(false),
						Children:  []*Permission{},
					},
				},
			},
			{
				ID:   uuid.New(),
				Name: "정책 알림",
				Key:  MiddlePolicyNotificationKey,
				Children: []*Permission{
					{
						ID:        uuid.New(),
						Name:      "조회",
						Key:       OperationRead,
						EdgeKey:   helper.StringP(TopNotificationKey + "-" + MiddlePolicyNotificationKey + "-" + OperationRead),
						IsAllowed: helper.BoolP(false),
						Children:  []*Permission{},
					},
					{
						ID:        uuid.New(),
						Name:      "다운로드",
						Key:       OperationDownload,
						EdgeKey:   helper.StringP(TopNotificationKey + "-" + MiddlePolicyNotificationKey + "-" + OperationDownload),
						IsAllowed: helper.BoolP(false),
						Children:  []*Permission{},
					},
				},
			},
		},
	}

	return notification
}

func newProject() *Permission {
	project := &Permission{
		ID:   uuid.New(),
		Name: string(ProjectPermission),
		Key:  TopProjectKey,
		Children: []*Permission{
			{
				ID:   uuid.New(),
				Name: "프로젝트 목록",
				Key:  MiddleProjectKey,
				Children: []*Permission{
					{
						ID:        uuid.New(),
						Name:      "조회",
						Key:       OperationRead,
						EdgeKey:   helper.StringP(TopProjectKey + "-" + MiddleProjectKey + "-" + OperationRead),
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "생성",
						Key:       OperationCreate,
						EdgeKey:   helper.StringP(TopProjectKey + "-" + MiddleProjectKey + "-" + OperationCreate),
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "수정",
						Key:       OperationUpdate,
						EdgeKey:   helper.StringP(TopProjectKey + "-" + MiddleProjectKey + "-" + OperationUpdate),
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "삭제",
						Key:       OperationDelete,
						EdgeKey:   helper.StringP(TopProjectKey + "-" + MiddleProjectKey + "-" + OperationDelete),
						IsAllowed: helper.BoolP(false),
					},
				},
			},
			{
				ID:   uuid.New(),
				Name: "일반 설정",
				Key:  MiddleProjectCommonConfigurationKey,
				Children: []*Permission{
					{
						ID:        uuid.New(),
						Name:      "조회",
						Key:       OperationRead,
						EdgeKey:   helper.StringP(TopProjectKey + "-" + MiddleProjectCommonConfigurationKey + "-" + OperationRead),
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "수정",
						Key:       OperationUpdate,
						EdgeKey:   helper.StringP(TopProjectKey + "-" + MiddleProjectCommonConfigurationKey + "-" + OperationUpdate),
						IsAllowed: helper.BoolP(false),
					},
				},
			},
			{
				ID:   uuid.New(),
				Name: "구성원 설정",
				Key:  MiddleProjectMemberConfigurationKey,
				Children: []*Permission{
					{
						ID:        uuid.New(),
						Name:      "조회",
						Key:       OperationRead,
						EdgeKey:   helper.StringP(TopProjectKey + "-" + MiddleProjectMemberConfigurationKey + "-" + OperationRead),
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "생성",
						Key:       OperationCreate,
						EdgeKey:   helper.StringP(TopProjectKey + "-" + MiddleProjectMemberConfigurationKey + "-" + OperationCreate),
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "수정",
						Key:       OperationUpdate,
						EdgeKey:   helper.StringP(TopProjectKey + "-" + MiddleProjectMemberConfigurationKey + "-" + OperationUpdate),
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "삭제",
						Key:       OperationDelete,
						EdgeKey:   helper.StringP(TopProjectKey + "-" + MiddleProjectMemberConfigurationKey + "-" + OperationDelete),
						IsAllowed: helper.BoolP(false),
					},
				},
			},
			{
				ID:   uuid.New(),
				Name: "네임스페이스",
				Key:  MiddleProjectNamespaceKey,
				Children: []*Permission{
					{
						ID:        uuid.New(),
						Name:      "조회",
						Key:       OperationRead,
						EdgeKey:   helper.StringP(TopProjectKey + "-" + MiddleProjectNamespaceKey + "-" + OperationRead),
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "생성",
						Key:       OperationCreate,
						EdgeKey:   helper.StringP(TopProjectKey + "-" + MiddleProjectNamespaceKey + "-" + OperationCreate),
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "수정",
						Key:       OperationUpdate,
						EdgeKey:   helper.StringP(TopProjectKey + "-" + MiddleProjectNamespaceKey + "-" + OperationUpdate),
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "삭제",
						Key:       OperationDelete,
						EdgeKey:   helper.StringP(TopProjectKey + "-" + MiddleProjectNamespaceKey + "-" + OperationDelete),
						IsAllowed: helper.BoolP(false),
					},
				},
			},
			{
				ID:   uuid.New(),
				Name: "앱 서빙",
				Key:  MiddleProjectAppServeKey,
				Children: []*Permission{
					{
						ID:        uuid.New(),
						Name:      "조회",
						Key:       OperationRead,
						EdgeKey:   helper.StringP(TopProjectKey + "-" + MiddleProjectAppServeKey + "-" + OperationRead),
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "생성",
						Key:       OperationCreate,
						EdgeKey:   helper.StringP(TopProjectKey + "-" + MiddleProjectAppServeKey + "-" + OperationCreate),
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "수정",
						Key:       OperationUpdate,
						EdgeKey:   helper.StringP(TopProjectKey + "-" + MiddleProjectAppServeKey + "-" + OperationUpdate),
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "삭제",
						Key:       OperationDelete,
						EdgeKey:   helper.StringP(TopProjectKey + "-" + MiddleProjectAppServeKey + "-" + OperationDelete),
						IsAllowed: helper.BoolP(false),
					},
				},
			},
		},
	}

	return project
}

func newConfiguration() *Permission {
	configuration := &Permission{
		ID:   uuid.New(),
		Name: string(ConfigurationPermission),
		Key:  TopConfigurationKey,
		Children: []*Permission{
			{
				ID:   uuid.New(),
				Name: "일반",
				Key:  MiddleConfigurationKey,
				Children: []*Permission{
					{
						ID:        uuid.New(),
						Name:      "조회",
						Key:       OperationRead,
						EdgeKey:   helper.StringP(TopConfigurationKey + "-" + MiddleConfigurationKey + "-" + OperationRead),
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "수정",
						Key:       OperationUpdate,
						EdgeKey:   helper.StringP(TopConfigurationKey + "-" + MiddleConfigurationKey + "-" + OperationUpdate),
						IsAllowed: helper.BoolP(false),
					},
				},
			},
			{
				ID:   uuid.New(),
				Name: "클라우드 계정",
				Key:  MiddleConfigurationCloudAccountKey,
				Children: []*Permission{
					{
						ID:        uuid.New(),
						Name:      "조회",
						Key:       OperationRead,
						EdgeKey:   helper.StringP(TopConfigurationKey + "-" + MiddleConfigurationCloudAccountKey + "-" + OperationRead),
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "생성",
						Key:       OperationCreate,
						EdgeKey:   helper.StringP(TopConfigurationKey + "-" + MiddleConfigurationCloudAccountKey + "-" + OperationCreate),
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "수정",
						Key:       OperationUpdate,
						EdgeKey:   helper.StringP(TopConfigurationKey + "-" + MiddleConfigurationCloudAccountKey + "-" + OperationUpdate),
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "삭제",
						Key:       OperationDelete,
						EdgeKey:   helper.StringP(TopConfigurationKey + "-" + MiddleConfigurationCloudAccountKey + "-" + OperationDelete),
						IsAllowed: helper.BoolP(false),
					},
				},
			},
			{
				ID:   uuid.New(),
				Name: "프로젝트",
				Key:  MiddleConfigurationProjectKey,
				Children: []*Permission{
					{
						ID:        uuid.New(),
						Name:      "조회",
						Key:       OperationRead,
						EdgeKey:   helper.StringP(TopConfigurationKey + "-" + MiddleConfigurationProjectKey + "-" + OperationRead),
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "생성",
						Key:       OperationCreate,
						EdgeKey:   helper.StringP(TopConfigurationKey + "-" + MiddleConfigurationProjectKey + "-" + OperationCreate),
						IsAllowed: helper.BoolP(false),
					},
				},
			},
			{
				ID:   uuid.New(),
				Name: "사용자",
				Key:  MiddleConfigurationUserKey,
				Children: []*Permission{
					{
						ID:        uuid.New(),
						Name:      "조회",
						Key:       OperationRead,
						EdgeKey:   helper.StringP(TopConfigurationKey + "-" + MiddleConfigurationUserKey + "-" + OperationRead),
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "생성",
						Key:       OperationCreate,
						EdgeKey:   helper.StringP(TopConfigurationKey + "-" + MiddleConfigurationUserKey + "-" + OperationCreate),
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "수정",
						Key:       OperationUpdate,
						EdgeKey:   helper.StringP(TopConfigurationKey + "-" + MiddleConfigurationUserKey + "-" + OperationUpdate),
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "삭제",
						Key:       OperationDelete,
						EdgeKey:   helper.StringP(TopConfigurationKey + "-" + MiddleConfigurationUserKey + "-" + OperationDelete),
						IsAllowed: helper.BoolP(false),
					},
				},
			},
			{
				ID:   uuid.New(),
				Name: "역할 및 권한",
				Key:  MiddleConfigurationRoleKey,
				Children: []*Permission{
					{
						ID:        uuid.New(),
						Name:      "조회",
						Key:       OperationRead,
						EdgeKey:   helper.StringP(TopConfigurationKey + "-" + MiddleConfigurationRoleKey + "-" + OperationRead),
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "생성",
						Key:       OperationCreate,
						EdgeKey:   helper.StringP(TopConfigurationKey + "-" + MiddleConfigurationRoleKey + "-" + OperationCreate),
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "수정",
						Key:       OperationUpdate,
						EdgeKey:   helper.StringP(TopConfigurationKey + "-" + MiddleConfigurationRoleKey + "-" + OperationUpdate),
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "삭제",
						Key:       OperationDelete,
						EdgeKey:   helper.StringP(TopConfigurationKey + "-" + MiddleConfigurationRoleKey + "-" + OperationDelete),
						IsAllowed: helper.BoolP(false),
					},
				},
			},
			{
				ID:   uuid.New(),
				Name: "시스템 알림",
				Key:  MiddleConfigurationSystemNotificationKey,
				Children: []*Permission{
					{
						ID:        uuid.New(),
						Name:      "조회",
						Key:       OperationRead,
						EdgeKey:   helper.StringP(TopConfigurationKey + "-" + MiddleConfigurationSystemNotificationKey + "-" + OperationRead),
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "생성",
						Key:       OperationCreate,
						EdgeKey:   helper.StringP(TopConfigurationKey + "-" + MiddleConfigurationSystemNotificationKey + "-" + OperationCreate),
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "수정",
						Key:       OperationUpdate,
						EdgeKey:   helper.StringP(TopConfigurationKey + "-" + MiddleConfigurationSystemNotificationKey + "-" + OperationUpdate),
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "삭제",
						Key:       OperationDelete,
						EdgeKey:   helper.StringP(TopConfigurationKey + "-" + MiddleConfigurationSystemNotificationKey + "-" + OperationDelete),
						IsAllowed: helper.BoolP(false),
					},
				},
			},
		},
	}

	return configuration
}

func newCommon() *Permission {
	common := &Permission{
		ID:        uuid.New(),
		Name:      "공통",
		IsAllowed: helper.BoolP(true),
		Key:       CommonKey,
		EdgeKey:   helper.StringP(CommonKey),
	}

	return common

}

//func newAdmin() *Permission {
//	admin := &Permission{
//		ID:        uuid.New(),
//		Name:      "관리자",
//		IsAllowed: helper.BoolP(true),
//		EdgeKey:   helper.StringP("admin"),
//		Endpoints: endpointObjects(
//			// Organization
//			api.Admin_CreateOrganization,
//			api.Admin_DeleteOrganization,
//			api.UpdateOrganization,
//			api.GetOrganization,
//			api.GetOrganizations,
//			api.UpdatePrimaryCluster,
//			api.CheckOrganizationName,
//
//			// User
//			api.ResetPassword,
//			api.CheckId,
//			api.CheckEmail,
//
//			// StackTemplate
//			api.Admin_GetStackTemplates,
//			api.Admin_GetStackTemplate,
//			api.Admin_GetStackTemplateServices,
//			api.Admin_CreateStackTemplate,
//			api.Admin_UpdateStackTemplate,
//			api.Admin_DeleteStackTemplate,
//			api.Admin_UpdateStackTemplateOrganizations,
//			api.Admin_CheckStackTemplateName,
//
//			// Admin
//			api.Admin_GetUser,
//			api.Admin_ListUser,
//			api.Admin_CreateUser,
//			api.Admin_UpdateUser,
//			api.Admin_DeleteUser,
//			api.Admin_GetSystemNotificationTemplate,
//			api.Admin_CreateSystemNotificationTemplate,
//			api.Admin_ListUser,
//			api.Admin_GetTksRole,
//			api.Admin_GetProjects,
//			api.Admin_UpdateSystemNotificationTemplate,
//			api.Admin_ListTksRoles,
//			api.Admin_GetSystemNotificationTemplates,
//
//			// Audit
//			api.GetAudits,
//			api.GetAudit,
//			api.DeleteAudit,
//
//			api.CreateSystemNotification,
//			api.DeleteSystemNotification,
//		),
//	}
//
//	return admin
//}

func (p *PermissionSet) SetAllowedPermissionSet() {
	edgePermissions := make([]*Permission, 0)
	edgePermissions = append(edgePermissions, GetEdgePermission(p.Dashboard, edgePermissions, nil)...)
	edgePermissions = append(edgePermissions, GetEdgePermission(p.Stack, edgePermissions, nil)...)
	edgePermissions = append(edgePermissions, GetEdgePermission(p.Policy, edgePermissions, nil)...)
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
	edgePermissions = append(edgePermissions, GetEdgePermission(p.Policy, edgePermissions, &f)...)
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
	setRoleIdToPermission(p.Policy, roleId)
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
