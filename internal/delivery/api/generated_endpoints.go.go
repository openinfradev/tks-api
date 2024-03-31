 // This is generated code. DO NOT EDIT.

package api

var ApiMap = map[Endpoint]EndpointInfo{
    Login: {
		Name: "Login", 
		Group: "Auth",
	},
    Logout: {
		Name: "Logout", 
		Group: "Auth",
	},
    RefreshToken: {
		Name: "RefreshToken", 
		Group: "Auth",
	},
    FindId: {
		Name: "FindId", 
		Group: "Auth",
	},
    FindPassword: {
		Name: "FindPassword", 
		Group: "Auth",
	},
    VerifyIdentityForLostId: {
		Name: "VerifyIdentityForLostId", 
		Group: "Auth",
	},
    VerifyIdentityForLostPassword: {
		Name: "VerifyIdentityForLostPassword", 
		Group: "Auth",
	},
    VerifyToken: {
		Name: "VerifyToken", 
		Group: "Auth",
	},
    CreateUser: {
		Name: "CreateUser", 
		Group: "User",
	},
    ListUser: {
		Name: "ListUser", 
		Group: "User",
	},
    GetUser: {
		Name: "GetUser", 
		Group: "User",
	},
    DeleteUser: {
		Name: "DeleteUser", 
		Group: "User",
	},
    UpdateUser: {
		Name: "UpdateUser", 
		Group: "User",
	},
    ResetPassword: {
		Name: "ResetPassword", 
		Group: "User",
	},
    CheckId: {
		Name: "CheckId", 
		Group: "User",
	},
    CheckEmail: {
		Name: "CheckEmail", 
		Group: "User",
	},
    GetPermissionsByAccountId: {
		Name: "GetPermissionsByAccountId", 
		Group: "User",
	},
    GetMyProfile: {
		Name: "GetMyProfile", 
		Group: "MyProfile",
	},
    UpdateMyProfile: {
		Name: "UpdateMyProfile", 
		Group: "MyProfile",
	},
    UpdateMyPassword: {
		Name: "UpdateMyPassword", 
		Group: "MyProfile",
	},
    RenewPasswordExpiredDate: {
		Name: "RenewPasswordExpiredDate", 
		Group: "MyProfile",
	},
    DeleteMyProfile: {
		Name: "DeleteMyProfile", 
		Group: "MyProfile",
	},
    Admin_CreateOrganization: {
		Name: "Admin_CreateOrganization", 
		Group: "Organization",
	},
    Admin_DeleteOrganization: {
		Name: "Admin_DeleteOrganization", 
		Group: "Organization",
	},
    GetOrganizations: {
		Name: "GetOrganizations", 
		Group: "Organization",
	},
    GetOrganization: {
		Name: "GetOrganization", 
		Group: "Organization",
	},
    CheckOrganizationName: {
		Name: "CheckOrganizationName", 
		Group: "Organization",
	},
    UpdateOrganization: {
		Name: "UpdateOrganization", 
		Group: "Organization",
	},
    UpdatePrimaryCluster: {
		Name: "UpdatePrimaryCluster", 
		Group: "Organization",
	},
    CreateCluster: {
		Name: "CreateCluster", 
		Group: "Cluster",
	},
    GetClusters: {
		Name: "GetClusters", 
		Group: "Cluster",
	},
    ImportCluster: {
		Name: "ImportCluster", 
		Group: "Cluster",
	},
    GetCluster: {
		Name: "GetCluster", 
		Group: "Cluster",
	},
    DeleteCluster: {
		Name: "DeleteCluster", 
		Group: "Cluster",
	},
    GetClusterSiteValues: {
		Name: "GetClusterSiteValues", 
		Group: "Cluster",
	},
    InstallCluster: {
		Name: "InstallCluster", 
		Group: "Cluster",
	},
    CreateBootstrapKubeconfig: {
		Name: "CreateBootstrapKubeconfig", 
		Group: "Cluster",
	},
    GetBootstrapKubeconfig: {
		Name: "GetBootstrapKubeconfig", 
		Group: "Cluster",
	},
    GetNodes: {
		Name: "GetNodes", 
		Group: "Cluster",
	},
    CreateAppgroup: {
		Name: "CreateAppgroup", 
		Group: "Appgroup",
	},
    GetAppgroups: {
		Name: "GetAppgroups", 
		Group: "Appgroup",
	},
    GetAppgroup: {
		Name: "GetAppgroup", 
		Group: "Appgroup",
	},
    DeleteAppgroup: {
		Name: "DeleteAppgroup", 
		Group: "Appgroup",
	},
    GetApplications: {
		Name: "GetApplications", 
		Group: "Appgroup",
	},
    CreateApplication: {
		Name: "CreateApplication", 
		Group: "Appgroup",
	},
    GetAppServeAppTasksByAppId: {
		Name: "GetAppServeAppTasksByAppId", 
		Group: "AppServeApp",
	},
    GetAppServeAppTaskDetail: {
		Name: "GetAppServeAppTaskDetail", 
		Group: "AppServeApp",
	},
    CreateAppServeApp: {
		Name: "CreateAppServeApp", 
		Group: "AppServeApp",
	},
    GetAppServeApps: {
		Name: "GetAppServeApps", 
		Group: "AppServeApp",
	},
    GetNumOfAppsOnStack: {
		Name: "GetNumOfAppsOnStack", 
		Group: "AppServeApp",
	},
    GetAppServeApp: {
		Name: "GetAppServeApp", 
		Group: "AppServeApp",
	},
    GetAppServeAppLatestTask: {
		Name: "GetAppServeAppLatestTask", 
		Group: "AppServeApp",
	},
    IsAppServeAppExist: {
		Name: "IsAppServeAppExist", 
		Group: "AppServeApp",
	},
    IsAppServeAppNameExist: {
		Name: "IsAppServeAppNameExist", 
		Group: "AppServeApp",
	},
    DeleteAppServeApp: {
		Name: "DeleteAppServeApp", 
		Group: "AppServeApp",
	},
    UpdateAppServeApp: {
		Name: "UpdateAppServeApp", 
		Group: "AppServeApp",
	},
    UpdateAppServeAppStatus: {
		Name: "UpdateAppServeAppStatus", 
		Group: "AppServeApp",
	},
    UpdateAppServeAppEndpoint: {
		Name: "UpdateAppServeAppEndpoint", 
		Group: "AppServeApp",
	},
    RollbackAppServeApp: {
		Name: "RollbackAppServeApp", 
		Group: "AppServeApp",
	},
    GetCloudAccounts: {
		Name: "GetCloudAccounts", 
		Group: "CloudAccount",
	},
    CreateCloudAccount: {
		Name: "CreateCloudAccount", 
		Group: "CloudAccount",
	},
    CheckCloudAccountName: {
		Name: "CheckCloudAccountName", 
		Group: "CloudAccount",
	},
    CheckAwsAccountId: {
		Name: "CheckAwsAccountId", 
		Group: "CloudAccount",
	},
    GetCloudAccount: {
		Name: "GetCloudAccount", 
		Group: "CloudAccount",
	},
    UpdateCloudAccount: {
		Name: "UpdateCloudAccount", 
		Group: "CloudAccount",
	},
    DeleteCloudAccount: {
		Name: "DeleteCloudAccount", 
		Group: "CloudAccount",
	},
    DeleteForceCloudAccount: {
		Name: "DeleteForceCloudAccount", 
		Group: "CloudAccount",
	},
    GetResourceQuota: {
		Name: "GetResourceQuota", 
		Group: "CloudAccount",
	},
    Admin_GetStackTemplates: {
		Name: "Admin_GetStackTemplates", 
		Group: "StackTemplate",
	},
    Admin_GetStackTemplate: {
		Name: "Admin_GetStackTemplate", 
		Group: "StackTemplate",
	},
    Admin_GetStackTemplateServices: {
		Name: "Admin_GetStackTemplateServices", 
		Group: "StackTemplate",
	},
    Admin_CreateStackTemplate: {
		Name: "Admin_CreateStackTemplate", 
		Group: "StackTemplate",
	},
    Admin_UpdateStackTemplate: {
		Name: "Admin_UpdateStackTemplate", 
		Group: "StackTemplate",
	},
    Admin_DeleteStackTemplate: {
		Name: "Admin_DeleteStackTemplate", 
		Group: "StackTemplate",
	},
    Admin_UpdateStackTemplateOrganizations: {
		Name: "Admin_UpdateStackTemplateOrganizations", 
		Group: "StackTemplate",
	},
    Admin_CheckStackTemplateName: {
		Name: "Admin_CheckStackTemplateName", 
		Group: "StackTemplate",
	},
    GetOrganizationStackTemplates: {
		Name: "GetOrganizationStackTemplates", 
		Group: "StackTemplate",
	},
    GetOrganizationStackTemplate: {
		Name: "GetOrganizationStackTemplate", 
		Group: "StackTemplate",
	},
    AddOrganizationStackTemplates: {
		Name: "AddOrganizationStackTemplates", 
		Group: "StackTemplate",
	},
    RemoveOrganizationStackTemplates: {
		Name: "RemoveOrganizationStackTemplates", 
		Group: "StackTemplate",
	},
    CreateDashboard: {
		Name: "CreateDashboard", 
		Group: "Dashboard",
	},
    GetDashboard: {
		Name: "GetDashboard", 
		Group: "Dashboard",
	},
    UpdateDashboard: {
		Name: "UpdateDashboard", 
		Group: "Dashboard",
	},
    GetChartsDashboard: {
		Name: "GetChartsDashboard", 
		Group: "Dashboard",
	},
    GetChartDashboard: {
		Name: "GetChartDashboard", 
		Group: "Dashboard",
	},
    GetStacksDashboard: {
		Name: "GetStacksDashboard", 
		Group: "Dashboard",
	},
    GetResourcesDashboard: {
		Name: "GetResourcesDashboard", 
		Group: "Dashboard",
	},
    Admin_CreateSystemNotificationTemplate: {
		Name: "Admin_CreateSystemNotificationTemplate", 
		Group: "SystemNotificationTemplate",
	},
    Admin_UpdateSystemNotificationTemplate: {
		Name: "Admin_UpdateSystemNotificationTemplate", 
		Group: "SystemNotificationTemplate",
	},
    Admin_DeleteSystemNotificationTemplate: {
		Name: "Admin_DeleteSystemNotificationTemplate", 
		Group: "SystemNotificationTemplate",
	},
    Admin_GetSystemNotificationTemplates: {
		Name: "Admin_GetSystemNotificationTemplates", 
		Group: "SystemNotificationTemplate",
	},
    Admin_GetSystemNotificationTemplate: {
		Name: "Admin_GetSystemNotificationTemplate", 
		Group: "SystemNotificationTemplate",
	},
    Admin_CheckSystemNotificationTemplateName: {
		Name: "Admin_CheckSystemNotificationTemplateName", 
		Group: "SystemNotificationTemplate",
	},
    GetOrganizationSystemNotificationTemplates: {
		Name: "GetOrganizationSystemNotificationTemplates", 
		Group: "SystemNotificationTemplate",
	},
    AddOrganizationSystemNotificationTemplates: {
		Name: "AddOrganizationSystemNotificationTemplates", 
		Group: "SystemNotificationTemplate",
	},
    RemoveOrganizationSystemNotificationTemplates: {
		Name: "RemoveOrganizationSystemNotificationTemplates", 
		Group: "SystemNotificationTemplate",
	},
    CreateSystemNotificationRule: {
		Name: "CreateSystemNotificationRule", 
		Group: "SystemNotificationRule",
	},
    GetSystemNotificationRules: {
		Name: "GetSystemNotificationRules", 
		Group: "SystemNotificationRule",
	},
    GetSystemNotificationRule: {
		Name: "GetSystemNotificationRule", 
		Group: "SystemNotificationRule",
	},
    CheckSystemNotificationRuleName: {
		Name: "CheckSystemNotificationRuleName", 
		Group: "SystemNotificationRule",
	},
    DeleteSystemNotificationRule: {
		Name: "DeleteSystemNotificationRule", 
		Group: "SystemNotificationRule",
	},
    UpdateSystemNotificationRule: {
		Name: "UpdateSystemNotificationRule", 
		Group: "SystemNotificationRule",
	},
    CreateSystemNotification: {
		Name: "CreateSystemNotification", 
		Group: "SystemNotification",
	},
    GetSystemNotifications: {
		Name: "GetSystemNotifications", 
		Group: "SystemNotification",
	},
    GetSystemNotification: {
		Name: "GetSystemNotification", 
		Group: "SystemNotification",
	},
    DeleteSystemNotification: {
		Name: "DeleteSystemNotification", 
		Group: "SystemNotification",
	},
    UpdateSystemNotification: {
		Name: "UpdateSystemNotification", 
		Group: "SystemNotification",
	},
    CreateSystemNotificationAction: {
		Name: "CreateSystemNotificationAction", 
		Group: "SystemNotification",
	},
    GetStacks: {
		Name: "GetStacks", 
		Group: "Stack",
	},
    CreateStack: {
		Name: "CreateStack", 
		Group: "Stack",
	},
    CheckStackName: {
		Name: "CheckStackName", 
		Group: "Stack",
	},
    GetStack: {
		Name: "GetStack", 
		Group: "Stack",
	},
    UpdateStack: {
		Name: "UpdateStack", 
		Group: "Stack",
	},
    DeleteStack: {
		Name: "DeleteStack", 
		Group: "Stack",
	},
    GetStackKubeConfig: {
		Name: "GetStackKubeConfig", 
		Group: "Stack",
	},
    GetStackStatus: {
		Name: "GetStackStatus", 
		Group: "Stack",
	},
    SetFavoriteStack: {
		Name: "SetFavoriteStack", 
		Group: "Stack",
	},
    DeleteFavoriteStack: {
		Name: "DeleteFavoriteStack", 
		Group: "Stack",
	},
    InstallStack: {
		Name: "InstallStack", 
		Group: "Stack",
	},
    CreateProject: {
		Name: "CreateProject", 
		Group: "Project",
	},
    GetProjectRoles: {
		Name: "GetProjectRoles", 
		Group: "Project",
	},
    GetProjectRole: {
		Name: "GetProjectRole", 
		Group: "Project",
	},
    GetProjects: {
		Name: "GetProjects", 
		Group: "Project",
	},
    GetProject: {
		Name: "GetProject", 
		Group: "Project",
	},
    UpdateProject: {
		Name: "UpdateProject", 
		Group: "Project",
	},
    DeleteProject: {
		Name: "DeleteProject", 
		Group: "Project",
	},
    AddProjectMember: {
		Name: "AddProjectMember", 
		Group: "Project",
	},
    GetProjectMember: {
		Name: "GetProjectMember", 
		Group: "Project",
	},
    GetProjectMembers: {
		Name: "GetProjectMembers", 
		Group: "Project",
	},
    RemoveProjectMember: {
		Name: "RemoveProjectMember", 
		Group: "Project",
	},
    UpdateProjectMemberRole: {
		Name: "UpdateProjectMemberRole", 
		Group: "Project",
	},
    CreateProjectNamespace: {
		Name: "CreateProjectNamespace", 
		Group: "Project",
	},
    GetProjectNamespaces: {
		Name: "GetProjectNamespaces", 
		Group: "Project",
	},
    GetProjectNamespace: {
		Name: "GetProjectNamespace", 
		Group: "Project",
	},
    UpdateProjectNamespace: {
		Name: "UpdateProjectNamespace", 
		Group: "Project",
	},
    DeleteProjectNamespace: {
		Name: "DeleteProjectNamespace", 
		Group: "Project",
	},
    SetFavoriteProject: {
		Name: "SetFavoriteProject", 
		Group: "Project",
	},
    SetFavoriteProjectNamespace: {
		Name: "SetFavoriteProjectNamespace", 
		Group: "Project",
	},
    UnSetFavoriteProject: {
		Name: "UnSetFavoriteProject", 
		Group: "Project",
	},
    UnSetFavoriteProjectNamespace: {
		Name: "UnSetFavoriteProjectNamespace", 
		Group: "Project",
	},
    GetProjectKubeconfig: {
		Name: "GetProjectKubeconfig", 
		Group: "Project",
	},
    GetProjectNamespaceK8sResources: {
		Name: "GetProjectNamespaceK8sResources", 
		Group: "Project",
	},
    GetAudits: {
		Name: "GetAudits", 
		Group: "Audit",
	},
    GetAudit: {
		Name: "GetAudit", 
		Group: "Audit",
	},
    DeleteAudit: {
		Name: "DeleteAudit", 
		Group: "Audit",
	},
    CreateTksRole: {
		Name: "CreateTksRole", 
		Group: "Role",
	},
    ListTksRoles: {
		Name: "ListTksRoles", 
		Group: "Role",
	},
    GetTksRole: {
		Name: "GetTksRole", 
		Group: "Role",
	},
    DeleteTksRole: {
		Name: "DeleteTksRole", 
		Group: "Role",
	},
    UpdateTksRole: {
		Name: "UpdateTksRole", 
		Group: "Role",
	},
    GetPermissionsByRoleId: {
		Name: "GetPermissionsByRoleId", 
		Group: "Role",
	},
    UpdatePermissionsByRoleId: {
		Name: "UpdatePermissionsByRoleId", 
		Group: "Role",
	},
    GetPermissionTemplates: {
		Name: "GetPermissionTemplates", 
		Group: "Permission",
	},
    Admin_CreateUser: {
		Name: "Admin_CreateUser", 
		Group: "Admin_User",
	},
    Admin_ListUser: {
		Name: "Admin_ListUser", 
		Group: "Admin_User",
	},
    Admin_GetUser: {
		Name: "Admin_GetUser", 
		Group: "Admin_User",
	},
    Admin_DeleteUser: {
		Name: "Admin_DeleteUser", 
		Group: "Admin_User",
	},
    Admin_UpdateUser: {
		Name: "Admin_UpdateUser", 
		Group: "Admin_User",
	},
    Admin_ListTksRoles: {
		Name: "Admin_ListTksRoles", 
		Group: "Admin Role",
	},
    Admin_GetTksRole: {
		Name: "Admin_GetTksRole", 
		Group: "Admin Role",
	},
    Admin_GetProjects: {
		Name: "Admin_GetProjects", 
		Group: "Admin Project",
	},
    Admin_ListPolicyTemplate: {
		Name: "Admin_ListPolicyTemplate", 
		Group: "PolicyTemplate",
	},
    Admin_CreatePolicyTemplate: {
		Name: "Admin_CreatePolicyTemplate", 
		Group: "PolicyTemplate",
	},
    Admin_DeletePolicyTemplate: {
		Name: "Admin_DeletePolicyTemplate", 
		Group: "PolicyTemplate",
	},
    Admin_GetPolicyTemplate: {
		Name: "Admin_GetPolicyTemplate", 
		Group: "PolicyTemplate",
	},
    Admin_UpdatePolicyTemplate: {
		Name: "Admin_UpdatePolicyTemplate", 
		Group: "PolicyTemplate",
	},
    Admin_GetPolicyTemplateDeploy: {
		Name: "Admin_GetPolicyTemplateDeploy", 
		Group: "PolicyTemplate",
	},
    Admin_ListPolicyTemplateStatistics: {
		Name: "Admin_ListPolicyTemplateStatistics", 
		Group: "PolicyTemplate",
	},
    Admin_ListPolicyTemplateVersions: {
		Name: "Admin_ListPolicyTemplateVersions", 
		Group: "PolicyTemplate",
	},
    Admin_CreatePolicyTemplateVersion: {
		Name: "Admin_CreatePolicyTemplateVersion", 
		Group: "PolicyTemplate",
	},
    Admin_DeletePolicyTemplateVersion: {
		Name: "Admin_DeletePolicyTemplateVersion", 
		Group: "PolicyTemplate",
	},
    Admin_GetPolicyTemplateVersion: {
		Name: "Admin_GetPolicyTemplateVersion", 
		Group: "PolicyTemplate",
	},
    Admin_ExistsPolicyTemplateKind: {
		Name: "Admin_ExistsPolicyTemplateKind", 
		Group: "PolicyTemplate",
	},
    Admin_ExistsPolicyTemplateName: {
		Name: "Admin_ExistsPolicyTemplateName", 
		Group: "PolicyTemplate",
	},
    ListClusterPolicyStatus: {
		Name: "ListClusterPolicyStatus", 
		Group: "ClusterPolicyStatus",
	},
    GetClusterPolicyTemplateStatus: {
		Name: "GetClusterPolicyTemplateStatus", 
		Group: "ClusterPolicyStatus",
	},
    UpdateClusterPolicyTemplateStatus: {
		Name: "UpdateClusterPolicyTemplateStatus", 
		Group: "ClusterPolicyStatus",
	},
    GetMandatoryPolicies: {
		Name: "GetMandatoryPolicies", 
		Group: "Policy",
	},
    SetMandatoryPolicies: {
		Name: "SetMandatoryPolicies", 
		Group: "Policy",
	},
    ListPolicy: {
		Name: "ListPolicy", 
		Group: "Policy",
	},
    CreatePolicy: {
		Name: "CreatePolicy", 
		Group: "Policy",
	},
    DeletePolicy: {
		Name: "DeletePolicy", 
		Group: "Policy",
	},
    GetPolicy: {
		Name: "GetPolicy", 
		Group: "Policy",
	},
    UpdatePolicy: {
		Name: "UpdatePolicy", 
		Group: "Policy",
	},
    UpdatePolicyTargetClusters: {
		Name: "UpdatePolicyTargetClusters", 
		Group: "Policy",
	},
    ExistsPolicyName: {
		Name: "ExistsPolicyName", 
		Group: "Policy",
	},
    GetPolicyEdit: {
		Name: "GetPolicyEdit", 
		Group: "Policy",
	},
    ListPolicyTemplate: {
		Name: "ListPolicyTemplate", 
		Group: "OrganizationPolicyTemplate",
	},
    CreatePolicyTemplate: {
		Name: "CreatePolicyTemplate", 
		Group: "OrganizationPolicyTemplate",
	},
    DeletePolicyTemplate: {
		Name: "DeletePolicyTemplate", 
		Group: "OrganizationPolicyTemplate",
	},
    GetPolicyTemplate: {
		Name: "GetPolicyTemplate", 
		Group: "OrganizationPolicyTemplate",
	},
    UpdatePolicyTemplate: {
		Name: "UpdatePolicyTemplate", 
		Group: "OrganizationPolicyTemplate",
	},
    GetPolicyTemplateDeploy: {
		Name: "GetPolicyTemplateDeploy", 
		Group: "OrganizationPolicyTemplate",
	},
    ListPolicyTemplateStatistics: {
		Name: "ListPolicyTemplateStatistics", 
		Group: "OrganizationPolicyTemplate",
	},
    ListPolicyTemplateVersions: {
		Name: "ListPolicyTemplateVersions", 
		Group: "OrganizationPolicyTemplate",
	},
    CreatePolicyTemplateVersion: {
		Name: "CreatePolicyTemplateVersion", 
		Group: "OrganizationPolicyTemplate",
	},
    DeletePolicyTemplateVersion: {
		Name: "DeletePolicyTemplateVersion", 
		Group: "OrganizationPolicyTemplate",
	},
    GetPolicyTemplateVersion: {
		Name: "GetPolicyTemplateVersion", 
		Group: "OrganizationPolicyTemplate",
	},
    ExistsPolicyTemplateKind: {
		Name: "ExistsPolicyTemplateKind", 
		Group: "OrganizationPolicyTemplate",
	},
    ExistsPolicyTemplateName: {
		Name: "ExistsPolicyTemplateName", 
		Group: "OrganizationPolicyTemplate",
	},
    ListPolicyTemplateExample: {
		Name: "ListPolicyTemplateExample", 
		Group: "PolicyTemplateExample",
	},
    GetPolicyTemplateExample: {
		Name: "GetPolicyTemplateExample", 
		Group: "PolicyTemplateExample",
	},
    UpdatePolicyTemplateExample: {
		Name: "UpdatePolicyTemplateExample", 
		Group: "PolicyTemplateExample",
	},
    DeletePolicyTemplateExample: {
		Name: "DeletePolicyTemplateExample", 
		Group: "PolicyTemplateExample",
	},
    CompileRego: {
		Name: "CompileRego", 
		Group: "Utility",
	},
}
func (e Endpoint) String() string {
	switch e {
	case Login:
		return "Login"
	case Logout:
		return "Logout"
	case RefreshToken:
		return "RefreshToken"
	case FindId:
		return "FindId"
	case FindPassword:
		return "FindPassword"
	case VerifyIdentityForLostId:
		return "VerifyIdentityForLostId"
	case VerifyIdentityForLostPassword:
		return "VerifyIdentityForLostPassword"
	case VerifyToken:
		return "VerifyToken"
	case CreateUser:
		return "CreateUser"
	case ListUser:
		return "ListUser"
	case GetUser:
		return "GetUser"
	case DeleteUser:
		return "DeleteUser"
	case UpdateUser:
		return "UpdateUser"
	case ResetPassword:
		return "ResetPassword"
	case CheckId:
		return "CheckId"
	case CheckEmail:
		return "CheckEmail"
	case GetPermissionsByAccountId:
		return "GetPermissionsByAccountId"
	case GetMyProfile:
		return "GetMyProfile"
	case UpdateMyProfile:
		return "UpdateMyProfile"
	case UpdateMyPassword:
		return "UpdateMyPassword"
	case RenewPasswordExpiredDate:
		return "RenewPasswordExpiredDate"
	case DeleteMyProfile:
		return "DeleteMyProfile"
	case Admin_CreateOrganization:
		return "Admin_CreateOrganization"
	case Admin_DeleteOrganization:
		return "Admin_DeleteOrganization"
	case GetOrganizations:
		return "GetOrganizations"
	case GetOrganization:
		return "GetOrganization"
	case CheckOrganizationName:
		return "CheckOrganizationName"
	case UpdateOrganization:
		return "UpdateOrganization"
	case UpdatePrimaryCluster:
		return "UpdatePrimaryCluster"
	case CreateCluster:
		return "CreateCluster"
	case GetClusters:
		return "GetClusters"
	case ImportCluster:
		return "ImportCluster"
	case GetCluster:
		return "GetCluster"
	case DeleteCluster:
		return "DeleteCluster"
	case GetClusterSiteValues:
		return "GetClusterSiteValues"
	case InstallCluster:
		return "InstallCluster"
	case CreateBootstrapKubeconfig:
		return "CreateBootstrapKubeconfig"
	case GetBootstrapKubeconfig:
		return "GetBootstrapKubeconfig"
	case GetNodes:
		return "GetNodes"
	case CreateAppgroup:
		return "CreateAppgroup"
	case GetAppgroups:
		return "GetAppgroups"
	case GetAppgroup:
		return "GetAppgroup"
	case DeleteAppgroup:
		return "DeleteAppgroup"
	case GetApplications:
		return "GetApplications"
	case CreateApplication:
		return "CreateApplication"
	case GetAppServeAppTasksByAppId:
		return "GetAppServeAppTasksByAppId"
	case GetAppServeAppTaskDetail:
		return "GetAppServeAppTaskDetail"
	case CreateAppServeApp:
		return "CreateAppServeApp"
	case GetAppServeApps:
		return "GetAppServeApps"
	case GetNumOfAppsOnStack:
		return "GetNumOfAppsOnStack"
	case GetAppServeApp:
		return "GetAppServeApp"
	case GetAppServeAppLatestTask:
		return "GetAppServeAppLatestTask"
	case IsAppServeAppExist:
		return "IsAppServeAppExist"
	case IsAppServeAppNameExist:
		return "IsAppServeAppNameExist"
	case DeleteAppServeApp:
		return "DeleteAppServeApp"
	case UpdateAppServeApp:
		return "UpdateAppServeApp"
	case UpdateAppServeAppStatus:
		return "UpdateAppServeAppStatus"
	case UpdateAppServeAppEndpoint:
		return "UpdateAppServeAppEndpoint"
	case RollbackAppServeApp:
		return "RollbackAppServeApp"
	case GetCloudAccounts:
		return "GetCloudAccounts"
	case CreateCloudAccount:
		return "CreateCloudAccount"
	case CheckCloudAccountName:
		return "CheckCloudAccountName"
	case CheckAwsAccountId:
		return "CheckAwsAccountId"
	case GetCloudAccount:
		return "GetCloudAccount"
	case UpdateCloudAccount:
		return "UpdateCloudAccount"
	case DeleteCloudAccount:
		return "DeleteCloudAccount"
	case DeleteForceCloudAccount:
		return "DeleteForceCloudAccount"
	case GetResourceQuota:
		return "GetResourceQuota"
	case Admin_GetStackTemplates:
		return "Admin_GetStackTemplates"
	case Admin_GetStackTemplate:
		return "Admin_GetStackTemplate"
	case Admin_GetStackTemplateServices:
		return "Admin_GetStackTemplateServices"
	case Admin_CreateStackTemplate:
		return "Admin_CreateStackTemplate"
	case Admin_UpdateStackTemplate:
		return "Admin_UpdateStackTemplate"
	case Admin_DeleteStackTemplate:
		return "Admin_DeleteStackTemplate"
	case Admin_UpdateStackTemplateOrganizations:
		return "Admin_UpdateStackTemplateOrganizations"
	case Admin_CheckStackTemplateName:
		return "Admin_CheckStackTemplateName"
	case GetOrganizationStackTemplates:
		return "GetOrganizationStackTemplates"
	case GetOrganizationStackTemplate:
		return "GetOrganizationStackTemplate"
	case AddOrganizationStackTemplates:
		return "AddOrganizationStackTemplates"
	case RemoveOrganizationStackTemplates:
		return "RemoveOrganizationStackTemplates"
	case CreateDashboard:
		return "CreateDashboard"
	case GetDashboard:
		return "GetDashboard"
	case UpdateDashboard:
		return "UpdateDashboard"
	case GetChartsDashboard:
		return "GetChartsDashboard"
	case GetChartDashboard:
		return "GetChartDashboard"
	case GetStacksDashboard:
		return "GetStacksDashboard"
	case GetResourcesDashboard:
		return "GetResourcesDashboard"
	case Admin_CreateSystemNotificationTemplate:
		return "Admin_CreateSystemNotificationTemplate"
	case Admin_UpdateSystemNotificationTemplate:
		return "Admin_UpdateSystemNotificationTemplate"
	case Admin_DeleteSystemNotificationTemplate:
		return "Admin_DeleteSystemNotificationTemplate"
	case Admin_GetSystemNotificationTemplates:
		return "Admin_GetSystemNotificationTemplates"
	case Admin_GetSystemNotificationTemplate:
		return "Admin_GetSystemNotificationTemplate"
	case Admin_CheckSystemNotificationTemplateName:
		return "Admin_CheckSystemNotificationTemplateName"
	case GetOrganizationSystemNotificationTemplates:
		return "GetOrganizationSystemNotificationTemplates"
	case AddOrganizationSystemNotificationTemplates:
		return "AddOrganizationSystemNotificationTemplates"
	case RemoveOrganizationSystemNotificationTemplates:
		return "RemoveOrganizationSystemNotificationTemplates"
	case CreateSystemNotificationRule:
		return "CreateSystemNotificationRule"
	case GetSystemNotificationRules:
		return "GetSystemNotificationRules"
	case GetSystemNotificationRule:
		return "GetSystemNotificationRule"
	case CheckSystemNotificationRuleName:
		return "CheckSystemNotificationRuleName"
	case DeleteSystemNotificationRule:
		return "DeleteSystemNotificationRule"
	case UpdateSystemNotificationRule:
		return "UpdateSystemNotificationRule"
	case CreateSystemNotification:
		return "CreateSystemNotification"
	case GetSystemNotifications:
		return "GetSystemNotifications"
	case GetSystemNotification:
		return "GetSystemNotification"
	case DeleteSystemNotification:
		return "DeleteSystemNotification"
	case UpdateSystemNotification:
		return "UpdateSystemNotification"
	case CreateSystemNotificationAction:
		return "CreateSystemNotificationAction"
	case GetStacks:
		return "GetStacks"
	case CreateStack:
		return "CreateStack"
	case CheckStackName:
		return "CheckStackName"
	case GetStack:
		return "GetStack"
	case UpdateStack:
		return "UpdateStack"
	case DeleteStack:
		return "DeleteStack"
	case GetStackKubeConfig:
		return "GetStackKubeConfig"
	case GetStackStatus:
		return "GetStackStatus"
	case SetFavoriteStack:
		return "SetFavoriteStack"
	case DeleteFavoriteStack:
		return "DeleteFavoriteStack"
	case InstallStack:
		return "InstallStack"
	case CreateProject:
		return "CreateProject"
	case GetProjectRoles:
		return "GetProjectRoles"
	case GetProjectRole:
		return "GetProjectRole"
	case GetProjects:
		return "GetProjects"
	case GetProject:
		return "GetProject"
	case UpdateProject:
		return "UpdateProject"
	case DeleteProject:
		return "DeleteProject"
	case AddProjectMember:
		return "AddProjectMember"
	case GetProjectMember:
		return "GetProjectMember"
	case GetProjectMembers:
		return "GetProjectMembers"
	case RemoveProjectMember:
		return "RemoveProjectMember"
	case UpdateProjectMemberRole:
		return "UpdateProjectMemberRole"
	case CreateProjectNamespace:
		return "CreateProjectNamespace"
	case GetProjectNamespaces:
		return "GetProjectNamespaces"
	case GetProjectNamespace:
		return "GetProjectNamespace"
	case UpdateProjectNamespace:
		return "UpdateProjectNamespace"
	case DeleteProjectNamespace:
		return "DeleteProjectNamespace"
	case SetFavoriteProject:
		return "SetFavoriteProject"
	case SetFavoriteProjectNamespace:
		return "SetFavoriteProjectNamespace"
	case UnSetFavoriteProject:
		return "UnSetFavoriteProject"
	case UnSetFavoriteProjectNamespace:
		return "UnSetFavoriteProjectNamespace"
	case GetProjectKubeconfig:
		return "GetProjectKubeconfig"
	case GetProjectNamespaceK8sResources:
		return "GetProjectNamespaceK8sResources"
	case GetAudits:
		return "GetAudits"
	case GetAudit:
		return "GetAudit"
	case DeleteAudit:
		return "DeleteAudit"
	case CreateTksRole:
		return "CreateTksRole"
	case ListTksRoles:
		return "ListTksRoles"
	case GetTksRole:
		return "GetTksRole"
	case DeleteTksRole:
		return "DeleteTksRole"
	case UpdateTksRole:
		return "UpdateTksRole"
	case GetPermissionsByRoleId:
		return "GetPermissionsByRoleId"
	case UpdatePermissionsByRoleId:
		return "UpdatePermissionsByRoleId"
	case GetPermissionTemplates:
		return "GetPermissionTemplates"
	case Admin_CreateUser:
		return "Admin_CreateUser"
	case Admin_ListUser:
		return "Admin_ListUser"
	case Admin_GetUser:
		return "Admin_GetUser"
	case Admin_DeleteUser:
		return "Admin_DeleteUser"
	case Admin_UpdateUser:
		return "Admin_UpdateUser"
	case Admin_ListTksRoles:
		return "Admin_ListTksRoles"
	case Admin_GetTksRole:
		return "Admin_GetTksRole"
	case Admin_GetProjects:
		return "Admin_GetProjects"
	case Admin_ListPolicyTemplate:
		return "Admin_ListPolicyTemplate"
	case Admin_CreatePolicyTemplate:
		return "Admin_CreatePolicyTemplate"
	case Admin_DeletePolicyTemplate:
		return "Admin_DeletePolicyTemplate"
	case Admin_GetPolicyTemplate:
		return "Admin_GetPolicyTemplate"
	case Admin_UpdatePolicyTemplate:
		return "Admin_UpdatePolicyTemplate"
	case Admin_GetPolicyTemplateDeploy:
		return "Admin_GetPolicyTemplateDeploy"
	case Admin_ListPolicyTemplateStatistics:
		return "Admin_ListPolicyTemplateStatistics"
	case Admin_ListPolicyTemplateVersions:
		return "Admin_ListPolicyTemplateVersions"
	case Admin_CreatePolicyTemplateVersion:
		return "Admin_CreatePolicyTemplateVersion"
	case Admin_DeletePolicyTemplateVersion:
		return "Admin_DeletePolicyTemplateVersion"
	case Admin_GetPolicyTemplateVersion:
		return "Admin_GetPolicyTemplateVersion"
	case Admin_ExistsPolicyTemplateKind:
		return "Admin_ExistsPolicyTemplateKind"
	case Admin_ExistsPolicyTemplateName:
		return "Admin_ExistsPolicyTemplateName"
	case ListClusterPolicyStatus:
		return "ListClusterPolicyStatus"
	case GetClusterPolicyTemplateStatus:
		return "GetClusterPolicyTemplateStatus"
	case UpdateClusterPolicyTemplateStatus:
		return "UpdateClusterPolicyTemplateStatus"
	case GetMandatoryPolicies:
		return "GetMandatoryPolicies"
	case SetMandatoryPolicies:
		return "SetMandatoryPolicies"
	case ListPolicy:
		return "ListPolicy"
	case CreatePolicy:
		return "CreatePolicy"
	case DeletePolicy:
		return "DeletePolicy"
	case GetPolicy:
		return "GetPolicy"
	case UpdatePolicy:
		return "UpdatePolicy"
	case UpdatePolicyTargetClusters:
		return "UpdatePolicyTargetClusters"
	case ExistsPolicyName:
		return "ExistsPolicyName"
	case GetPolicyEdit:
		return "GetPolicyEdit"
	case ListPolicyTemplate:
		return "ListPolicyTemplate"
	case CreatePolicyTemplate:
		return "CreatePolicyTemplate"
	case DeletePolicyTemplate:
		return "DeletePolicyTemplate"
	case GetPolicyTemplate:
		return "GetPolicyTemplate"
	case UpdatePolicyTemplate:
		return "UpdatePolicyTemplate"
	case GetPolicyTemplateDeploy:
		return "GetPolicyTemplateDeploy"
	case ListPolicyTemplateStatistics:
		return "ListPolicyTemplateStatistics"
	case ListPolicyTemplateVersions:
		return "ListPolicyTemplateVersions"
	case CreatePolicyTemplateVersion:
		return "CreatePolicyTemplateVersion"
	case DeletePolicyTemplateVersion:
		return "DeletePolicyTemplateVersion"
	case GetPolicyTemplateVersion:
		return "GetPolicyTemplateVersion"
	case ExistsPolicyTemplateKind:
		return "ExistsPolicyTemplateKind"
	case ExistsPolicyTemplateName:
		return "ExistsPolicyTemplateName"
	case ListPolicyTemplateExample:
		return "ListPolicyTemplateExample"
	case GetPolicyTemplateExample:
		return "GetPolicyTemplateExample"
	case UpdatePolicyTemplateExample:
		return "UpdatePolicyTemplateExample"
	case DeletePolicyTemplateExample:
		return "DeletePolicyTemplateExample"
	case CompileRego:
		return "CompileRego"
	default:
		return ""
	}
}
func GetEndpoint(name string) Endpoint {
	switch name {
	case "Login":
		return Login
	case "Logout":
		return Logout
	case "RefreshToken":
		return RefreshToken
	case "FindId":
		return FindId
	case "FindPassword":
		return FindPassword
	case "VerifyIdentityForLostId":
		return VerifyIdentityForLostId
	case "VerifyIdentityForLostPassword":
		return VerifyIdentityForLostPassword
	case "VerifyToken":
		return VerifyToken
	case "CreateUser":
		return CreateUser
	case "ListUser":
		return ListUser
	case "GetUser":
		return GetUser
	case "DeleteUser":
		return DeleteUser
	case "UpdateUser":
		return UpdateUser
	case "ResetPassword":
		return ResetPassword
	case "CheckId":
		return CheckId
	case "CheckEmail":
		return CheckEmail
	case "GetPermissionsByAccountId":
		return GetPermissionsByAccountId
	case "GetMyProfile":
		return GetMyProfile
	case "UpdateMyProfile":
		return UpdateMyProfile
	case "UpdateMyPassword":
		return UpdateMyPassword
	case "RenewPasswordExpiredDate":
		return RenewPasswordExpiredDate
	case "DeleteMyProfile":
		return DeleteMyProfile
	case "Admin_CreateOrganization":
		return Admin_CreateOrganization
	case "Admin_DeleteOrganization":
		return Admin_DeleteOrganization
	case "GetOrganizations":
		return GetOrganizations
	case "GetOrganization":
		return GetOrganization
	case "CheckOrganizationName":
		return CheckOrganizationName
	case "UpdateOrganization":
		return UpdateOrganization
	case "UpdatePrimaryCluster":
		return UpdatePrimaryCluster
	case "CreateCluster":
		return CreateCluster
	case "GetClusters":
		return GetClusters
	case "ImportCluster":
		return ImportCluster
	case "GetCluster":
		return GetCluster
	case "DeleteCluster":
		return DeleteCluster
	case "GetClusterSiteValues":
		return GetClusterSiteValues
	case "InstallCluster":
		return InstallCluster
	case "CreateBootstrapKubeconfig":
		return CreateBootstrapKubeconfig
	case "GetBootstrapKubeconfig":
		return GetBootstrapKubeconfig
	case "GetNodes":
		return GetNodes
	case "CreateAppgroup":
		return CreateAppgroup
	case "GetAppgroups":
		return GetAppgroups
	case "GetAppgroup":
		return GetAppgroup
	case "DeleteAppgroup":
		return DeleteAppgroup
	case "GetApplications":
		return GetApplications
	case "CreateApplication":
		return CreateApplication
	case "GetAppServeAppTasksByAppId":
		return GetAppServeAppTasksByAppId
	case "GetAppServeAppTaskDetail":
		return GetAppServeAppTaskDetail
	case "CreateAppServeApp":
		return CreateAppServeApp
	case "GetAppServeApps":
		return GetAppServeApps
	case "GetNumOfAppsOnStack":
		return GetNumOfAppsOnStack
	case "GetAppServeApp":
		return GetAppServeApp
	case "GetAppServeAppLatestTask":
		return GetAppServeAppLatestTask
	case "IsAppServeAppExist":
		return IsAppServeAppExist
	case "IsAppServeAppNameExist":
		return IsAppServeAppNameExist
	case "DeleteAppServeApp":
		return DeleteAppServeApp
	case "UpdateAppServeApp":
		return UpdateAppServeApp
	case "UpdateAppServeAppStatus":
		return UpdateAppServeAppStatus
	case "UpdateAppServeAppEndpoint":
		return UpdateAppServeAppEndpoint
	case "RollbackAppServeApp":
		return RollbackAppServeApp
	case "GetCloudAccounts":
		return GetCloudAccounts
	case "CreateCloudAccount":
		return CreateCloudAccount
	case "CheckCloudAccountName":
		return CheckCloudAccountName
	case "CheckAwsAccountId":
		return CheckAwsAccountId
	case "GetCloudAccount":
		return GetCloudAccount
	case "UpdateCloudAccount":
		return UpdateCloudAccount
	case "DeleteCloudAccount":
		return DeleteCloudAccount
	case "DeleteForceCloudAccount":
		return DeleteForceCloudAccount
	case "GetResourceQuota":
		return GetResourceQuota
	case "Admin_GetStackTemplates":
		return Admin_GetStackTemplates
	case "Admin_GetStackTemplate":
		return Admin_GetStackTemplate
	case "Admin_GetStackTemplateServices":
		return Admin_GetStackTemplateServices
	case "Admin_CreateStackTemplate":
		return Admin_CreateStackTemplate
	case "Admin_UpdateStackTemplate":
		return Admin_UpdateStackTemplate
	case "Admin_DeleteStackTemplate":
		return Admin_DeleteStackTemplate
	case "Admin_UpdateStackTemplateOrganizations":
		return Admin_UpdateStackTemplateOrganizations
	case "Admin_CheckStackTemplateName":
		return Admin_CheckStackTemplateName
	case "GetOrganizationStackTemplates":
		return GetOrganizationStackTemplates
	case "GetOrganizationStackTemplate":
		return GetOrganizationStackTemplate
	case "AddOrganizationStackTemplates":
		return AddOrganizationStackTemplates
	case "RemoveOrganizationStackTemplates":
		return RemoveOrganizationStackTemplates
	case "CreateDashboard":
		return CreateDashboard
	case "GetDashboard":
		return GetDashboard
	case "UpdateDashboard":
		return UpdateDashboard
	case "GetChartsDashboard":
		return GetChartsDashboard
	case "GetChartDashboard":
		return GetChartDashboard
	case "GetStacksDashboard":
		return GetStacksDashboard
	case "GetResourcesDashboard":
		return GetResourcesDashboard
	case "Admin_CreateSystemNotificationTemplate":
		return Admin_CreateSystemNotificationTemplate
	case "Admin_UpdateSystemNotificationTemplate":
		return Admin_UpdateSystemNotificationTemplate
	case "Admin_DeleteSystemNotificationTemplate":
		return Admin_DeleteSystemNotificationTemplate
	case "Admin_GetSystemNotificationTemplates":
		return Admin_GetSystemNotificationTemplates
	case "Admin_GetSystemNotificationTemplate":
		return Admin_GetSystemNotificationTemplate
	case "Admin_CheckSystemNotificationTemplateName":
		return Admin_CheckSystemNotificationTemplateName
	case "GetOrganizationSystemNotificationTemplates":
		return GetOrganizationSystemNotificationTemplates
	case "AddOrganizationSystemNotificationTemplates":
		return AddOrganizationSystemNotificationTemplates
	case "RemoveOrganizationSystemNotificationTemplates":
		return RemoveOrganizationSystemNotificationTemplates
	case "CreateSystemNotificationRule":
		return CreateSystemNotificationRule
	case "GetSystemNotificationRules":
		return GetSystemNotificationRules
	case "GetSystemNotificationRule":
		return GetSystemNotificationRule
	case "CheckSystemNotificationRuleName":
		return CheckSystemNotificationRuleName
	case "DeleteSystemNotificationRule":
		return DeleteSystemNotificationRule
	case "UpdateSystemNotificationRule":
		return UpdateSystemNotificationRule
	case "CreateSystemNotification":
		return CreateSystemNotification
	case "GetSystemNotifications":
		return GetSystemNotifications
	case "GetSystemNotification":
		return GetSystemNotification
	case "DeleteSystemNotification":
		return DeleteSystemNotification
	case "UpdateSystemNotification":
		return UpdateSystemNotification
	case "CreateSystemNotificationAction":
		return CreateSystemNotificationAction
	case "GetStacks":
		return GetStacks
	case "CreateStack":
		return CreateStack
	case "CheckStackName":
		return CheckStackName
	case "GetStack":
		return GetStack
	case "UpdateStack":
		return UpdateStack
	case "DeleteStack":
		return DeleteStack
	case "GetStackKubeConfig":
		return GetStackKubeConfig
	case "GetStackStatus":
		return GetStackStatus
	case "SetFavoriteStack":
		return SetFavoriteStack
	case "DeleteFavoriteStack":
		return DeleteFavoriteStack
	case "InstallStack":
		return InstallStack
	case "CreateProject":
		return CreateProject
	case "GetProjectRoles":
		return GetProjectRoles
	case "GetProjectRole":
		return GetProjectRole
	case "GetProjects":
		return GetProjects
	case "GetProject":
		return GetProject
	case "UpdateProject":
		return UpdateProject
	case "DeleteProject":
		return DeleteProject
	case "AddProjectMember":
		return AddProjectMember
	case "GetProjectMember":
		return GetProjectMember
	case "GetProjectMembers":
		return GetProjectMembers
	case "RemoveProjectMember":
		return RemoveProjectMember
	case "UpdateProjectMemberRole":
		return UpdateProjectMemberRole
	case "CreateProjectNamespace":
		return CreateProjectNamespace
	case "GetProjectNamespaces":
		return GetProjectNamespaces
	case "GetProjectNamespace":
		return GetProjectNamespace
	case "UpdateProjectNamespace":
		return UpdateProjectNamespace
	case "DeleteProjectNamespace":
		return DeleteProjectNamespace
	case "SetFavoriteProject":
		return SetFavoriteProject
	case "SetFavoriteProjectNamespace":
		return SetFavoriteProjectNamespace
	case "UnSetFavoriteProject":
		return UnSetFavoriteProject
	case "UnSetFavoriteProjectNamespace":
		return UnSetFavoriteProjectNamespace
	case "GetProjectKubeconfig":
		return GetProjectKubeconfig
	case "GetProjectNamespaceK8sResources":
		return GetProjectNamespaceK8sResources
	case "GetAudits":
		return GetAudits
	case "GetAudit":
		return GetAudit
	case "DeleteAudit":
		return DeleteAudit
	case "CreateTksRole":
		return CreateTksRole
	case "ListTksRoles":
		return ListTksRoles
	case "GetTksRole":
		return GetTksRole
	case "DeleteTksRole":
		return DeleteTksRole
	case "UpdateTksRole":
		return UpdateTksRole
	case "GetPermissionsByRoleId":
		return GetPermissionsByRoleId
	case "UpdatePermissionsByRoleId":
		return UpdatePermissionsByRoleId
	case "GetPermissionTemplates":
		return GetPermissionTemplates
	case "Admin_CreateUser":
		return Admin_CreateUser
	case "Admin_ListUser":
		return Admin_ListUser
	case "Admin_GetUser":
		return Admin_GetUser
	case "Admin_DeleteUser":
		return Admin_DeleteUser
	case "Admin_UpdateUser":
		return Admin_UpdateUser
	case "Admin_ListTksRoles":
		return Admin_ListTksRoles
	case "Admin_GetTksRole":
		return Admin_GetTksRole
	case "Admin_GetProjects":
		return Admin_GetProjects
	case "Admin_ListPolicyTemplate":
		return Admin_ListPolicyTemplate
	case "Admin_CreatePolicyTemplate":
		return Admin_CreatePolicyTemplate
	case "Admin_DeletePolicyTemplate":
		return Admin_DeletePolicyTemplate
	case "Admin_GetPolicyTemplate":
		return Admin_GetPolicyTemplate
	case "Admin_UpdatePolicyTemplate":
		return Admin_UpdatePolicyTemplate
	case "Admin_GetPolicyTemplateDeploy":
		return Admin_GetPolicyTemplateDeploy
	case "Admin_ListPolicyTemplateStatistics":
		return Admin_ListPolicyTemplateStatistics
	case "Admin_ListPolicyTemplateVersions":
		return Admin_ListPolicyTemplateVersions
	case "Admin_CreatePolicyTemplateVersion":
		return Admin_CreatePolicyTemplateVersion
	case "Admin_DeletePolicyTemplateVersion":
		return Admin_DeletePolicyTemplateVersion
	case "Admin_GetPolicyTemplateVersion":
		return Admin_GetPolicyTemplateVersion
	case "Admin_ExistsPolicyTemplateKind":
		return Admin_ExistsPolicyTemplateKind
	case "Admin_ExistsPolicyTemplateName":
		return Admin_ExistsPolicyTemplateName
	case "ListClusterPolicyStatus":
		return ListClusterPolicyStatus
	case "GetClusterPolicyTemplateStatus":
		return GetClusterPolicyTemplateStatus
	case "UpdateClusterPolicyTemplateStatus":
		return UpdateClusterPolicyTemplateStatus
	case "GetMandatoryPolicies":
		return GetMandatoryPolicies
	case "SetMandatoryPolicies":
		return SetMandatoryPolicies
	case "ListPolicy":
		return ListPolicy
	case "CreatePolicy":
		return CreatePolicy
	case "DeletePolicy":
		return DeletePolicy
	case "GetPolicy":
		return GetPolicy
	case "UpdatePolicy":
		return UpdatePolicy
	case "UpdatePolicyTargetClusters":
		return UpdatePolicyTargetClusters
	case "ExistsPolicyName":
		return ExistsPolicyName
	case "GetPolicyEdit":
		return GetPolicyEdit
	case "ListPolicyTemplate":
		return ListPolicyTemplate
	case "CreatePolicyTemplate":
		return CreatePolicyTemplate
	case "DeletePolicyTemplate":
		return DeletePolicyTemplate
	case "GetPolicyTemplate":
		return GetPolicyTemplate
	case "UpdatePolicyTemplate":
		return UpdatePolicyTemplate
	case "GetPolicyTemplateDeploy":
		return GetPolicyTemplateDeploy
	case "ListPolicyTemplateStatistics":
		return ListPolicyTemplateStatistics
	case "ListPolicyTemplateVersions":
		return ListPolicyTemplateVersions
	case "CreatePolicyTemplateVersion":
		return CreatePolicyTemplateVersion
	case "DeletePolicyTemplateVersion":
		return DeletePolicyTemplateVersion
	case "GetPolicyTemplateVersion":
		return GetPolicyTemplateVersion
	case "ExistsPolicyTemplateKind":
		return ExistsPolicyTemplateKind
	case "ExistsPolicyTemplateName":
		return ExistsPolicyTemplateName
	case "ListPolicyTemplateExample":
		return ListPolicyTemplateExample
	case "GetPolicyTemplateExample":
		return GetPolicyTemplateExample
	case "UpdatePolicyTemplateExample":
		return UpdatePolicyTemplateExample
	case "DeletePolicyTemplateExample":
		return DeletePolicyTemplateExample
	case "CompileRego":
		return CompileRego
	default:
		return -1
	}
}
