package role

import internalApi "github.com/openinfradev/tks-api/internal/delivery/api"

var defaults = []*defaultPermission{
	&defaultPermissionOfAdminInMaster,
	&defaultPermissionOfUserInMaster,
	&defaultPermissionOfAdmin,
	&defaultPermissionOfUser,
	&defaultLeaderPermission,
	&defaultMemberPermission,
	&defaultViewerPermission,
}

func getDefaultPermissions() []*defaultPermission {
	return defaults
}

type defaultPermission struct {
	role        Role
	permissions *[]internalApi.Endpoint
}

var defaultPermissionOfAdminInMaster = defaultPermission{
	role:        Admin,
	permissions: &[]internalApi.Endpoint{},
}

var defaultPermissionOfUserInMaster = defaultPermission{
	role:        User,
	permissions: &[]internalApi.Endpoint{},
}

var defaultPermissionOfAdmin = defaultPermission{
	role: Admin,
	permissions: &[]internalApi.Endpoint{
		// Auth
		internalApi.Logout,
		internalApi.RefreshToken,
		internalApi.VerifyToken,

		// User
		internalApi.CreateUser,
		internalApi.ListUser,
		internalApi.GetUser,
		internalApi.DeleteUser,
		internalApi.UpdateUser,
		internalApi.ResetPassword,
		internalApi.CheckId,
		internalApi.CheckEmail,

		// MyProfile
		internalApi.GetMyProfile,
		internalApi.UpdateMyProfile,
		internalApi.UpdateMyPassword,
		internalApi.RenewPasswordExpiredDate,
		internalApi.DeleteMyProfile,

		// Organization
		internalApi.CreateOrganization,
		internalApi.GetOrganizations,
		internalApi.GetOrganization,
		internalApi.DeleteOrganization,
		internalApi.UpdateOrganization,

		// Cluster
		internalApi.UpdatePrimaryCluster,
		internalApi.CreateCluster,
		internalApi.GetClusters,
		internalApi.ImportCluster,
		internalApi.GetCluster,
		internalApi.DeleteCluster,
		internalApi.GetClusterSiteValues,
		internalApi.InstallCluster,
		internalApi.CreateBootstrapKubeconfig,
		internalApi.GetBootstrapKubeconfig,
		internalApi.GetNodes,

		// Appgroup
		internalApi.CreateAppgroup,
		internalApi.GetAppgroups,
		internalApi.GetAppgroup,
		internalApi.DeleteAppgroup,
		internalApi.GetApplications,
		internalApi.CreateApplication,

		// AppServeApp
		internalApi.CreateAppServeApp,
		internalApi.GetAppServeApps,
		internalApi.GetNumOfAppsOnStack,
		internalApi.GetAppServeApp,
		internalApi.GetAppServeAppLatestTask,
		internalApi.IsAppServeAppExist,
		internalApi.IsAppServeAppNameExist,
		internalApi.DeleteAppServeApp,
		internalApi.UpdateAppServeApp,
		internalApi.UpdateAppServeAppStatus,
		internalApi.UpdateAppServeAppEndpoint,
		internalApi.RollbackAppServeApp,

		// CloudAccount
		internalApi.GetCloudAccounts,
		internalApi.CreateCloudAccount,
		internalApi.CheckCloudAccountName,
		internalApi.CheckAwsAccountId,
		internalApi.GetCloudAccount,
		internalApi.UpdateCloudAccount,
		internalApi.DeleteCloudAccount,
		internalApi.DeleteForceCloudAccount,
		internalApi.GetResourceQuota,

		// Dashboard
		internalApi.GetChartsDashboard,
		internalApi.GetChartDashboard,
		internalApi.GetStacksDashboard,
		internalApi.GetResourcesDashboard,

		// Alert
		internalApi.CreateAlert,
		internalApi.GetAlerts,
		internalApi.GetAlert,
		internalApi.DeleteAlert,
		internalApi.UpdateAlert,
		internalApi.CreateAlertAction,

		// Stack
		internalApi.GetStacks,
		internalApi.CreateStack,
		internalApi.CheckStackName,
		internalApi.GetStack,
		internalApi.UpdateStack,
		internalApi.DeleteStack,
		internalApi.GetStackKubeConfig,
		internalApi.GetStackStatus,
		internalApi.SetFavoriteStack,
		internalApi.DeleteFavoriteStack,
		internalApi.InstallStack,

		// Project
		internalApi.CreateProject,
		internalApi.GetProjects,
		internalApi.GetProject,
		internalApi.UpdateProject,
		internalApi.DeleteProject,
		internalApi.AddProjectMember,
		internalApi.GetProjectMembers,
		internalApi.RemoveProjectMember,
		internalApi.UpdateProjectMemberRole,
		internalApi.CreateProjectNamespace,
		internalApi.GetProjectNamespaces,
		internalApi.GetProjectNamespace,
		internalApi.DeleteProjectNamespace,
		internalApi.SetFavoriteProject,
		internalApi.SetFavoriteProjectNamespace,
		internalApi.UnSetFavoriteProject,
		internalApi.UnSetFavoriteProjectNamespace,
	},
}

