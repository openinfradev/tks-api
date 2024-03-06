package route

import (
	"net/http"
	"time"

	internalApi "github.com/openinfradev/tks-api/internal/delivery/api"
	"github.com/openinfradev/tks-api/internal/middleware/audit"
	"github.com/openinfradev/tks-api/internal/middleware/auth/requestRecoder"
	"github.com/openinfradev/tks-api/internal/middleware/logging"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/internal"
	delivery "github.com/openinfradev/tks-api/internal/delivery/http"
	"github.com/openinfradev/tks-api/internal/keycloak"
	internalMiddleware "github.com/openinfradev/tks-api/internal/middleware"
	"github.com/openinfradev/tks-api/internal/middleware/auth/authenticator"
	authKeycloak "github.com/openinfradev/tks-api/internal/middleware/auth/authenticator/keycloak"
	"github.com/openinfradev/tks-api/internal/middleware/auth/authorizer"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/internal/usecase"
	argowf "github.com/openinfradev/tks-api/pkg/argo-client"
	gcache "github.com/patrickmn/go-cache"
	httpSwagger "github.com/swaggo/http-swagger"
	"gorm.io/gorm"
)

var (
	API_PREFIX      = internal.API_PREFIX
	API_VERSION     = internal.API_VERSION
	ADMINAPI_PREFIX = internal.ADMINAPI_PREFIX

	SYSTEM_API_VERSION = internal.SYSTEM_API_VERSION
	SYSTEM_API_PREFIX  = internal.SYSTEM_API_PREFIX
)

