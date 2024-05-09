package model

import (
	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/delivery/api"
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
)

type Permission struct {
	gorm.Model

	ID   uuid.UUID `gorm:"primarykey;type:uuid;" json:"ID"`
	Name string    `json:"name"`
	Key  string    `gorm:"type:text;" json:"key,omitempty"`

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
	Policy            *Permission `gorm:"-:all" json:"policy,omitempty"`
	ProjectManagement *Permission `gorm:"-:all" json:"project_management,omitempty"`
	Notification      *Permission `gorm:"-:all" json:"notification,omitempty"`
	Configuration     *Permission `gorm:"-:all" json:"configuration,omitempty"`
	Common            *Permission `gorm:"-:all" json:"common,omitempty"`
	Admin             *Permission `gorm:"-:all" json:"admin,omitempty"`
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
		Admin:             nil,
	}
}

func NewAdminPermissionSet() *PermissionSet {
	return &PermissionSet{
		Admin:             newAdmin(),
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
		if f != nil && !(*f)(*root) {
			return edgePermissions
		}
		return append(edgePermissions, root)
	}

	for _, child := range root.Children {
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
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.GetChartsDashboard,
							api.GetChartDashboard,
							api.GetStacksDashboard,
							api.GetResourcesDashboard,
						),
					},
					{
						ID:        uuid.New(),
						Name:      "수정",
						Key:       OperationUpdate,
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
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.GetStacks,
							api.GetStack,
							api.CheckStackName,
							api.GetStackStatus,
							api.GetStackKubeConfig,

							api.SetFavoriteStack,
							api.DeleteFavoriteStack,

							// Cluster
							api.GetCluster,
							api.GetClusters,
							api.GetClusterSiteValues,
							api.GetBootstrapKubeconfig,
							api.GetNodes,

							// AppGroup
							api.GetAppgroups,
							api.GetAppgroup,
							api.GetApplications,
						),
					},
					{
						ID:        uuid.New(),
						Name:      "생성",
						Key:       OperationCreate,
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.CreateStack,
							api.InstallStack,
							api.CreateAppgroup,

							// Cluster
							api.CreateCluster,
							api.ImportCluster,
							api.InstallCluster,
							api.CreateBootstrapKubeconfig,

							// AppGroup
							api.CreateAppgroup,
							api.CreateApplication,
						),
					},
					{
						ID:        uuid.New(),
						Name:      "수정",
						Key:       OperationUpdate,
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.UpdateStack,
						),
					},
					{
						ID:        uuid.New(),
						Name:      "삭제",
						Key:       OperationDelete,
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.DeleteStack,

							// Cluster
							api.DeleteCluster,

							// AppGroup
							api.DeleteAppgroup,
						),
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
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							// PolicyTemplate
							api.Admin_ListPolicyTemplate,
							api.Admin_GetPolicyTemplate,
							api.Admin_GetPolicyTemplateDeploy,
							api.Admin_ListPolicyTemplateStatistics,
							api.Admin_ListPolicyTemplateVersions,
							api.Admin_GetPolicyTemplateVersion,
							api.Admin_ExistsPolicyTemplateName,
							api.Admin_ExistsPolicyTemplateKind,

							// StackPolicyStatus
							api.ListStackPolicyStatus,
							api.GetStackPolicyTemplateStatus,

							// Policy
							api.GetMandatoryPolicies,
							api.ListPolicy,
							api.GetPolicy,
							api.ExistsPolicyName,

							// OrganizationPolicyTemplate
							api.ListPolicyTemplate,
							api.GetPolicyTemplate,
							api.GetPolicyTemplateDeploy,
							api.ListPolicyTemplateStatistics,
							api.ListPolicyTemplateVersions,
							api.GetPolicyTemplateVersion,
							api.ExistsPolicyTemplateKind,
							api.ExistsPolicyTemplateName,

							// PolicyTemplateExample
							api.ListPolicyTemplateExample,
							api.GetPolicyTemplateExample,
						),
					},
					{
						ID:        uuid.New(),
						Name:      "생성",
						Key:       OperationCreate,
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							// PolicyTemplate
							api.Admin_CreatePolicyTemplate,
							api.Admin_CreatePolicyTemplateVersion,

							// Policy
							api.SetMandatoryPolicies,
							api.CreatePolicy,

							// OrganizationPolicyTemplate
							api.CreatePolicyTemplate,
							api.CreatePolicyTemplateVersion,
						),
					},
					{
						ID:        uuid.New(),
						Name:      "수정",
						Key:       OperationUpdate,
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							// PolicyTemplate
							api.Admin_UpdatePolicyTemplate,

							// ClusterPolicyStatus
							api.UpdateStackPolicyTemplateStatus,

							// Policy
							api.UpdatePolicy,
							api.UpdatePolicyTargetClusters,

							// OrganizationPolicyTemplate
							api.UpdatePolicyTemplate,

							// PolicyTemplateExample
							api.UpdatePolicyTemplateExample,
						),
					},
					{
						ID:        uuid.New(),
						Name:      "삭제",
						Key:       OperationDelete,
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							// PolicyTemplate
							api.Admin_DeletePolicyTemplate,
							api.Admin_DeletePolicyTemplateVersion,

							// Policy
							api.DeletePolicy,

							// OrganizationPolicyTemplate
							api.DeletePolicyTemplate,
							api.DeletePolicyTemplateVersion,

							// PolicyTemplateExample
							api.DeletePolicyTemplateExample,
						),
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
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.GetSystemNotification,
							api.GetSystemNotifications,
						),
					},
					{
						ID:        uuid.New(),
						Name:      "수정",
						Key:       OperationUpdate,
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.UpdateSystemNotification,
							api.CreateSystemNotificationAction,
						),
					},
					{
						ID:        uuid.New(),
						Name:      "다운로드",
						Key:       OperationDownload,
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
						IsAllowed: helper.BoolP(false),
						Children:  []*Permission{},
					},
					{
						ID:        uuid.New(),
						Name:      "다운로드",
						Key:       OperationDownload,
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
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.GetProjects,
							api.GetProject,
							api.GetProjectKubeconfig,
						),
					},
					{
						ID:        uuid.New(),
						Name:      "생성",
						Key:       OperationCreate,
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.CreateProject,
						),
					},
					{
						ID:        uuid.New(),
						Name:      "수정",
						Key:       OperationUpdate,
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.UpdateProject,
						),
					},
					{
						ID:        uuid.New(),
						Name:      "삭제",
						Key:       OperationDelete,
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.DeleteProject,
						),
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
						Key:       OperationUpdate,
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.UpdateProject,
						),
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
						Key:       OperationCreate,
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.AddProjectMember,
						),
					},
					{
						ID:        uuid.New(),
						Name:      "수정",
						Key:       OperationUpdate,
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.UpdateProjectMemberRole,
						),
					},
					{
						ID:        uuid.New(),
						Name:      "삭제",
						Key:       OperationDelete,
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.RemoveProjectMember,
						),
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
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.GetProjectNamespaces,
							api.GetProjectNamespace,
							api.GetProjectNamespaceK8sResources,
						),
					},
					{
						ID:        uuid.New(),
						Name:      "생성",
						Key:       OperationCreate,
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.CreateProjectNamespace,
						),
					},
					{
						ID:        uuid.New(),
						Name:      "수정",
						Key:       OperationUpdate,
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.UpdateProjectNamespace,
						),
					},
					{
						ID:        uuid.New(),
						Name:      "삭제",
						Key:       OperationDelete,
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.DeleteProjectNamespace,
						),
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
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.GetAppServeApps,
							api.GetAppServeApp,
							api.GetNumOfAppsOnStack,
							api.GetAppServeAppLatestTask,
							api.IsAppServeAppExist,
							api.IsAppServeAppNameExist,
							api.GetAppServeAppTaskDetail,
							api.GetAppServeAppTasksByAppId,
						),
					},
					{
						ID:        uuid.New(),
						Name:      "생성",
						Key:       OperationCreate,
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
						Name:      "수정",
						Key:       OperationUpdate,
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
						Key:       OperationDelete,
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.DeleteAppServeApp,
						),
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
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "수정",
						Key:       OperationUpdate,
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
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.GetCloudAccounts,
							api.GetCloudAccount,
							api.CheckCloudAccountName,
							api.CheckAwsAccountId,
							api.GetResourceQuota,
						),
					},
					{
						ID:        uuid.New(),
						Name:      "생성",
						Key:       OperationCreate,
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.CreateCloudAccount,
						),
					},
					{
						ID:        uuid.New(),
						Name:      "수정",
						Key:       OperationUpdate,
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.UpdateCloudAccount,
						),
					},
					{
						ID:        uuid.New(),
						Name:      "삭제",
						Key:       OperationDelete,
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.DeleteCloudAccount,
							api.DeleteForceCloudAccount,
						),
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
						IsAllowed: helper.BoolP(false),
					},
					{
						ID:        uuid.New(),
						Name:      "생성",
						Key:       OperationCreate,
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
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.ListUser,
							api.GetUser,
							api.CheckId,
							api.CheckEmail,
						),
					},
					{
						ID:        uuid.New(),
						Name:      "생성",
						Key:       OperationCreate,
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.CreateUser,
							api.CheckId,
							api.CheckEmail,
						),
					},
					{
						ID:        uuid.New(),
						Name:      "수정",
						Key:       OperationUpdate,
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.UpdateUser,
							api.ResetPassword,
						),
					},
					{
						ID:        uuid.New(),
						Name:      "삭제",
						Key:       OperationDelete,
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.DeleteUser,
						),
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
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.ListTksRoles,
							api.GetTksRole,
							api.GetPermissionsByRoleId,
							api.GetPermissionTemplates,
						),
					},
					{
						ID:        uuid.New(),
						Name:      "생성",
						Key:       OperationCreate,
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.CreateTksRole,
						),
					},
					{
						ID:        uuid.New(),
						Name:      "수정",
						Key:       OperationUpdate,
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.UpdateTksRole,
							api.UpdatePermissionsByRoleId,
						),
					},
					{
						ID:        uuid.New(),
						Name:      "삭제",
						Key:       OperationDelete,
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.DeleteTksRole,
						),
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
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.GetSystemNotificationRules,
							api.GetSystemNotificationRule,
						),
					},
					{
						ID:        uuid.New(),
						Name:      "생성",
						Key:       OperationCreate,
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.CreateSystemNotificationRule,
						),
					},
					{
						ID:        uuid.New(),
						Name:      "수정",
						Key:       OperationUpdate,
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.UpdateSystemNotificationRule,
						),
					},
					{
						ID:        uuid.New(),
						Name:      "삭제",
						Key:       OperationDelete,
						IsAllowed: helper.BoolP(false),
						Endpoints: endpointObjects(
							api.DeleteSystemNotificationRule,
						),
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
		Endpoints: endpointObjects(
			// Auth
			api.Login,
			api.Logout,
			api.RefreshToken,
			api.FindId,
			api.FindPassword,
			api.VerifyIdentityForLostId,
			api.VerifyIdentityForLostPassword,
			api.VerifyToken,

			// Stack
			api.SetFavoriteStack,
			api.DeleteFavoriteStack,

			// Project
			api.SetFavoriteProject,
			api.SetFavoriteProjectNamespace,
			api.UnSetFavoriteProject,
			api.UnSetFavoriteProjectNamespace,

			// MyProfile
			api.GetMyProfile,
			api.UpdateMyProfile,
			api.UpdateMyPassword,
			api.RenewPasswordExpiredDate,
			api.DeleteMyProfile,

			// StackTemplate
			api.GetOrganizationStackTemplates,
			api.GetOrganizationStackTemplate,

			// Utiliy
			api.CompileRego,
		),
	}

	return common

}

