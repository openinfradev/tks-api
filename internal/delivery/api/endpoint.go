package api

type Endpoint int
type EndpointInfo struct {
	Name  string
	Group string
}

// Comment below is special purpose for code generation.
// Do not edit this comment.
// Endpoint for Code Generation
const (
	// Auth
	Login Endpoint = iota
	Logout
	RefreshToken
	FindId
	FindPassword
	VerifyIdentityForLostId
	VerifyIdentityForLostPassword
	VerifyToken

	// User
	CreateUser
	ListUser
	GetUser
	DeleteUser
	UpdateUsers
	UpdateUser
	ResetPassword
	CheckId
	CheckEmail
	GetPermissionsByAccountId

	// MyProfile
	GetMyProfile
	UpdateMyProfile
	UpdateMyPassword
	RenewPasswordExpiredDate
	DeleteMyProfile

	// Organization
	Admin_CreateOrganization
	Admin_DeleteOrganization
	GetOrganizations
	GetOrganization
	CheckOrganizationName
	UpdateOrganization
	UpdatePrimaryCluster

	// Cluster
	CreateCluster
	GetClusters
	ImportCluster
	GetCluster
	DeleteCluster
	GetClusterSiteValues
	InstallCluster
	ResumeCluster
	CreateBootstrapKubeconfig
	GetBootstrapKubeconfig
	GetNodes

	//Appgroup
	CreateAppgroup
	GetAppgroups
	GetAppgroup
	DeleteAppgroup
	GetApplications
	CreateApplication

	// AppServeApp
	GetAppServeAppTasksByAppId
	GetAppServeAppTaskDetail
	CreateAppServeApp         // 프로젝트 관리/앱 서빙/배포 // 프로젝트 관리/앱 서빙/빌드
	GetAppServeApps           // 프로젝트 관리/앱 서빙/조회
	GetNumOfAppsOnStack       // 프로젝트 관리/앱 서빙/조회
	GetAppServeApp            // 프로젝트 관리/앱 서빙/조회
	GetAppServeAppLatestTask  // 프로젝트 관리/앱 서빙/조회
	GetAppServeAppLog         // 프로젝트 관리/앱 서빙/조회
	IsAppServeAppExist        // 프로젝트 관리/앱 서빙/조회 // 프로젝트 관리/앱 서빙/배포 // 프로젝트 관리/앱 서빙/빌드
	IsAppServeAppNameExist    // 프로젝트 관리/앱 서빙/조회 // 프로젝트 관리/앱 서빙/배포 // 프로젝트 관리/앱 서빙/빌드
	DeleteAppServeApp         // 프로젝트 관리/앱 서빙/삭제
	UpdateAppServeApp         // 프로젝트 관리/앱 서빙/배포 // 프로젝트 관리/앱 서빙/빌드
	UpdateAppServeAppStatus   // 프로젝트 관리/앱 서빙/배포 // 프로젝트 관리/앱 서빙/빌드
	UpdateAppServeAppEndpoint // 프로젝트 관리/앱 서빙/배포 // 프로젝트 관리/앱 서빙/빌드
	RollbackAppServeApp       // 프로젝트 관리/앱 서빙/배포 // 프로젝트 관리/앱 서빙/빌드

	// CloudAccount
	GetCloudAccounts
	CreateCloudAccount
	CheckCloudAccountName
	CheckAwsAccountId
	GetCloudAccount
	UpdateCloudAccount
	DeleteCloudAccount
	DeleteForceCloudAccount
	GetResourceQuota

	// StackTemplate
	Admin_GetStackTemplates
	Admin_GetStackTemplate
	Admin_GetStackTemplateServices
	Admin_GetStackTemplateTemplateIds
	Admin_CreateStackTemplate
	Admin_UpdateStackTemplate
	Admin_DeleteStackTemplate
	Admin_UpdateStackTemplateOrganizations
	Admin_CheckStackTemplateName
	GetOrganizationStackTemplates
	GetOrganizationStackTemplate
	AddOrganizationStackTemplates
	RemoveOrganizationStackTemplates
	GetOrganizationCloudServices

	// Dashboard
	CreateDashboard
	GetDashboard
	UpdateDashboard
	GetChartsDashboard    // 대시보드/대시보드/조회
	GetChartDashboard     // 대시보드/대시보드/조회
	GetStacksDashboard    // 대시보드/대시보드/조회
	GetResourcesDashboard // 대시보드/대시보드/조회
	GetPolicyStatusDashboard
	GetPolicyUpdateDashboard
	GetPolicyEnforcementDashboard
	GetPolicyViolationDashboard
	GetPolicyViolationLogDashboard
	GetPolicyStatisticsDashboard
	GetWorkloadDashboard
	GetPolicyViolationTop5Dashboard

	// SystemNotificationTemplate
	Admin_CreateSystemNotificationTemplate
	Admin_UpdateSystemNotificationTemplate
	Admin_DeleteSystemNotificationTemplate
	Admin_GetSystemNotificationTemplates
	Admin_GetSystemNotificationTemplate
	Admin_CheckSystemNotificationTemplateName
	GetOrganizationSystemNotificationTemplates
	GetOrganizationSystemNotificationTemplate
	AddOrganizationSystemNotificationTemplates
	RemoveOrganizationSystemNotificationTemplates

	// SystemNotificationRule
	CreateSystemNotificationRule
	GetSystemNotificationRules
	GetSystemNotificationRule
	CheckSystemNotificationRuleName
	DeleteSystemNotificationRule
	UpdateSystemNotificationRule
	MakeDefaultSystemNotificationRules

	// SystemNotification
	CreateSystemNotification
	GetSystemNotifications
	GetSystemNotification
	DeleteSystemNotification
	UpdateSystemNotification
	CreateSystemNotificationAction

	// PolicyNotification
	GetPolicyNotifications
	GetPolicyNotification

	// Stack
	GetStacks           // 스택관리/조회
	CreateStack         // 스택관리/생성
	ImportStack         // 스택관리/생성
	CheckStackName      // 스택관리/조회
	GetStack            // 스택관리/조회
	UpdateStack         // 스택관리/수정
	DeleteStack         // 스택관리/삭제
	GetStackKubeconfig  // 스택관리/조회
	GetStackStatus      // 스택관리/조회
	SetFavoriteStack    // 스택관리/조회
	DeleteFavoriteStack // 스택관리/조회
	InstallStack        // 스택관리/조회

	// Project
	CreateProject           // 프로젝트 관리/프로젝트/생성
	GetProjectRoles         // 프로젝트 관리/설정-일반/조회 // 프로젝트 관리/설정-멤버/조회
	GetProjectRole          // 프로젝트 관리/설정-일반/조회 // 프로젝트 관리/설정-멤버/조회
	GetProjects             // 프로젝트 관리/프로젝트/조회 // 프로젝트 관리/설정-일반/조회
	GetProject              // 프로젝트 관리/프로젝트/조회 // 프로젝트 관리/설정-일반/조회
	UpdateProject           // 프로젝트 관리/설정-일반/수정
	DeleteProject           // 프로젝트 관리/설정-일반/삭제
	AddProjectMember        // 프로젝트 관리/설정-멤버/생성
	GetProjectMember        // 프로젝트 관리/설정-멤버/조회
	GetProjectMembers       // 프로젝트 관리/설정-멤버/조회
	RemoveProjectMember     // 프로젝트 관리/설정-멤버/삭제
	UpdateProjectMemberRole // 프로젝트 관리/설정-멤버/수정
	CreateProjectNamespace  // 프로젝트 관리/설정-네임스페이스/생성
	GetProjectNamespaces    // 프로젝트 관리/설정-네임스페이스/조회
	GetProjectNamespace     // 프로젝트 관리/설정-네임스페이스/조회
	UpdateProjectNamespace
	DeleteProjectNamespace // 프로젝트 관리/설정-네임스페이스/삭제
	SetFavoriteProject
	SetFavoriteProjectNamespace
	UnSetFavoriteProject
	UnSetFavoriteProjectNamespace
	GetProjectKubeconfig
	GetProjectNamespaceK8sResources
	GetProjectNamespaceKubeconfig

	// Audit
	GetAudits
	GetAudit
	DeleteAudit

	// Role
	CreateTksRole
	ListTksRoles
	GetTksRole
	DeleteTksRole
	UpdateTksRole
	GetPermissionsByRoleId
	UpdatePermissionsByRoleId
	IsRoleNameExisted
	AppendUsersToRole
	GetUsersInRoleId
	RemoveUsersFromRole

	// Permission
	GetPermissionTemplates

	// Admin_User
	Admin_CreateUser
	Admin_ListUser
	Admin_GetUser
	Admin_DeleteUser
	Admin_UpdateUser

	// Admin Role
	Admin_ListTksRoles
	Admin_GetTksRole

	// Admin Project
	Admin_GetProjects

	// PolicyTemplate
	Admin_ListPolicyTemplate
	Admin_CreatePolicyTemplate
	Admin_DeletePolicyTemplate
	Admin_GetPolicyTemplate
	Admin_UpdatePolicyTemplate
	Admin_GetPolicyTemplateDeploy
	Admin_ListPolicyTemplateStatistics
	Admin_ListPolicyTemplateVersions
	Admin_CreatePolicyTemplateVersion
	Admin_DeletePolicyTemplateVersion
	Admin_GetPolicyTemplateVersion
	Admin_ExistsPolicyTemplateKind
	Admin_ExistsPolicyTemplateName
	Admin_ExtractParameters
	Admin_AddPermittedPolicyTemplatesForOrganization
	Admin_DeletePermittedPolicyTemplatesForOrganization

	// StackPolicyStatus
	ListStackPolicyStatus
	GetStackPolicyTemplateStatus
	UpdateStackPolicyTemplateStatus

	// Policy
	GetMandatoryPolicies
	SetMandatoryPolicies
	GetPolicyStatistics
	ListPolicy
	CreatePolicy
	DeletePolicy
	GetPolicy
	UpdatePolicy
	UpdatePolicyTargetClusters
	ExistsPolicyName
	ExistsPolicyResourceName
	GetPolicyEdit
	AddPoliciesForStack
	DeletePoliciesForStack
	StackPolicyStatistics

	// OrganizationPolicyTemplate
	ListPolicyTemplate
	CreatePolicyTemplate
	DeletePolicyTemplate
	GetPolicyTemplate
	UpdatePolicyTemplate
	GetPolicyTemplateDeploy
	ListPolicyTemplateStatistics
	ListPolicyTemplateVersions
	CreatePolicyTemplateVersion
	DeletePolicyTemplateVersion
	GetPolicyTemplateVersion
	ExistsPolicyTemplateKind
	ExistsPolicyTemplateName
	ExtractParameters

	// PolicyTemplateExample
	ListPolicyTemplateExample
	GetPolicyTemplateExample
	UpdatePolicyTemplateExample
	DeletePolicyTemplateExample

	// Utility
	CompileRego
)