func SetupRouter(db *gorm.DB, argoClient argowf.ArgoClient, kc keycloak.IKeycloak, asset http.Handler) http.Handler {
	r := mux.NewRouter()

	cache := gcache.New(5*time.Minute, 10*time.Minute)

	repoFactory := repository.Repository{
		Auth:          repository.NewAuthRepository(db),
		User:          repository.NewUserRepository(db),
		Cluster:       repository.NewClusterRepository(db),
		Organization:  repository.NewOrganizationRepository(db),
		AppGroup:      repository.NewAppGroupRepository(db),
		AppServeApp:   repository.NewAppServeAppRepository(db),
		CloudAccount:  repository.NewCloudAccountRepository(db),
		StackTemplate: repository.NewStackTemplateRepository(db),
		Alert:         repository.NewAlertRepository(db),
		Role:          repository.NewRoleRepository(db),
		Project:       repository.NewProjectRepository(db),
		Permission:    repository.NewPermissionRepository(db),
		Endpoint:      repository.NewEndpointRepository(db),
		Audit:         repository.NewAuditRepository(db),
	}

	usecaseFactory := usecase.Usecase{
		Auth:          usecase.NewAuthUsecase(repoFactory, kc),
		User:          usecase.NewUserUsecase(repoFactory, kc),
		Cluster:       usecase.NewClusterUsecase(repoFactory, argoClient, cache),
		Organization:  usecase.NewOrganizationUsecase(repoFactory, argoClient, kc),
		AppGroup:      usecase.NewAppGroupUsecase(repoFactory, argoClient),
		AppServeApp:   usecase.NewAppServeAppUsecase(repoFactory, argoClient),
		CloudAccount:  usecase.NewCloudAccountUsecase(repoFactory, argoClient),
		StackTemplate: usecase.NewStackTemplateUsecase(repoFactory),
		Dashboard:     usecase.NewDashboardUsecase(repoFactory, cache),
		Alert:         usecase.NewAlertUsecase(repoFactory),
		Stack:         usecase.NewStackUsecase(repoFactory, argoClient, usecase.NewDashboardUsecase(repoFactory, cache)),
		Project:       usecase.NewProjectUsecase(repoFactory, kc, argoClient),
		Audit:         usecase.NewAuditUsecase(repoFactory),
		Role:          usecase.NewRoleUsecase(repoFactory),
		Permission:    usecase.NewPermissionUsecase(repoFactory),
	}

	customMiddleware := internalMiddleware.NewMiddleware(
		authenticator.NewAuthenticator(authKeycloak.NewKeycloakAuthenticator(kc)),
		authorizer.NewDefaultAuthorization(repoFactory),
		requestRecoder.NewDefaultRequestRecoder(),
		audit.NewDefaultAudit(repoFactory))

	r.Use(logging.LoggingMiddleware)

	// [TODO] Transaction
	//r.Use(transactionMiddleware(db))

	authHandler := delivery.NewAuthHandler(usecaseFactory)
	r.HandleFunc(API_PREFIX+API_VERSION+"/auth/login", authHandler.Login).Methods(http.MethodPost)
	r.HandleFunc(API_PREFIX+API_VERSION+"/auth/ping", authHandler.PingToken).Methods(http.MethodPost)
	r.Handle(API_PREFIX+API_VERSION+"/auth/logout", customMiddleware.Handle(internalApi.Logout, http.HandlerFunc(authHandler.Logout))).Methods(http.MethodPost)
	r.Handle(API_PREFIX+API_VERSION+"/auth/refresh", customMiddleware.Handle(internalApi.RefreshToken, http.HandlerFunc(authHandler.RefreshToken))).Methods(http.MethodPost)
	r.HandleFunc(API_PREFIX+API_VERSION+"/auth/find-id/verification", authHandler.FindId).Methods(http.MethodPost)
	r.HandleFunc(API_PREFIX+API_VERSION+"/auth/find-password/verification", authHandler.FindPassword).Methods(http.MethodPost)
	r.HandleFunc(API_PREFIX+API_VERSION+"/auth/find-id/code", authHandler.VerifyIdentityForLostId).Methods(http.MethodPost)
	r.HandleFunc(API_PREFIX+API_VERSION+"/auth/find-password/code", authHandler.VerifyIdentityForLostPassword).Methods(http.MethodPost)
	r.Handle(API_PREFIX+API_VERSION+"/auth/verify-token", customMiddleware.Handle(internalApi.VerifyToken, http.HandlerFunc(authHandler.VerifyToken))).Methods(http.MethodGet)
	//r.HandleFunc(API_PREFIX+API_VERSION+"/cookie-test", authHandler.CookieTest).Methods(http.MethodPost)
	//r.HandleFunc(API_PREFIX+API_VERSION+"/auth/callback", authHandler.CookieTestCallback).Methods(http.MethodGet)

	userHandler := delivery.NewUserHandler(usecaseFactory)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/users", customMiddleware.Handle(internalApi.CreateUser, http.HandlerFunc(userHandler.Create))).Methods(http.MethodPost)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/users", customMiddleware.Handle(internalApi.ListUser, http.HandlerFunc(userHandler.List))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/users/{accountId}", customMiddleware.Handle(internalApi.GetUser, http.HandlerFunc(userHandler.Get))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/users/{accountId}", customMiddleware.Handle(internalApi.UpdateUser, http.HandlerFunc(userHandler.Update))).Methods(http.MethodPut)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/users/{accountId}/reset-password", customMiddleware.Handle(internalApi.ResetPassword, http.HandlerFunc(userHandler.ResetPassword))).Methods(http.MethodPut)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/users/{accountId}", customMiddleware.Handle(internalApi.DeleteUser, http.HandlerFunc(userHandler.Delete))).Methods(http.MethodDelete)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/users/account-id/{accountId}/existence", customMiddleware.Handle(internalApi.CheckId, http.HandlerFunc(userHandler.CheckId))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/users/email/{email}/existence", customMiddleware.Handle(internalApi.CheckEmail, http.HandlerFunc(userHandler.CheckEmail))).Methods(http.MethodGet)

	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/my-profile", customMiddleware.Handle(internalApi.GetMyProfile, http.HandlerFunc(userHandler.GetMyProfile))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/my-profile", customMiddleware.Handle(internalApi.UpdateMyProfile, http.HandlerFunc(userHandler.UpdateMyProfile))).Methods(http.MethodPut)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/my-profile/password", customMiddleware.Handle(internalApi.UpdateMyPassword, http.HandlerFunc(userHandler.UpdateMyPassword))).Methods(http.MethodPut)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/my-profile/next-password-change", customMiddleware.Handle(internalApi.RenewPasswordExpiredDate, http.HandlerFunc(userHandler.RenewPasswordExpiredDate))).Methods(http.MethodPut)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/my-profile", customMiddleware.Handle(internalApi.DeleteMyProfile, http.HandlerFunc(userHandler.DeleteMyProfile))).Methods(http.MethodDelete)

	// Admin
	r.Handle(API_PREFIX+API_VERSION+ADMINAPI_PREFIX+"/organizations/{organizationId}/users", customMiddleware.Handle(internalApi.Admin_CreateUser, http.HandlerFunc(userHandler.Admin_Create))).Methods(http.MethodPost)
	r.Handle(API_PREFIX+API_VERSION+ADMINAPI_PREFIX+"/organizations/{organizationId}/users", customMiddleware.Handle(internalApi.Admin_ListUser, http.HandlerFunc(userHandler.Admin_List))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+ADMINAPI_PREFIX+"/organizations/{organizationId}/users/{accountId}", customMiddleware.Handle(internalApi.Admin_GetUser, http.HandlerFunc(userHandler.Admin_Get))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+ADMINAPI_PREFIX+"/organizations/{organizationId}/users/{accountId}", customMiddleware.Handle(internalApi.Admin_UpdateUser, http.HandlerFunc(userHandler.Admin_Update))).Methods(http.MethodPut)
	r.Handle(API_PREFIX+API_VERSION+ADMINAPI_PREFIX+"/organizations/{organizationId}/users/{accountId}", customMiddleware.Handle(internalApi.Admin_DeleteUser, http.HandlerFunc(userHandler.Admin_Delete))).Methods(http.MethodDelete)

	organizationHandler := delivery.NewOrganizationHandler(usecaseFactory)
	r.Handle(API_PREFIX+API_VERSION+"/organizations", customMiddleware.Handle(internalApi.CreateOrganization, http.HandlerFunc(organizationHandler.CreateOrganization))).Methods(http.MethodPost)
	r.Handle(API_PREFIX+API_VERSION+"/organizations", customMiddleware.Handle(internalApi.GetOrganizations, http.HandlerFunc(organizationHandler.GetOrganizations))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}", customMiddleware.Handle(internalApi.GetOrganization, http.HandlerFunc(organizationHandler.GetOrganization))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}", customMiddleware.Handle(internalApi.DeleteOrganization, http.HandlerFunc(organizationHandler.DeleteOrganization))).Methods(http.MethodDelete)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}", customMiddleware.Handle(internalApi.UpdateOrganization, http.HandlerFunc(organizationHandler.UpdateOrganization))).Methods(http.MethodPut)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/primary-cluster", customMiddleware.Handle(internalApi.UpdatePrimaryCluster, http.HandlerFunc(organizationHandler.UpdatePrimaryCluster))).Methods(http.MethodPatch)

	clusterHandler := delivery.NewClusterHandler(usecaseFactory)
	r.Handle(API_PREFIX+API_VERSION+"/clusters", customMiddleware.Handle(internalApi.CreateCluster, http.HandlerFunc(clusterHandler.CreateCluster))).Methods(http.MethodPost)
	r.Handle(API_PREFIX+API_VERSION+"/clusters", customMiddleware.Handle(internalApi.GetClusters, http.HandlerFunc(clusterHandler.GetClusters))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/clusters/import", customMiddleware.Handle(internalApi.ImportCluster, http.HandlerFunc(clusterHandler.ImportCluster))).Methods(http.MethodPost)
	r.Handle(API_PREFIX+API_VERSION+"/clusters/{clusterId}", customMiddleware.Handle(internalApi.GetCluster, http.HandlerFunc(clusterHandler.GetCluster))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/clusters/{clusterId}", customMiddleware.Handle(internalApi.DeleteCluster, http.HandlerFunc(clusterHandler.DeleteCluster))).Methods(http.MethodDelete)
	r.Handle(API_PREFIX+API_VERSION+"/clusters/{clusterId}/site-values", customMiddleware.Handle(internalApi.GetClusterSiteValues, http.HandlerFunc(clusterHandler.GetClusterSiteValues))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/clusters/{clusterId}/install", customMiddleware.Handle(internalApi.InstallCluster, http.HandlerFunc(clusterHandler.InstallCluster))).Methods(http.MethodPost)
	r.Handle(API_PREFIX+API_VERSION+"/clusters/{clusterId}/bootstrap-kubeconfig", customMiddleware.Handle(internalApi.CreateBootstrapKubeconfig, http.HandlerFunc(clusterHandler.CreateBootstrapKubeconfig))).Methods(http.MethodPost)
	r.Handle(API_PREFIX+API_VERSION+"/clusters/{clusterId}/bootstrap-kubeconfig", customMiddleware.Handle(internalApi.GetBootstrapKubeconfig, http.HandlerFunc(clusterHandler.GetBootstrapKubeconfig))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/clusters/{clusterId}/nodes", customMiddleware.Handle(internalApi.GetNodes, http.HandlerFunc(clusterHandler.GetNodes))).Methods(http.MethodGet)

	appGroupHandler := delivery.NewAppGroupHandler(usecaseFactory)
	r.Handle(API_PREFIX+API_VERSION+"/app-groups", customMiddleware.Handle(internalApi.CreateAppgroup, http.HandlerFunc(appGroupHandler.CreateAppGroup))).Methods(http.MethodPost)
	r.Handle(API_PREFIX+API_VERSION+"/app-groups", customMiddleware.Handle(internalApi.GetAppgroups, http.HandlerFunc(appGroupHandler.GetAppGroups))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/app-groups/{appGroupId}", customMiddleware.Handle(internalApi.GetAppgroup, http.HandlerFunc(appGroupHandler.GetAppGroup))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/app-groups/{appGroupId}", customMiddleware.Handle(internalApi.DeleteAppgroup, http.HandlerFunc(appGroupHandler.DeleteAppGroup))).Methods(http.MethodDelete)
	r.Handle(API_PREFIX+API_VERSION+"/app-groups/{appGroupId}/applications", customMiddleware.Handle(internalApi.GetApplications, http.HandlerFunc(appGroupHandler.GetApplications))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/app-groups/{appGroupId}/applications", customMiddleware.Handle(internalApi.CreateApplication, http.HandlerFunc(appGroupHandler.CreateApplication))).Methods(http.MethodPost)

	appServeAppHandler := delivery.NewAppServeAppHandler(usecaseFactory)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/projects/{projectId}/app-serve-apps", customMiddleware.Handle(internalApi.CreateAppServeApp, http.HandlerFunc(appServeAppHandler.CreateAppServeApp))).Methods(http.MethodPost)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/projects/{projectId}/app-serve-apps", customMiddleware.Handle(internalApi.GetAppServeApps, http.HandlerFunc(appServeAppHandler.GetAppServeApps))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/projects/{projectId}/app-serve-apps/count", customMiddleware.Handle(internalApi.GetNumOfAppsOnStack, http.HandlerFunc(appServeAppHandler.GetNumOfAppsOnStack))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/projects/{projectId}/app-serve-apps/{appId}", customMiddleware.Handle(internalApi.GetAppServeApp, http.HandlerFunc(appServeAppHandler.GetAppServeApp))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/projects/{projectId}/app-serve-apps/{appId}/tasks", customMiddleware.Handle(internalApi.GetAppServeAppTasksByAppId, http.HandlerFunc(appServeAppHandler.GetAppServeAppTasksByAppId))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/projects/{projectId}/app-serve-apps/{appId}/tasks/{taskId}", customMiddleware.Handle(internalApi.GetAppServeAppTaskDetail, http.HandlerFunc(appServeAppHandler.GetAppServeAppTaskDetail))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/projects/{projectId}/app-serve-apps/{appId}/latest-task", customMiddleware.Handle(internalApi.GetAppServeAppLatestTask, http.HandlerFunc(appServeAppHandler.GetAppServeAppLatestTask))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/projects/{projectId}/app-serve-apps/{appId}/exist", customMiddleware.Handle(internalApi.IsAppServeAppExist, http.HandlerFunc(appServeAppHandler.IsAppServeAppExist))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/projects/{projectId}/app-serve-apps/name/{name}/existence", customMiddleware.Handle(internalApi.IsAppServeAppNameExist, http.HandlerFunc(appServeAppHandler.IsAppServeAppNameExist))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/projects/{projectId}/app-serve-apps/{appId}", customMiddleware.Handle(internalApi.DeleteAppServeApp, http.HandlerFunc(appServeAppHandler.DeleteAppServeApp))).Methods(http.MethodDelete)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/projects/{projectId}/app-serve-apps/{appId}", customMiddleware.Handle(internalApi.UpdateAppServeApp, http.HandlerFunc(appServeAppHandler.UpdateAppServeApp))).Methods(http.MethodPut)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/projects/{projectId}/app-serve-apps/{appId}/status", customMiddleware.Handle(internalApi.UpdateAppServeAppStatus, http.HandlerFunc(appServeAppHandler.UpdateAppServeAppStatus))).Methods(http.MethodPatch)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/projects/{projectId}/app-serve-apps/{appId}/endpoint", customMiddleware.Handle(internalApi.UpdateAppServeAppEndpoint, http.HandlerFunc(appServeAppHandler.UpdateAppServeAppEndpoint))).Methods(http.MethodPatch)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/projects/{projectId}/app-serve-apps/{appId}/rollback", customMiddleware.Handle(internalApi.RollbackAppServeApp, http.HandlerFunc(appServeAppHandler.RollbackAppServeApp))).Methods(http.MethodPost)

	cloudAccountHandler := delivery.NewCloudAccountHandler(usecaseFactory)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/cloud-accounts", customMiddleware.Handle(internalApi.GetCloudAccounts, http.HandlerFunc(cloudAccountHandler.GetCloudAccounts))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/cloud-accounts", customMiddleware.Handle(internalApi.CreateCloudAccount, http.HandlerFunc(cloudAccountHandler.CreateCloudAccount))).Methods(http.MethodPost)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/cloud-accounts/name/{name}/existence", customMiddleware.Handle(internalApi.CheckCloudAccountName, http.HandlerFunc(cloudAccountHandler.CheckCloudAccountName))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/cloud-accounts/aws-account-id/{awsAccountId}/existence", customMiddleware.Handle(internalApi.CheckAwsAccountId, http.HandlerFunc(cloudAccountHandler.CheckAwsAccountId))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/cloud-accounts/{cloudAccountId}", customMiddleware.Handle(internalApi.GetCloudAccount, http.HandlerFunc(cloudAccountHandler.GetCloudAccount))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/cloud-accounts/{cloudAccountId}", customMiddleware.Handle(internalApi.UpdateCloudAccount, http.HandlerFunc(cloudAccountHandler.UpdateCloudAccount))).Methods(http.MethodPut)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/cloud-accounts/{cloudAccountId}", customMiddleware.Handle(internalApi.DeleteCloudAccount, http.HandlerFunc(cloudAccountHandler.DeleteCloudAccount))).Methods(http.MethodDelete)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/cloud-accounts/{cloudAccountId}/error", customMiddleware.Handle(internalApi.DeleteForceCloudAccount, http.HandlerFunc(cloudAccountHandler.DeleteForceCloudAccount))).Methods(http.MethodDelete)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/cloud-accounts/{cloudAccountId}/quotas", customMiddleware.Handle(internalApi.GetResourceQuota, http.HandlerFunc(cloudAccountHandler.GetResourceQuota))).Methods(http.MethodGet)

	stackTemplateHandler := delivery.NewStackTemplateHandler(usecaseFactory)
	/* REMOVE START */
	r.Handle(API_PREFIX+API_VERSION+"/stack-templates", customMiddleware.Handle(internalApi.GetStackTemplates, http.HandlerFunc(stackTemplateHandler.GetStackTemplates))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/stack-templates", customMiddleware.Handle(internalApi.CreateStackTemplate, http.HandlerFunc(stackTemplateHandler.CreateStackTemplate))).Methods(http.MethodPost)
	r.Handle(API_PREFIX+API_VERSION+"/stack-templates/{stackTemplateId}", customMiddleware.Handle(internalApi.GetStackTemplate, http.HandlerFunc(stackTemplateHandler.GetStackTemplate))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/stack-templates/{stackTemplateId}", customMiddleware.Handle(internalApi.UpdateStackTemplate, http.HandlerFunc(stackTemplateHandler.UpdateStackTemplate))).Methods(http.MethodPut)
	r.Handle(API_PREFIX+API_VERSION+"/stack-templates/{stackTemplateId}", customMiddleware.Handle(internalApi.DeleteStackTemplate, http.HandlerFunc(stackTemplateHandler.DeleteStackTemplate))).Methods(http.MethodDelete)
	/* REMOVE END */

	r.Handle(API_PREFIX+API_VERSION+ADMINAPI_PREFIX+"/stack-templates", customMiddleware.Handle(internalApi.GetStackTemplates, http.HandlerFunc(stackTemplateHandler.GetStackTemplates))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+ADMINAPI_PREFIX+"/stack-templates", customMiddleware.Handle(internalApi.CreateStackTemplate, http.HandlerFunc(stackTemplateHandler.CreateStackTemplate))).Methods(http.MethodPost)
	r.Handle(API_PREFIX+API_VERSION+ADMINAPI_PREFIX+"/stack-templates/services", customMiddleware.Handle(internalApi.GetStackTemplates, http.HandlerFunc(stackTemplateHandler.GetStackTemplateServices))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+ADMINAPI_PREFIX+"/stack-templates/{stackTemplateId}", customMiddleware.Handle(internalApi.GetStackTemplates, http.HandlerFunc(stackTemplateHandler.GetStackTemplate))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+ADMINAPI_PREFIX+"/stack-templates/{stackTemplateId}/organizations", customMiddleware.Handle(internalApi.UpdateStackTemplateOrganizations, http.HandlerFunc(stackTemplateHandler.UpdateStackTemplateOrganizations))).Methods(http.MethodPut)

	dashboardHandler := delivery.NewDashboardHandler(usecaseFactory)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/dashboard/charts", customMiddleware.Handle(internalApi.GetChartsDashboard, http.HandlerFunc(dashboardHandler.GetCharts))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/dashboard/charts/{chartType}", customMiddleware.Handle(internalApi.GetChartDashboard, http.HandlerFunc(dashboardHandler.GetChart))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/dashboard/stacks", customMiddleware.Handle(internalApi.GetStacksDashboard, http.HandlerFunc(dashboardHandler.GetStacks))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/dashboard/resources", customMiddleware.Handle(internalApi.GetResourcesDashboard, http.HandlerFunc(dashboardHandler.GetResources))).Methods(http.MethodGet)

	alertHandler := delivery.NewAlertHandler(usecaseFactory)
	r.HandleFunc(SYSTEM_API_PREFIX+SYSTEM_API_VERSION+"/alerts", alertHandler.CreateAlert).Methods(http.MethodPost)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/alerts", customMiddleware.Handle(internalApi.GetAlerts, http.HandlerFunc(alertHandler.GetAlerts))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/alerts/{alertId}", customMiddleware.Handle(internalApi.GetAlert, http.HandlerFunc(alertHandler.GetAlert))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/alerts/{alertId}", customMiddleware.Handle(internalApi.DeleteAlert, http.HandlerFunc(alertHandler.DeleteAlert))).Methods(http.MethodDelete)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/alerts/{alertId}", customMiddleware.Handle(internalApi.UpdateAlert, http.HandlerFunc(alertHandler.UpdateAlert))).Methods(http.MethodPut)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/alerts/{alertId}/actions", customMiddleware.Handle(internalApi.CreateAlertAction, http.HandlerFunc(alertHandler.CreateAlertAction))).Methods(http.MethodPost)
	//r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/alerts/{alertId}/actions/status", customMiddleware.Handle(http.HandlerFunc(alertHandler.UpdateAlertActionStatus))).Methods(http.MethodPatch)

	stackHandler := delivery.NewStackHandler(usecaseFactory)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/stacks", customMiddleware.Handle(internalApi.GetStacks, http.HandlerFunc(stackHandler.GetStacks))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/stacks", customMiddleware.Handle(internalApi.CreateStack, http.HandlerFunc(stackHandler.CreateStack))).Methods(http.MethodPost)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/stacks/name/{name}/existence", customMiddleware.Handle(internalApi.CheckStackName, http.HandlerFunc(stackHandler.CheckStackName))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/stacks/{stackId}", customMiddleware.Handle(internalApi.GetStack, http.HandlerFunc(stackHandler.GetStack))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/stacks/{stackId}", customMiddleware.Handle(internalApi.UpdateStack, http.HandlerFunc(stackHandler.UpdateStack))).Methods(http.MethodPut)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/stacks/{stackId}", customMiddleware.Handle(internalApi.DeleteStack, http.HandlerFunc(stackHandler.DeleteStack))).Methods(http.MethodDelete)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/stacks/{stackId}/kube-config", customMiddleware.Handle(internalApi.GetStackKubeConfig, http.HandlerFunc(stackHandler.GetStackKubeConfig))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/stacks/{stackId}/status", customMiddleware.Handle(internalApi.GetStackStatus, http.HandlerFunc(stackHandler.GetStackStatus))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/stacks/{stackId}/favorite", customMiddleware.Handle(internalApi.SetFavoriteStack, http.HandlerFunc(stackHandler.SetFavorite))).Methods(http.MethodPost)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/stacks/{stackId}/favorite", customMiddleware.Handle(internalApi.DeleteFavoriteStack, http.HandlerFunc(stackHandler.DeleteFavorite))).Methods(http.MethodDelete)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/stacks/{stackId}/install", customMiddleware.Handle(internalApi.InstallStack, http.HandlerFunc(stackHandler.InstallStack))).Methods(http.MethodPost)

	projectHandler := delivery.NewProjectHandler(usecaseFactory)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/projects", customMiddleware.Handle(internalApi.CreateProject, http.HandlerFunc(projectHandler.CreateProject))).Methods(http.MethodPost)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/projects", customMiddleware.Handle(internalApi.GetProjects, http.HandlerFunc(projectHandler.GetProjects))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/projects/project-roles", customMiddleware.Handle(internalApi.GetProjectRoles, http.HandlerFunc(projectHandler.GetProjectRoles))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/projects/project-roles/{projectRoleId}", customMiddleware.Handle(internalApi.GetProjectRole, http.HandlerFunc(projectHandler.GetProjectRole))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/projects/existence", customMiddleware.Handle(internalApi.GetProjectNamespace, http.HandlerFunc(projectHandler.IsProjectNameExist))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/projects/{projectId}", customMiddleware.Handle(internalApi.GetProject, http.HandlerFunc(projectHandler.GetProject))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/projects/{projectId}", customMiddleware.Handle(internalApi.UpdateProject, http.HandlerFunc(projectHandler.UpdateProject))).Methods(http.MethodPut)
	//r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/projects/{projectId}", customMiddleware.Handle(internalApi.DeleteProject, http.HandlerFunc(projectHandler.DeleteProject))).Methods(http.MethodDelete)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/projects/{projectId}/members", customMiddleware.Handle(internalApi.AddProjectMember, http.HandlerFunc(projectHandler.AddProjectMember))).Methods(http.MethodPost)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/projects/{projectId}/members/count", customMiddleware.Handle(internalApi.GetProjectMembers, http.HandlerFunc(projectHandler.GetProjectMemberCount))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/projects/{projectId}/members/{projectMemberId}", customMiddleware.Handle(internalApi.GetProjectMember, http.HandlerFunc(projectHandler.GetProjectMember))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/projects/{projectId}/members", customMiddleware.Handle(internalApi.GetProjectMembers, http.HandlerFunc(projectHandler.GetProjectMembers))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/projects/{projectId}/members/{projectMemberId}", customMiddleware.Handle(internalApi.RemoveProjectMember, http.HandlerFunc(projectHandler.RemoveProjectMember))).Methods(http.MethodDelete)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/projects/{projectId}/members", customMiddleware.Handle(internalApi.RemoveProjectMember, http.HandlerFunc(projectHandler.RemoveProjectMembers))).Methods(http.MethodDelete)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/projects/{projectId}/members", customMiddleware.Handle(internalApi.UpdateProjectMemberRole, http.HandlerFunc(projectHandler.UpdateProjectMembersRole))).Methods(http.MethodPut)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/projects/{projectId}/members/{projectMemberId}/role", customMiddleware.Handle(internalApi.UpdateProjectMemberRole, http.HandlerFunc(projectHandler.UpdateProjectMemberRole))).Methods(http.MethodPut)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/projects/{projectId}/namespaces", customMiddleware.Handle(internalApi.CreateProjectNamespace, http.HandlerFunc(projectHandler.CreateProjectNamespace))).Methods(http.MethodPost)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/projects/{projectId}/namespaces/{projectNamespace}/stacks/{stackId}/existence", customMiddleware.Handle(internalApi.GetProjectNamespace, http.HandlerFunc(projectHandler.IsProjectNamespaceExist))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/projects/{projectId}/namespaces/{projectNamespace}/stacks/{stackId}/k8s-resources", customMiddleware.Handle(internalApi.GetProjectNamespaceK8sResources, http.HandlerFunc(projectHandler.GetProjectNamespaceK8sResources))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/projects/{projectId}/namespaces/{projectNamespace}/stacks/{stackId}/resources-usage", customMiddleware.Handle(internalApi.GetProjectNamespaceK8sResources, http.HandlerFunc(projectHandler.GetProjectNamespaceResourcesUsage))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/projects/{projectId}/namespaces", customMiddleware.Handle(internalApi.GetProjectNamespaces, http.HandlerFunc(projectHandler.GetProjectNamespaces))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/projects/{projectId}/namespaces/{projectNamespace}/stacks/{stackId}", customMiddleware.Handle(internalApi.GetProjectNamespace, http.HandlerFunc(projectHandler.GetProjectNamespace))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/projects/{projectId}/namespaces/{projectNamespace}/stacks/{stackId}", customMiddleware.Handle(internalApi.UpdateProjectNamespace, http.HandlerFunc(projectHandler.UpdateProjectNamespace))).Methods(http.MethodPut)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/projects/{projectId}/namespaces/{projectNamespace}/stacks/{stackId}", customMiddleware.Handle(internalApi.DeleteProjectNamespace, http.HandlerFunc(projectHandler.DeleteProjectNamespace))).Methods(http.MethodDelete)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/projects/{projectId}/kubeconfig", customMiddleware.Handle(internalApi.GetProjectKubeconfig, http.HandlerFunc(projectHandler.GetProjectKubeconfig))).Methods(http.MethodGet)

	auditHandler := delivery.NewAuditHandler(usecaseFactory)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/audits", customMiddleware.Handle(internalApi.GetAudits, http.HandlerFunc(auditHandler.GetAudits))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/audits/{auditId}", customMiddleware.Handle(internalApi.GetAudit, http.HandlerFunc(auditHandler.GetAudit))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/audits/{auditId}", customMiddleware.Handle(internalApi.DeleteAudit, http.HandlerFunc(auditHandler.DeleteAudit))).Methods(http.MethodDelete)

	roleHandler := delivery.NewRoleHandler(usecaseFactory)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/roles", customMiddleware.Handle(internalApi.CreateTksRole, http.HandlerFunc(roleHandler.CreateTksRole))).Methods(http.MethodPost)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/roles", customMiddleware.Handle(internalApi.ListTksRoles, http.HandlerFunc(roleHandler.ListTksRoles))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/roles/{roleId}", customMiddleware.Handle(internalApi.GetTksRole, http.HandlerFunc(roleHandler.GetTksRole))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/roles/{roleId}", customMiddleware.Handle(internalApi.DeleteTksRole, http.HandlerFunc(roleHandler.DeleteTksRole))).Methods(http.MethodDelete)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/roles/{roleId}", customMiddleware.Handle(internalApi.UpdateTksRole, http.HandlerFunc(roleHandler.UpdateTksRole))).Methods(http.MethodPut)

	// Admin
	r.Handle(API_PREFIX+API_VERSION+ADMINAPI_PREFIX+"/organizations/{organizationId}/roles", customMiddleware.Handle(internalApi.Admin_ListTksRoles, http.HandlerFunc(roleHandler.Admin_ListTksRoles))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+ADMINAPI_PREFIX+"/organizations/{organizationId}/roles/{roleId}", customMiddleware.Handle(internalApi.Admin_GetTksRole, http.HandlerFunc(roleHandler.Admin_GetTksRole))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+ADMINAPI_PREFIX+"/organizations/{organizationId}/projects", customMiddleware.Handle(internalApi.Admin_GetProjects, http.HandlerFunc(projectHandler.Admin_GetProjects))).Methods(http.MethodGet)

	permissionHandler := delivery.NewPermissionHandler(usecaseFactory)
	r.Handle(API_PREFIX+API_VERSION+"/permissions/templates", customMiddleware.Handle(internalApi.GetPermissionTemplates, http.HandlerFunc(permissionHandler.GetPermissionTemplates))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/roles/{roleId}/permissions", customMiddleware.Handle(internalApi.GetPermissionsByRoleId, http.HandlerFunc(permissionHandler.GetPermissionsByRoleId))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/roles/{roleId}/permissions", customMiddleware.Handle(internalApi.UpdatePermissionsByRoleId, http.HandlerFunc(permissionHandler.UpdatePermissionsByRoleId))).Methods(http.MethodPut)

	r.HandleFunc(API_PREFIX+API_VERSION+"/alerttest", alertHandler.CreateAlert).Methods(http.MethodPost)
	// assets
	r.PathPrefix("/api/").HandlerFunc(http.NotFound)
	r.PathPrefix("/").Handler(httpSwagger.WrapHandler).Methods(http.MethodGet)

	//withLog := handlers.LoggingHandler(os.Stdout, r)

	credentials := handlers.AllowCredentials()
	headersOk := handlers.AllowedHeaders([]string{"content-type", "Authorization", "Authorization-Type"})
	originsOk := handlers.AllowedOrigins([]string{"http://localhost:3000"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"})

	return handlers.CORS(credentials, headersOk, originsOk, methodsOk)(r)
}

/*
func transactionMiddleware(db *gorm.DB) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			txHandle := db.Begin()
			log.DebugWithContext(r.Context(),"beginning database transaction")

			defer func() {
				if r := recover(); r != nil {
					txHandle.Rollback()
				}
			}()

			recorder := &StatusRecorder{
				ResponseWriter: w,
				Status:         200,
			}

			r = r.WithContext(context.WithValue(ctx, "txHandle", txHandle))
			next.ServeHTTP(recorder, r)

			if StatusInList(recorder.Status, []int{http.StatusOK}) {
				log.DebugWithContext(r.Context(),"committing transactions")
				if err := txHandle.Commit().Error; err != nil {
					log.DebugWithContext(r.Context(),"trx commit error: ", err)
				}
			} else {
				log.DebugWithContext(r.Context(),"rolling back transaction due to status code: ", recorder.Status)
				txHandle.Rollback()
			}
		})
	}
}

// StatusInList -> checks if the given status is in the list
func StatusInList(status int, statusList []int) bool {
	for _, i := range statusList {
		if i == status {
			return true
		}
	}
	return false
}
*/
