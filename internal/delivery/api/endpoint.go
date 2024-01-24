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
	PingToken
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
	UpdateUser
	ResetPassword
	CheckId
	CheckEmail

	// MyProfile
	GetMyProfile
	UpdateMyProfile
	UpdateMyPassword
	RenewPasswordExpiredDate
	DeleteMyProfile

	// Organization
	CreateOrganization
	GetOrganizations
	GetOrganization
	DeleteOrganization
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
	CreateAppServeApp
	GetAppServeApps
	GetNumOfAppsOnStack
	GetAppServeApp
	GetAppServeAppLatestTask
	IsAppServeAppExist
	IsAppServeAppNameExist
	DeleteAppServeApp
	UpdateAppServeApp
	UpdateAppServeAppStatus
	UpdateAppServeAppEndpoint
	RollbackAppServeApp

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
	GetStackTemplates
	CreateStackTemplate
	GetStackTemplate
	UpdateStackTemplate
	DeleteStackTemplate

	// Dashboard
	GetChartsDashboard
	GetChartDashboard
	GetStacksDashboard
	GetResourcesDashboard

	// Alert
	CreateAlert
	GetAlerts
	GetAlert
	DeleteAlert
	UpdateAlert
	CreateAlertAction

	// Stack
	GetStacks
	CreateStack
	CheckStackName
	GetStack
	UpdateStack
	DeleteStack
	GetStackKubeConfig
	GetStackStatus
	SetFavoriteStack
	DeleteFavoriteStack
	InstallStack

	// Project
	CreateProject
	GetProjects
	GetProject
	UpdateProject
	DeleteProject
	AddProjectMember
	GetProjectMembers
	RemoveProjectMember
	UpdateProjectMemberRole
	CreateProjectNamespace
	GetProjectNamespaces
	GetProjectNamespace
	DeleteProjectNamespace
	SetFavoriteProject
	SetFavoriteProjectNamespace
	UnSetFavoriteProject
	UnSetFavoriteProjectNamespace
)

var ApiMap = map[Endpoint]EndpointInfo{
	Login: {
		Name:  "Login",
		Group: "Auth",
	},
	PingToken: {
		Name:  "PingToken",
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
	CreateOrganization: {
		Name:  "CreateOrganization",
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
	DeleteOrganization: {
		Name:  "DeleteOrganization",
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
	GetStackTemplates: {
		Name:  "GetStackTemplates",
		Group: "StackTemplate",
	},
	CreateStackTemplate: {
		Name:  "CreateStackTemplate",
		Group: "StackTemplate",
	},
	GetStackTemplate: {
		Name:  "GetStackTemplate",
		Group: "StackTemplate",
	},
	UpdateStackTemplate: {
		Name:  "UpdateStackTemplate",
		Group: "StackTemplate",
	},
	DeleteStackTemplate: {
		Name:  "DeleteStackTemplate",
		Group: "StackTemplate",
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
	CreateAlert: {
		Name:  "CreateAlert",
		Group: "Alert",
	},
	GetAlerts: {
		Name:  "GetAlerts",
		Group: "Alert",
	},
	GetAlert: {
		Name:  "GetAlert",
		Group: "Alert",
	},
	DeleteAlert: {
		Name:  "DeleteAlert",
		Group: "Alert",
	},
	UpdateAlert: {
		Name:  "UpdateAlert",
		Group: "Alert",
	},
	CreateAlertAction: {
		Name:  "CreateAlertAction",
		Group: "Alert",
	},
	GetStacks: {
		Name:  "GetStacks",
		Group: "Stack",
	},
	CreateStack: {
		Name:  "CreateStack",
		Group: "Stack",
	},
	CheckStackName: {
		Name:  "CheckStackName",
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
	GetStackKubeConfig: {
		Name:  "GetStackKubeConfig",
		Group: "Stack",
	},
	GetStackStatus: {
		Name:  "GetStackStatus",
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
}

func (e Endpoint) String() string {
	switch e {
	case Login:
		return "Login"
	case PingToken:
		return "PingToken"
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
	case CreateOrganization:
		return "CreateOrganization"
	case GetOrganizations:
		return "GetOrganizations"
	case GetOrganization:
		return "GetOrganization"
	case DeleteOrganization:
		return "DeleteOrganization"
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
	case GetStackTemplates:
		return "GetStackTemplates"
	case CreateStackTemplate:
		return "CreateStackTemplate"
	case GetStackTemplate:
		return "GetStackTemplate"
	case UpdateStackTemplate:
		return "UpdateStackTemplate"
	case DeleteStackTemplate:
		return "DeleteStackTemplate"
	case GetChartsDashboard:
		return "GetChartsDashboard"
	case GetChartDashboard:
		return "GetChartDashboard"
	case GetStacksDashboard:
		return "GetStacksDashboard"
	case GetResourcesDashboard:
		return "GetResourcesDashboard"
	case CreateAlert:
		return "CreateAlert"
	case GetAlerts:
		return "GetAlerts"
	case GetAlert:
		return "GetAlert"
	case DeleteAlert:
		return "DeleteAlert"
	case UpdateAlert:
		return "UpdateAlert"
	case CreateAlertAction:
		return "CreateAlertAction"
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
	default:
		return ""
	}
}
func GetEndpoint(name string) Endpoint {
	switch name {
	case "Login":
		return Login
	case "PingToken":
		return PingToken
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
	case "CreateOrganization":
		return CreateOrganization
	case "GetOrganizations":
		return GetOrganizations
	case "GetOrganization":
		return GetOrganization
	case "DeleteOrganization":
		return DeleteOrganization
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
	case "GetStackTemplates":
		return GetStackTemplates
	case "CreateStackTemplate":
		return CreateStackTemplate
	case "GetStackTemplate":
		return GetStackTemplate
	case "UpdateStackTemplate":
		return UpdateStackTemplate
	case "DeleteStackTemplate":
		return DeleteStackTemplate
	case "GetChartsDashboard":
		return GetChartsDashboard
	case "GetChartDashboard":
		return GetChartDashboard
	case "GetStacksDashboard":
		return GetStacksDashboard
	case "GetResourcesDashboard":
		return GetResourcesDashboard
	case "CreateAlert":
		return CreateAlert
	case "GetAlerts":
		return GetAlerts
	case "GetAlert":
		return GetAlert
	case "DeleteAlert":
		return DeleteAlert
	case "UpdateAlert":
		return UpdateAlert
	case "CreateAlertAction":
		return CreateAlertAction
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
	default:
		return -1
	}
}