// TODO: check-up the permission of User
var defaultPermissionOfUser = defaultPermission{
	role: User,
	permissions: &[]internalApi.Endpoint{
		// Auth
		internalApi.Logout,
		internalApi.RefreshToken,
		internalApi.VerifyToken,

		// User
		internalApi.ListUser,
		internalApi.GetUser,
		internalApi.CheckId,
		internalApi.CheckEmail,

		// MyProfile
		internalApi.GetMyProfile,
		internalApi.UpdateMyProfile,
		internalApi.UpdateMyPassword,
		internalApi.RenewPasswordExpiredDate,
		internalApi.DeleteMyProfile,

		// Organization
		internalApi.GetOrganizations,
		internalApi.GetOrganization,

		// Cluster
		internalApi.GetClusters,
		internalApi.GetCluster,
		internalApi.GetClusterSiteValues,
		internalApi.GetBootstrapKubeconfig,
		internalApi.GetNodes,

		// Appgroup
		internalApi.CreateAppgroup,
		internalApi.GetAppgroups,
		internalApi.GetAppgroup,
		internalApi.DeleteAppgroup,
		internalApi.GetApplications,
		internalApi.CreateApplication,

		// AppServeApp
		internalApi.CreateAppServeApp,
		internalApi.GetAppServeApps,
		internalApi.GetNumOfAppsOnStack,
		internalApi.GetAppServeApp,
		internalApi.GetAppServeAppLatestTask,
		internalApi.IsAppServeAppExist,
		internalApi.IsAppServeAppNameExist,
		internalApi.DeleteAppServeApp,
		internalApi.UpdateAppServeApp,
		internalApi.UpdateAppServeAppStatus,
		internalApi.UpdateAppServeAppEndpoint,
		internalApi.RollbackAppServeApp,

		// CloudAccount
		internalApi.GetCloudAccounts,
		internalApi.GetCloudAccount,
		internalApi.GetResourceQuota,

		// StackTemplate
		internalApi.Admin_GetStackTemplates,
		internalApi.Admin_GetStackTemplate,

		// Dashboard
		internalApi.GetChartsDashboard,
		internalApi.GetChartDashboard,
		internalApi.GetStacksDashboard,
		internalApi.GetResourcesDashboard,

		// Alert
		internalApi.GetAlerts,
		internalApi.GetAlert,

		// Stack
		internalApi.GetStacks,
		internalApi.GetStack,
		internalApi.GetStackKubeConfig,
		internalApi.GetStackStatus,
		internalApi.SetFavoriteStack,
		internalApi.DeleteFavoriteStack,

		// Project
		internalApi.CreateProject,
		internalApi.GetProjects,
		internalApi.GetProject,
		internalApi.UpdateProject,
		internalApi.DeleteProject,
		internalApi.AddProjectMember,
		internalApi.GetProjectMembers,
		internalApi.RemoveProjectMember,
		internalApi.UpdateProjectMemberRole,
		internalApi.CreateProjectNamespace,
		internalApi.GetProjectNamespaces,
		internalApi.GetProjectNamespace,
		internalApi.DeleteProjectNamespace,
		internalApi.SetFavoriteProject,
		internalApi.SetFavoriteProjectNamespace,
		internalApi.UnSetFavoriteProject,
		internalApi.UnSetFavoriteProjectNamespace,
	},
}

var defaultLeaderPermission = defaultPermission{
	role: leader,
	permissions: &[]internalApi.Endpoint{
		// Project
		internalApi.CreateProject,
		internalApi.GetProjects,
		internalApi.GetProject,
		internalApi.UpdateProject,
		internalApi.DeleteProject,
		internalApi.AddProjectMember,
		internalApi.GetProjectMembers,
		internalApi.RemoveProjectMember,
		internalApi.UpdateProjectMemberRole,
		internalApi.CreateProjectNamespace,
		internalApi.GetProjectNamespaces,
		internalApi.GetProjectNamespace,
		internalApi.DeleteProjectNamespace,
		internalApi.SetFavoriteProject,
		internalApi.SetFavoriteProjectNamespace,
		internalApi.UnSetFavoriteProject,
		internalApi.UnSetFavoriteProjectNamespace,
	},
}

var defaultMemberPermission = defaultPermission{
	role: member,
	permissions: &[]internalApi.Endpoint{
		// Project
		internalApi.GetProjects,
		internalApi.GetProject,
		internalApi.GetProjectMembers,
		internalApi.GetProjectNamespaces,
		internalApi.GetProjectNamespace,
		internalApi.SetFavoriteProject,
		internalApi.SetFavoriteProjectNamespace,
		internalApi.UnSetFavoriteProject,
		internalApi.UnSetFavoriteProjectNamespace,
	},
}

var defaultViewerPermission = defaultPermission{
	role: viewer,
	permissions: &[]internalApi.Endpoint{
		// Project
		internalApi.GetProjects,
		internalApi.GetProject,
		internalApi.GetProjectMembers,
		internalApi.GetProjectNamespaces,
		internalApi.GetProjectNamespace,
		internalApi.SetFavoriteProject,
		internalApi.SetFavoriteProjectNamespace,
		internalApi.UnSetFavoriteProject,
		internalApi.UnSetFavoriteProjectNamespace,
	},
}
