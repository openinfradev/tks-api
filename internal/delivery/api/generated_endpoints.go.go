// This is generated code. DO NOT EDIT.

package api

var MapWithEndpoint = map[Endpoint]EndpointInfo{
	Login: {
		Name:  "Login",
		Group: "Auth",
	},
	Logout: {
		Name:  "Logout",
		Group: "Auth",
	},
	RefreshToken: {
		Name:  "RefreshToken",
		Group: "Auth",
	},
	FindId: {
		Name:  "FindId",
		Group: "Auth",
	},
	FindPassword: {
		Name:  "FindPassword",
		Group: "Auth",
	},
	VerifyIdentityForLostId: {
		Name:  "VerifyIdentityForLostId",
		Group: "Auth",
	},
	VerifyIdentityForLostPassword: {
		Name:  "VerifyIdentityForLostPassword",
		Group: "Auth",
	},
	VerifyToken: {
		Name:  "VerifyToken",
		Group: "Auth",
	},
	CreateUser: {
		Name:  "CreateUser",
		Group: "User",
	},
	ListUser: {
		Name:  "ListUser",
		Group: "User",
	},
	GetUser: {
		Name:  "GetUser",
		Group: "User",
	},
	DeleteUser: {
		Name:  "DeleteUser",
		Group: "User",
	},
	UpdateUsers: {
		Name:  "UpdateUsers",
		Group: "User",
	},
	UpdateUser: {
		Name:  "UpdateUser",
		Group: "User",
	},
	ResetPassword: {
		Name:  "ResetPassword",
		Group: "User",
	},
	CheckId: {
		Name:  "CheckId",
		Group: "User",
	},
	CheckEmail: {
		Name:  "CheckEmail",
		Group: "User",
	},
	GetPermissionsByAccountId: {
		Name:  "GetPermissionsByAccountId",
		Group: "User",
	},
	GetMyProfile: {
		Name:  "GetMyProfile",
		Group: "MyProfile",
	},
	UpdateMyProfile: {
		Name:  "UpdateMyProfile",
		Group: "MyProfile",
	},
	UpdateMyPassword: {
		Name:  "UpdateMyPassword",
		Group: "MyProfile",
	},
	RenewPasswordExpiredDate: {
		Name:  "RenewPasswordExpiredDate",
		Group: "MyProfile",
	},
	DeleteMyProfile: {
		Name:  "DeleteMyProfile",
		Group: "MyProfile",
	},
	Admin_CreateOrganization: {
		Name:  "Admin_CreateOrganization",
		Group: "Organization",
	},
	Admin_DeleteOrganization: {
		Name:  "Admin_DeleteOrganization",
		Group: "Organization",
	},
	GetOrganizations: {
		Name:  "GetOrganizations",
		Group: "Organization",
	},
	GetOrganization: {
		Name:  "GetOrganization",
		Group: "Organization",
	},
	CheckOrganizationName: {
		Name:  "CheckOrganizationName",
		Group: "Organization",
	},
	UpdateOrganization: {
		Name:  "UpdateOrganization",
		Group: "Organization",
	},
	UpdatePrimaryCluster: {
		Name:  "UpdatePrimaryCluster",
		Group: "Organization",
	},
	CreateCluster: {
		Name:  "CreateCluster",
		Group: "Cluster",
	},
	GetClusters: {
		Name:  "GetClusters",
		Group: "Cluster",
	},
	ImportCluster: {
		Name:  "ImportCluster",
		Group: "Cluster",
	},
	GetCluster: {
		Name:  "GetCluster",
		Group: "Cluster",
	},
	DeleteCluster: {
		Name:  "DeleteCluster",
		Group: "Cluster",
	},
	GetClusterSiteValues: {
		Name:  "GetClusterSiteValues",
		Group: "Cluster",
	},
	InstallCluster: {
		Name:  "InstallCluster",
		Group: "Cluster",
	},
	CreateBootstrapKubeconfig: {
		Name:  "CreateBootstrapKubeconfig",
		Group: "Cluster",
	},
	GetBootstrapKubeconfig: {
		Name:  "GetBootstrapKubeconfig",
		Group: "Cluster",
	},
	GetNodes: {
		Name:  "GetNodes",
		Group: "Cluster",
	},
	CreateAppgroup: {
		Name:  "CreateAppgroup",
		Group: "Appgroup",
	},
	GetAppgroups: {
		Name:  "GetAppgroups",
		Group: "Appgroup",
	},
	GetAppgroup: {
		Name:  "GetAppgroup",
		Group: "Appgroup",
	},
	DeleteAppgroup: {
		Name:  "DeleteAppgroup",
		Group: "Appgroup",
	},
	GetApplications: {
		Name:  "GetApplications",
		Group: "Appgroup",
	},
	CreateApplication: {
		Name:  "CreateApplication",
		Group: "Appgroup",
	},
	GetAppServeAppTasksByAppId: {
		Name:  "GetAppServeAppTasksByAppId",
		Group: "AppServeApp",
	},
	GetAppServeAppTaskDetail: {
		Name:  "GetAppServeAppTaskDetail",
		Group: "AppServeApp",
	},
	CreateAppServeApp: {
		Name:  "CreateAppServeApp",
		Group: "AppServeApp",
	},
	GetAppServeApps: {
		Name:  "GetAppServeApps",
		Group: "AppServeApp",
	},
	GetNumOfAppsOnStack: {
		Name:  "GetNumOfAppsOnStack",
		Group: "AppServeApp",
	},
	GetAppServeApp: {
		Name:  "GetAppServeApp",
		Group: "AppServeApp",
	},
	GetAppServeAppLatestTask: {
		Name:  "GetAppServeAppLatestTask",
		Group: "AppServeApp",
	},
	IsAppServeAppExist: {
		Name:  "IsAppServeAppExist",
		Group: "AppServeApp",
	},
	IsAppServeAppNameExist: {
		Name:  "IsAppServeAppNameExist",
		Group: "AppServeApp",
	},
	DeleteAppServeApp: {
		Name:  "DeleteAppServeApp",
		Group: "AppServeApp",
	},
	UpdateAppServeApp: {
		Name:  "UpdateAppServeApp",
		Group: "AppServeApp",
	},
	UpdateAppServeAppStatus: {
		Name:  "UpdateAppServeAppStatus",
		Group: "AppServeApp",
	},
	UpdateAppServeAppEndpoint: {
		Name:  "UpdateAppServeAppEndpoint",
		Group: "AppServeApp",
	},
	RollbackAppServeApp: {
		Name:  "RollbackAppServeApp",
		Group: "AppServeApp",
	},
	GetCloudAccounts: {
		Name:  "GetCloudAccounts",
		Group: "CloudAccount",
	},
	CreateCloudAccount: {
		Name:  "CreateCloudAccount",
		Group: "CloudAccount",
	},
	CheckCloudAccountName: {
		Name:  "CheckCloudAccountName",
		Group: "CloudAccount",
	},
	CheckAwsAccountId: {
		Name:  "CheckAwsAccountId",
		Group: "CloudAccount",
	},
	GetCloudAccount: {
		Name:  "GetCloudAccount",
		Group: "CloudAccount",
	},
	UpdateCloudAccount: {
		Name:  "UpdateCloudAccount",
		Group: "CloudAccount",
	},
	DeleteCloudAccount: {
		Name:  "DeleteCloudAccount",
		Group: "CloudAccount",
	},
	DeleteForceCloudAccount: {
		Name:  "DeleteForceCloudAccount",
		Group: "CloudAccount",
	},
	GetResourceQuota: {
		Name:  "GetResourceQuota",
		Group: "CloudAccount",
	},
	Admin_GetStackTemplates: {
		Name:  "Admin_GetStackTemplates",
		Group: "StackTemplate",
	},
	Admin_GetStackTemplate: {
		Name:  "Admin_GetStackTemplate",
		Group: "StackTemplate",
	},
	Admin_GetStackTemplateServices: {
		Name:  "Admin_GetStackTemplateServices",
		Group: "StackTemplate",
	},
	Admin_GetStackTemplateTemplateIds: {
		Name:  "Admin_GetStackTemplateTemplateIds",
		Group: "StackTemplate",
	},
	Admin_CreateStackTemplate: {
		Name:  "Admin_CreateStackTemplate",
		Group: "StackTemplate",
	},
	Admin_UpdateStackTemplate: {
		Name:  "Admin_UpdateStackTemplate",
		Group: "StackTemplate",
	},
	Admin_DeleteStackTemplate: {
		Name:  "Admin_DeleteStackTemplate",
		Group: "StackTemplate",
	},
	Admin_UpdateStackTemplateOrganizations: {
		Name:  "Admin_UpdateStackTemplateOrganizations",
		Group: "StackTemplate",
	},
	Admin_CheckStackTemplateName: {
		Name:  "Admin_CheckStackTemplateName",
		Group: "StackTemplate",
	},
	GetOrganizationStackTemplates: {
		Name:  "GetOrganizationStackTemplates",
		Group: "StackTemplate",
	},
	GetOrganizationStackTemplate: {
		Name:  "GetOrganizationStackTemplate",
		Group: "StackTemplate",
	},
	AddOrganizationStackTemplates: {
		Name:  "AddOrganizationStackTemplates",
		Group: "StackTemplate",
	},
	RemoveOrganizationStackTemplates: {
		Name:  "RemoveOrganizationStackTemplates",
		Group: "StackTemplate",
	},
	CreateDashboard: {
		Name:  "CreateDashboard",
		Group: "Dashboard",
	},
	GetDashboard: {
		Name:  "GetDashboard",
		Group: "Dashboard",
	},
	UpdateDashboard: {
		Name:  "UpdateDashboard",
		Group: "Dashboard",
	},
	GetChartsDashboard: {
		Name:  "GetChartsDashboard",
		Group: "Dashboard",
	},
	GetChartDashboard: {
		Name:  "GetChartDashboard",
		Group: "Dashboard",
	},
	GetStacksDashboard: {
		Name:  "GetStacksDashboard",
		Group: "Dashboard",
	},
	GetResourcesDashboard: {
		Name:  "GetResourcesDashboard",
		Group: "Dashboard",
	},
	GetPolicyStatusDashboard: {
		Name:  "GetPolicyStatusDashboard",
		Group: "Dashboard",
	},
	GetPolicyUpdateDashboard: {
		Name:  "GetPolicyUpdateDashboard",
		Group: "Dashboard",
	},
	GetPolicyEnforcementDashboard: {
		Name:  "GetPolicyEnforcementDashboard",
		Group: "Dashboard",
	},
	GetPolicyViolationDashboard: {
		Name:  "GetPolicyViolationDashboard",
		Group: "Dashboard",
	},
	GetPolicyViolationLogDashboard: {
		Name:  "GetPolicyViolationLogDashboard",
		Group: "Dashboard",
	},
	GetPolicyStatisticsDashboard: {
		Name:  "GetPolicyStatisticsDashboard",
		Group: "Dashboard",
	},
	GetWorkloadDashboard: {
		Name:  "GetWorkloadDashboard",
		Group: "Dashboard",
	},
	GetPolicyViolationTop5Dashboard: {
		Name:  "GetPolicyViolationTop5Dashboard",
		Group: "Dashboard",
	},
	Admin_CreateSystemNotificationTemplate: {
		Name:  "Admin_CreateSystemNotificationTemplate",
		Group: "SystemNotificationTemplate",
	},
	Admin_UpdateSystemNotificationTemplate: {
		Name:  "Admin_UpdateSystemNotificationTemplate",
		Group: "SystemNotificationTemplate",
	},
	Admin_DeleteSystemNotificationTemplate: {
		Name:  "Admin_DeleteSystemNotificationTemplate",
		Group: "SystemNotificationTemplate",
	},
	Admin_GetSystemNotificationTemplates: {
		Name:  "Admin_GetSystemNotificationTemplates",
		Group: "SystemNotificationTemplate",
	},
	Admin_GetSystemNotificationTemplate: {
		Name:  "Admin_GetSystemNotificationTemplate",
		Group: "SystemNotificationTemplate",
	},
	Admin_CheckSystemNotificationTemplateName: {
		Name:  "Admin_CheckSystemNotificationTemplateName",
		Group: "SystemNotificationTemplate",
	},
	GetOrganizationSystemNotificationTemplates: {
		Name:  "GetOrganizationSystemNotificationTemplates",
		Group: "SystemNotificationTemplate",
	},
	GetOrganizationSystemNotificationTemplate: {
		Name:  "GetOrganizationSystemNotificationTemplate",
		Group: "SystemNotificationTemplate",
	},
	AddOrganizationSystemNotificationTemplates: {
		Name:  "AddOrganizationSystemNotificationTemplates",
		Group: "SystemNotificationTemplate",
	},
	RemoveOrganizationSystemNotificationTemplates: {
		Name:  "RemoveOrganizationSystemNotificationTemplates",
		Group: "SystemNotificationTemplate",
	},
	CreateSystemNotificationRule: {
		Name:  "CreateSystemNotificationRule",
		Group: "SystemNotificationRule",
	},
	GetSystemNotificationRules: {
		Name:  "GetSystemNotificationRules",
		Group: "SystemNotificationRule",
	},
	GetSystemNotificationRule: {
		Name:  "GetSystemNotificationRule",
		Group: "SystemNotificationRule",
	},
	CheckSystemNotificationRuleName: {
		Name:  "CheckSystemNotificationRuleName",
		Group: "SystemNotificationRule",
	},
	DeleteSystemNotificationRule: {
		Name:  "DeleteSystemNotificationRule",
		Group: "SystemNotificationRule",
	},
	UpdateSystemNotificationRule: {
		Name:  "UpdateSystemNotificationRule",
		Group: "SystemNotificationRule",
	},
	MakeDefaultSystemNotificationRules: {
		Name:  "MakeDefaultSystemNotificationRules",
		Group: "SystemNotificationRule",
	},
	CreateSystemNotification: {
		Name:  "CreateSystemNotification",
		Group: "SystemNotification",
	},
	GetSystemNotifications: {
		Name:  "GetSystemNotifications",
		Group: "SystemNotification",
	},
	GetSystemNotification: {
		Name:  "GetSystemNotification",
		Group: "SystemNotification",
	},
	DeleteSystemNotification: {
		Name:  "DeleteSystemNotification",
		Group: "SystemNotification",
	},
	UpdateSystemNotification: {
		Name:  "UpdateSystemNotification",
		Group: "SystemNotification",
	},
	CreateSystemNotificationAction: {
		Name:  "CreateSystemNotificationAction",
		Group: "SystemNotification",
	},
	GetPolicyNotifications: {
		Name:  "GetPolicyNotifications",
		Group: "PolicyNotification",
	},
	GetPolicyNotification: {
		Name:  "GetPolicyNotification",
		Group: "PolicyNotification",
	},
	CreateStack: {
		Name:  "CreateStack",
		Group: "Stack",
	},
	GetStacks: {
		Name:  "GetStacks",
		Group: "Stack",
	},
	GetStack: {
		Name:  "GetStack",
		Group: "Stack",
	},
	UpdateStack: {
		Name:  "UpdateStack",
		Group: "Stack",
	},
	DeleteStack: {
		Name:  "DeleteStack",
		Group: "Stack",
	},
	CheckStackName: {
		Name:  "CheckStackName",
		Group: "Stack",
	},
	GetStackStatus: {
		Name:  "GetStackStatus",
		Group: "Stack",
	},
	GetStackKubeConfig: {
		Name:  "GetStackKubeConfig",
		Group: "Stack",
	},
	SetFavoriteStack: {
		Name:  "SetFavoriteStack",
		Group: "Stack",
	},
	DeleteFavoriteStack: {
		Name:  "DeleteFavoriteStack",
		Group: "Stack",
	},
	InstallStack: {
		Name:  "InstallStack",
		Group: "Stack",
	},
	CreateProject: {
		Name:  "CreateProject",
		Group: "Project",
	},
	GetProjectRoles: {
		Name:  "GetProjectRoles",
		Group: "Project",
	},
	GetProjectRole: {
		Name:  "GetProjectRole",
		Group: "Project",
	},
	GetProjects: {
		Name:  "GetProjects",
		Group: "Project",
	},
	GetProject: {
		Name:  "GetProject",
		Group: "Project",
	},
	UpdateProject: {
		Name:  "UpdateProject",
		Group: "Project",
	},
	DeleteProject: {
		Name:  "DeleteProject",
		Group: "Project",
	},
	AddProjectMember: {
		Name:  "AddProjectMember",
		Group: "Project",
	},
	GetProjectMember: {
		Name:  "GetProjectMember",
		Group: "Project",
	},
	GetProjectMembers: {
		Name:  "GetProjectMembers",
		Group: "Project",
	},
	RemoveProjectMember: {
		Name:  "RemoveProjectMember",
		Group: "Project",
	},
	UpdateProjectMemberRole: {
		Name:  "UpdateProjectMemberRole",
		Group: "Project",
	},
	CreateProjectNamespace: {
		Name:  "CreateProjectNamespace",
		Group: "Project",
	},
	GetProjectNamespaces: {
		Name:  "GetProjectNamespaces",
		Group: "Project",
	},
	GetProjectNamespace: {
		Name:  "GetProjectNamespace",
		Group: "Project",
	},
	UpdateProjectNamespace: {
		Name:  "UpdateProjectNamespace",
		Group: "Project",
	},
	DeleteProjectNamespace: {
		Name:  "DeleteProjectNamespace",
		Group: "Project",
	},
	SetFavoriteProject: {
		Name:  "SetFavoriteProject",
		Group: "Project",
	},
	SetFavoriteProjectNamespace: {
		Name:  "SetFavoriteProjectNamespace",
		Group: "Project",
	},
	UnSetFavoriteProject: {
		Name:  "UnSetFavoriteProject",
		Group: "Project",
	},
	UnSetFavoriteProjectNamespace: {
		Name:  "UnSetFavoriteProjectNamespace",
		Group: "Project",
	},
	GetProjectKubeconfig: {
		Name:  "GetProjectKubeconfig",
		Group: "Project",
	},
	GetProjectNamespaceK8sResources: {
		Name:  "GetProjectNamespaceK8sResources",
		Group: "Project",
	},
	GetProjectNamespaceKubeconfig: {
		Name:  "GetProjectNamespaceKubeconfig",
		Group: "Project",
	},
	GetAudits: {
		Name:  "GetAudits",
		Group: "Audit",
	},
	GetAudit: {
		Name:  "GetAudit",
		Group: "Audit",
	},
	DeleteAudit: {
		Name:  "DeleteAudit",
		Group: "Audit",
	},
	CreateTksRole: {
		Name:  "CreateTksRole",
		Group: "Role",
	},
	ListTksRoles: {
		Name:  "ListTksRoles",
		Group: "Role",
	},
	GetTksRole: {
		Name:  "GetTksRole",
		Group: "Role",
	},
	DeleteTksRole: {
		Name:  "DeleteTksRole",
		Group: "Role",
	},
	UpdateTksRole: {
		Name:  "UpdateTksRole",
		Group: "Role",
	},
	GetPermissionsByRoleId: {
		Name:  "GetPermissionsByRoleId",
		Group: "Role",
	},
	UpdatePermissionsByRoleId: {
		Name:  "UpdatePermissionsByRoleId",
		Group: "Role",
	},
	IsRoleNameExisted: {
		Name:  "IsRoleNameExisted",
		Group: "Role",
	},
	AppendUsersToRole: {
		Name:  "AppendUsersToRole",
		Group: "Role",
	},
	GetUsersInRoleId: {
		Name:  "GetUsersInRoleId",
		Group: "Role",
	},
	RemoveUsersFromRole: {
		Name:  "RemoveUsersFromRole",
		Group: "Role",
	},
	GetPermissionTemplates: {
		Name:  "GetPermissionTemplates",
		Group: "Permission",
	},
	Admin_GetEndpoints: {
		Name:  "Admin_GetEndpoints",
		Group: "Endpoint",
	},
	Admin_CreateUser: {
		Name:  "Admin_CreateUser",
		Group: "Admin_User",
	},
	Admin_ListUser: {
		Name:  "Admin_ListUser",
		Group: "Admin_User",
	},
	Admin_GetUser: {
		Name:  "Admin_GetUser",
		Group: "Admin_User",
	},
	Admin_DeleteUser: {
		Name:  "Admin_DeleteUser",
		Group: "Admin_User",
	},
	Admin_UpdateUser: {
		Name:  "Admin_UpdateUser",
		Group: "Admin_User",
	},
	Admin_ListTksRoles: {
		Name:  "Admin_ListTksRoles",
		Group: "Admin Role",
	},
	Admin_GetTksRole: {
		Name:  "Admin_GetTksRole",
		Group: "Admin Role",
	},
	Admin_GetProjects: {
		Name:  "Admin_GetProjects",
		Group: "Admin Project",
	},
	Admin_ListPolicyTemplate: {
		Name:  "Admin_ListPolicyTemplate",
		Group: "PolicyTemplate",
	},
	Admin_CreatePolicyTemplate: {
		Name:  "Admin_CreatePolicyTemplate",
		Group: "PolicyTemplate",
	},
	Admin_DeletePolicyTemplate: {
		Name:  "Admin_DeletePolicyTemplate",
		Group: "PolicyTemplate",
	},
	Admin_GetPolicyTemplate: {
		Name:  "Admin_GetPolicyTemplate",
		Group: "PolicyTemplate",
	},
	Admin_UpdatePolicyTemplate: {
		Name:  "Admin_UpdatePolicyTemplate",
		Group: "PolicyTemplate",
	},
	Admin_GetPolicyTemplateDeploy: {
		Name:  "Admin_GetPolicyTemplateDeploy",
		Group: "PolicyTemplate",
	},
	Admin_ListPolicyTemplateStatistics: {
		Name:  "Admin_ListPolicyTemplateStatistics",
		Group: "PolicyTemplate",
	},
	Admin_ListPolicyTemplateVersions: {
		Name:  "Admin_ListPolicyTemplateVersions",
		Group: "PolicyTemplate",
	},
	Admin_CreatePolicyTemplateVersion: {
		Name:  "Admin_CreatePolicyTemplateVersion",
		Group: "PolicyTemplate",
	},
	Admin_DeletePolicyTemplateVersion: {
		Name:  "Admin_DeletePolicyTemplateVersion",
		Group: "PolicyTemplate",
	},
	Admin_GetPolicyTemplateVersion: {
		Name:  "Admin_GetPolicyTemplateVersion",
		Group: "PolicyTemplate",
	},
	Admin_ExistsPolicyTemplateKind: {
		Name:  "Admin_ExistsPolicyTemplateKind",
		Group: "PolicyTemplate",
	},
	Admin_ExistsPolicyTemplateName: {
		Name:  "Admin_ExistsPolicyTemplateName",
		Group: "PolicyTemplate",
	},
	Admin_ExtractParameters: {
		Name:  "Admin_ExtractParameters",
		Group: "PolicyTemplate",
	},
	Admin_AddPermittedPolicyTemplatesForOrganization: {
		Name:  "Admin_AddPermittedPolicyTemplatesForOrganization",
		Group: "PolicyTemplate",
	},
	Admin_DeletePermittedPolicyTemplatesForOrganization: {
		Name:  "Admin_DeletePermittedPolicyTemplatesForOrganization",
		Group: "PolicyTemplate",
	},
	ListStackPolicyStatus: {
		Name:  "ListStackPolicyStatus",
		Group: "StackPolicyStatus",
	},
	GetStackPolicyTemplateStatus: {
		Name:  "GetStackPolicyTemplateStatus",
		Group: "StackPolicyStatus",
	},
	UpdateStackPolicyTemplateStatus: {
		Name:  "UpdateStackPolicyTemplateStatus",
		Group: "StackPolicyStatus",
	},
	GetMandatoryPolicies: {
		Name:  "GetMandatoryPolicies",
		Group: "Policy",
	},
	SetMandatoryPolicies: {
		Name:  "SetMandatoryPolicies",
		Group: "Policy",
	},
	GetPolicyStatistics: {
		Name:  "GetPolicyStatistics",
		Group: "Policy",
	},
	ListPolicy: {
		Name:  "ListPolicy",
		Group: "Policy",
	},
	CreatePolicy: {
		Name:  "CreatePolicy",
		Group: "Policy",
	},
	DeletePolicy: {
		Name:  "DeletePolicy",
		Group: "Policy",
	},
	GetPolicy: {
		Name:  "GetPolicy",
		Group: "Policy",
	},
	UpdatePolicy: {
		Name:  "UpdatePolicy",
		Group: "Policy",
	},
	UpdatePolicyTargetClusters: {
		Name:  "UpdatePolicyTargetClusters",
		Group: "Policy",
	},
	ExistsPolicyName: {
		Name:  "ExistsPolicyName",
		Group: "Policy",
	},
	ExistsPolicyResourceName: {
		Name:  "ExistsPolicyResourceName",
		Group: "Policy",
	},
	GetPolicyEdit: {
		Name:  "GetPolicyEdit",
		Group: "Policy",
	},
	AddPoliciesForStack: {
		Name:  "AddPoliciesForStack",
		Group: "Policy",
	},
	DeletePoliciesForStack: {
		Name:  "DeletePoliciesForStack",
		Group: "Policy",
	},
	StackPolicyStatistics: {
		Name:  "StackPolicyStatistics",
		Group: "Policy",
	},
	ListPolicyTemplate: {
		Name:  "ListPolicyTemplate",
		Group: "OrganizationPolicyTemplate",
	},
	CreatePolicyTemplate: {
		Name:  "CreatePolicyTemplate",
		Group: "OrganizationPolicyTemplate",
	},
	DeletePolicyTemplate: {
		Name:  "DeletePolicyTemplate",
		Group: "OrganizationPolicyTemplate",
	},
	GetPolicyTemplate: {
		Name:  "GetPolicyTemplate",
		Group: "OrganizationPolicyTemplate",
	},
	UpdatePolicyTemplate: {
		Name:  "UpdatePolicyTemplate",
		Group: "OrganizationPolicyTemplate",
	},
	GetPolicyTemplateDeploy: {
		Name:  "GetPolicyTemplateDeploy",
		Group: "OrganizationPolicyTemplate",
	},
	ListPolicyTemplateStatistics: {
		Name:  "ListPolicyTemplateStatistics",
		Group: "OrganizationPolicyTemplate",
	},
	ListPolicyTemplateVersions: {
		Name:  "ListPolicyTemplateVersions",
		Group: "OrganizationPolicyTemplate",
	},
	CreatePolicyTemplateVersion: {
		Name:  "CreatePolicyTemplateVersion",
		Group: "OrganizationPolicyTemplate",
	},
	DeletePolicyTemplateVersion: {
		Name:  "DeletePolicyTemplateVersion",
		Group: "OrganizationPolicyTemplate",
	},
	GetPolicyTemplateVersion: {
		Name:  "GetPolicyTemplateVersion",
		Group: "OrganizationPolicyTemplate",
	},
	ExistsPolicyTemplateKind: {
		Name:  "ExistsPolicyTemplateKind",
		Group: "OrganizationPolicyTemplate",
	},
	ExistsPolicyTemplateName: {
		Name:  "ExistsPolicyTemplateName",
		Group: "OrganizationPolicyTemplate",
	},
	ExtractParameters: {
		Name:  "ExtractParameters",
		Group: "OrganizationPolicyTemplate",
	},
	ListPolicyTemplateExample: {
		Name:  "ListPolicyTemplateExample",
		Group: "PolicyTemplateExample",
	},
	GetPolicyTemplateExample: {
		Name:  "GetPolicyTemplateExample",
		Group: "PolicyTemplateExample",
	},
	UpdatePolicyTemplateExample: {
		Name:  "UpdatePolicyTemplateExample",
		Group: "PolicyTemplateExample",
	},
	DeletePolicyTemplateExample: {
		Name:  "DeletePolicyTemplateExample",
		Group: "PolicyTemplateExample",
	},
	CompileRego: {
		Name:  "CompileRego",
		Group: "Utility",
	},
}
var MapWithName = reverseApiMap()

func reverseApiMap() map[string]Endpoint {
	m := make(map[string]Endpoint)
	for k, v := range MapWithEndpoint {
		m[v.Name] = k
	}
	return m
}

func (e Endpoint) String() string {
	return MapWithEndpoint[e].Name
}

func GetEndpoint(name string) Endpoint {
	return MapWithName[name]
}
