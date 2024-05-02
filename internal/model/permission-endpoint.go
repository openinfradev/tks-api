package model

import (
	"github.com/openinfradev/tks-api/internal/delivery/api"
	"gorm.io/gorm"
	"sort"
)

type PermissionEndpoint struct {
	EdgeKey      string `gorm:"primaryKey;type:text;"`
	EndpointName string `gorm:"primaryKey;type:text;"`

	Permission Permission `gorm:"foreignKey:EdgeKey;references:EdgeKey"`
	Endpoint   Endpoint   `gorm:"foreignKey:EndpointName;references:Name"`
}

func (PermissionEndpoint) TableName() string {
	return "permission_endpoints"
}

var (
	// map[EdgeKey][]Endpoints
	edgeKeyEndpointMap = map[string][]Endpoint{
		TopDashboardKey + "-" + MiddleDashboardKey + "-" + OperationRead: endpointObjects(
			api.GetDashboard,
			api.GetChartsDashboard,
			api.GetChartDashboard,
			api.GetStacksDashboard,
			api.GetResourcesDashboard,
		),
		TopDashboardKey + "-" + MiddleDashboardKey + "-" + OperationUpdate: endpointObjects(
			api.CreateDashboard,
			api.UpdateDashboard,
		),

		TopStackKey + "-" + MiddleStackKey + "-" + OperationRead: endpointObjects(
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

		TopStackKey + "-" + MiddleStackKey + "-" + OperationCreate: endpointObjects(
			api.CreateStack,
			api.InstallStack,

			// Cluster
			api.CreateBootstrapKubeconfig,
			api.GetBootstrapKubeconfig,
			api.GetNodes,
		),
		TopStackKey + "-" + MiddleStackKey + "-" + OperationUpdate: endpointObjects(
			api.UpdateStack,
		),
		TopStackKey + "-" + MiddleStackKey + "-" + OperationDelete: endpointObjects(
			// Stack
			api.DeleteStack,

			// Cluster
			api.DeleteCluster,

			// AppGroup
			api.DeleteAppgroup,
		),
		TopPolicyKey + "-" + MiddlePolicyKey + "-" + OperationRead: endpointObjects(
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
		TopPolicyKey + "-" + MiddlePolicyKey + "-" + OperationCreate: endpointObjects(
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
		TopPolicyKey + "-" + MiddlePolicyKey + "-" + OperationUpdate: endpointObjects(
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

		TopPolicyKey + "-" + MiddlePolicyKey + "-" + OperationDelete: endpointObjects(
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
		TopNotificationKey + "-" + MiddleNotificationKey + "-" + OperationRead: endpointObjects(
			api.GetSystemNotification,
			api.GetSystemNotifications,
		),
		TopNotificationKey + "-" + MiddleNotificationKey + "-" + OperationUpdate: endpointObjects(
			api.UpdateSystemNotification,
			api.CreateSystemNotificationAction,
		),
		TopNotificationKey + "-" + MiddleNotificationKey + "-" + OperationDownload: endpointObjects(),

		TopProjectKey + "-" + MiddleProjectKey + "-" + OperationRead: endpointObjects(
			api.GetProjects,
			api.GetProject,
			api.GetProjectKubeconfig,
		),
		TopProjectKey + "-" + MiddleProjectKey + "-" + OperationCreate: endpointObjects(
			api.CreateProject,
		), TopProjectKey + "-" + MiddleProjectKey + "-" + OperationUpdate: endpointObjects(
			api.UpdateProject,
		),
		TopProjectKey + "-" + MiddleProjectKey + "-" + OperationDelete: endpointObjects(
			api.DeleteProject,
		),
		TopProjectKey + "-" + MiddleProjectCommonConfigurationKey + "-" + OperationRead: endpointObjects(
			api.GetProjects,
			api.GetProject,

			api.GetProjectRoles,
			api.GetProjectRole,
		),
		TopProjectKey + "-" + MiddleProjectCommonConfigurationKey + "-" + OperationUpdate: endpointObjects(
			api.UpdateProject,
		),
		TopProjectKey + "-" + MiddleProjectMemberConfigurationKey + "-" + OperationRead: endpointObjects(
			api.GetProjectMembers,
			api.GetProjectMember,
			api.GetProjectRoles,
			api.GetProjectRole,
		),
		TopProjectKey + "-" + MiddleProjectMemberConfigurationKey + "-" + OperationCreate: endpointObjects(
			api.AddProjectMember,
		),
		TopProjectKey + "-" + MiddleProjectMemberConfigurationKey + "-" + OperationUpdate: endpointObjects(
			api.UpdateProjectMemberRole,
		),
		TopProjectKey + "-" + MiddleProjectMemberConfigurationKey + "-" + OperationDelete: endpointObjects(
			api.RemoveProjectMember,
		),
		TopProjectKey + "-" + MiddleProjectNamespaceKey + "-" + OperationRead: endpointObjects(
			api.GetProjectNamespaces,
			api.GetProjectNamespace,
			api.GetProjectNamespaceK8sResources,
		),
		TopProjectKey + "-" + MiddleProjectNamespaceKey + "-" + OperationCreate: endpointObjects(
			api.CreateProjectNamespace,
		),
		TopProjectKey + "-" + MiddleProjectNamespaceKey + "-" + OperationUpdate: endpointObjects(
			api.UpdateProjectNamespace,
		),
		TopProjectKey + "-" + MiddleProjectNamespaceKey + "-" + OperationDelete: endpointObjects(
			api.DeleteProjectNamespace,
		),
		TopProjectKey + "-" + MiddleProjectAppServeKey + "-" + OperationRead: endpointObjects(
			api.GetAppServeApps,
			api.GetAppServeApp,
			api.GetNumOfAppsOnStack,
			api.GetAppServeAppLatestTask,
			api.IsAppServeAppExist,
			api.IsAppServeAppNameExist,
			api.GetAppServeAppTaskDetail,
			api.GetAppServeAppTasksByAppId,
		),
		TopProjectKey + "-" + MiddleProjectAppServeKey + "-" + OperationCreate: endpointObjects(
			api.CreateAppServeApp,
			api.IsAppServeAppExist,
			api.IsAppServeAppNameExist,
			api.UpdateAppServeApp,
			api.UpdateAppServeAppEndpoint,
			api.UpdateAppServeAppStatus,
			api.RollbackAppServeApp,
		),
		TopProjectKey + "-" + MiddleProjectAppServeKey + "-" + OperationUpdate: endpointObjects(
			api.CreateAppServeApp,
			api.IsAppServeAppExist,
			api.IsAppServeAppNameExist,
			api.UpdateAppServeApp,
			api.UpdateAppServeAppEndpoint,
			api.UpdateAppServeAppStatus,
			api.RollbackAppServeApp,
		),
		TopProjectKey + "-" + MiddleProjectAppServeKey + "-" + OperationDelete: endpointObjects(
			api.DeleteAppServeApp,
		),
		TopConfigurationKey + "-" + MiddleConfigurationKey + "-" + OperationRead:   endpointObjects(),
		TopConfigurationKey + "-" + MiddleConfigurationKey + "-" + OperationUpdate: endpointObjects(),
		TopConfigurationKey + "-" + MiddleConfigurationCloudAccountKey + "-" + OperationRead: endpointObjects(
			api.GetCloudAccounts,
			api.GetCloudAccount,
			api.CheckCloudAccountName,
			api.CheckAwsAccountId,
			api.GetResourceQuota,
		),
		TopConfigurationKey + "-" + MiddleConfigurationCloudAccountKey + "-" + OperationCreate: endpointObjects(
			api.CreateCloudAccount,
		),
		TopConfigurationKey + "-" + MiddleConfigurationCloudAccountKey + "-" + OperationUpdate: endpointObjects(
			api.UpdateCloudAccount,
		),
		TopConfigurationKey + "-" + MiddleConfigurationCloudAccountKey + "-" + OperationDelete: endpointObjects(
			api.DeleteCloudAccount,
			api.DeleteForceCloudAccount,
		),
		TopConfigurationKey + "-" + MiddleConfigurationProjectKey + "-" + OperationRead:   endpointObjects(),
		TopConfigurationKey + "-" + MiddleConfigurationProjectKey + "-" + OperationCreate: endpointObjects(),
		TopConfigurationKey + "-" + MiddleConfigurationUserKey + "-" + OperationRead: endpointObjects(
			api.ListUser,
			api.GetUser,
			api.CheckId,
			api.CheckEmail,
			api.GetPermissionsByAccountId,
		),
		TopConfigurationKey + "-" + MiddleConfigurationUserKey + "-" + OperationCreate: endpointObjects(
			api.CreateUser,
		),
		TopConfigurationKey + "-" + MiddleConfigurationUserKey + "-" + OperationUpdate: endpointObjects(
			api.UpdateUser,
			api.ResetPassword,
		),
		TopConfigurationKey + "-" + MiddleConfigurationUserKey + "-" + OperationDelete: endpointObjects(
			api.DeleteUser,
		),
		TopConfigurationKey + "-" + MiddleConfigurationRoleKey + "-" + OperationRead: endpointObjects(
			api.ListTksRoles,
			api.GetTksRole,
			api.GetPermissionsByRoleId,
			api.GetPermissionTemplates,
		),
		TopConfigurationKey + "-" + MiddleConfigurationRoleKey + "-" + OperationCreate: endpointObjects(
			api.CreateTksRole,
		),
		TopConfigurationKey + "-" + MiddleConfigurationRoleKey + "-" + OperationUpdate: endpointObjects(
			api.UpdateTksRole,
			api.UpdatePermissionsByRoleId,
		),
		TopConfigurationKey + "-" + MiddleConfigurationRoleKey + "-" + OperationDelete: endpointObjects(
			api.DeleteTksRole,
		),
		TopConfigurationKey + "-" + MiddleConfigurationSystemNotificationKey + "-" + OperationRead: endpointObjects(
			api.GetSystemNotificationRules,
			api.GetSystemNotificationRule,
		),
		TopConfigurationKey + "-" + MiddleConfigurationSystemNotificationKey + "-" + OperationCreate: endpointObjects(
			api.CreateSystemNotificationRule,
		),
		TopConfigurationKey + "-" + MiddleConfigurationSystemNotificationKey + "-" + OperationUpdate: endpointObjects(
			api.UpdateSystemNotificationRule,
		),
		TopConfigurationKey + "-" + MiddleConfigurationSystemNotificationKey + "-" + OperationDelete: endpointObjects(
			api.DeleteSystemNotificationRule,
		),
		CommonKey: endpointObjects(
			// Auth
			api.Login,
			api.Logout,
			api.RefreshToken,
			api.FindId,
			api.FindPassword,
			api.VerifyIdentityForLostId,
			api.VerifyIdentityForLostPassword,
			api.VerifyToken,

			// User
			api.GetUser,
			api.GetPermissionsByAccountId,

			// MyProfile
			api.GetMyProfile,
			api.UpdateMyProfile,
			api.UpdateMyPassword,
			api.RenewPasswordExpiredDate,
			api.DeleteMyProfile,

			// Organization
			api.GetOrganization,

			// Role
			api.GetPermissionsByRoleId,

			// Utiliy
			api.CompileRego,
		),
	}
)

// ForceSyncToLatestPermissionEndpointMapping is used to sync the permission endpoint mapping to the latest version.
func ForceSyncToLatestPermissionEndpointMapping(db *gorm.DB, permissionSet *PermissionSet) error {
	var storedPermissionEndpoints []PermissionEndpoint
	var storedEdgeKeyEndpointMaps = make(map[string][]Endpoint)
	if err := db.Find(&storedPermissionEndpoints).Error; err != nil {
		return err
	}
	for _, pe := range storedPermissionEndpoints {
		storedEdgeKeyEndpointMaps[pe.EdgeKey] = append(storedEdgeKeyEndpointMaps[pe.EdgeKey], pe.Endpoint)
	}

	var shouldInsertEdgeKeyEndpointMaps, shouldReplaceEdgeKeyEndpointMaps map[string][]Endpoint
	shouldReplaceEdgeKeyEndpointMaps = make(map[string][]Endpoint)
	shouldInsertEdgeKeyEndpointMaps = edgeKeyEndpointMap

	for edgeKey, endpoints := range storedEdgeKeyEndpointMaps {
		if compareEndpointArrays(endpoints, edgeKeyEndpointMap[edgeKey]) {
			delete(shouldInsertEdgeKeyEndpointMaps, edgeKey)
		} else {
			shouldReplaceEdgeKeyEndpointMaps[edgeKey] = endpoints
		}
	}

	for edgeKey, endpoints := range shouldInsertEdgeKeyEndpointMaps {
		for _, endpoint := range endpoints {
			if err := db.Create(&PermissionEndpoint{
				EdgeKey:      edgeKey,
				EndpointName: endpoint.Name,
			}).Error; err != nil {
				return err
			}
		}
	}

	for edgeKey, endpoints := range shouldReplaceEdgeKeyEndpointMaps {
		if err := db.Where("edge_key = ?", edgeKey).Delete(&PermissionEndpoint{}).Error; err != nil {
			return err
		}
		for _, endpoint := range endpoints {
			if err := db.Create(&PermissionEndpoint{
				EdgeKey:      edgeKey,
				EndpointName: endpoint.Name,
			}).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

// Compare two arrays of Endpoint objects
func compareEndpointArrays(a, b []Endpoint) bool {
	if len(a) != len(b) {
		return false
	}

	// sort the arrays
	sort.Slice(a, func(i, j int) bool {
		return a[i].Name < a[j].Name
	})

	sort.Slice(b, func(i, j int) bool {
		return b[i].Name < b[j].Name
	})

	// compare the arrays
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func endpointObjects(eps ...api.Endpoint) []Endpoint {
	var result []Endpoint
	for _, ep := range eps {
		result = append(result, Endpoint{
			Name:  api.MapWithEndpoint[ep].Name,
			Group: api.MapWithEndpoint[ep].Group,
		})
	}
	return result
}