func newAdmin() *Permission {
	admin := &Permission{
		ID:        uuid.New(),
		Name:      "관리자",
		IsAllowed: helper.BoolP(true),
		Endpoints: endpointObjects(
			// Organization
			api.Admin_CreateOrganization,
			api.Admin_DeleteOrganization,
			api.UpdateOrganization,
			api.GetOrganization,
			api.GetOrganizations,
			api.UpdatePrimaryCluster,
			api.CheckOrganizationName,

			// User
			api.ResetPassword,
			api.CheckId,
			api.CheckEmail,

			// StackTemplate
			api.Admin_GetStackTemplates,
			api.Admin_GetStackTemplate,
			api.Admin_GetStackTemplateServices,
			api.Admin_CreateStackTemplate,
			api.Admin_UpdateStackTemplate,
			api.Admin_DeleteStackTemplate,
			api.Admin_UpdateStackTemplateOrganizations,
			api.Admin_CheckStackTemplateName,

			// Admin
			api.Admin_GetUser,
			api.Admin_ListUser,
			api.Admin_CreateUser,
			api.Admin_UpdateUser,
			api.Admin_DeleteUser,
			api.Admin_GetSystemNotificationTemplate,
			api.Admin_CreateSystemNotificationTemplate,
			api.Admin_ListUser,
			api.Admin_GetTksRole,
			api.Admin_GetProjects,
			api.Admin_UpdateSystemNotificationTemplate,
			api.Admin_ListTksRoles,
			api.Admin_GetSystemNotificationTemplates,

			// Audit
			api.GetAudits,
			api.GetAudit,
			api.DeleteAudit,

			api.CreateSystemNotification,
			api.DeleteSystemNotification,
		),
	}

	return admin
}

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
