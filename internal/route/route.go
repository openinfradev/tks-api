package route

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/internal"
	delivery "github.com/openinfradev/tks-api/internal/delivery/http"
	"github.com/openinfradev/tks-api/internal/keycloak"
	"github.com/openinfradev/tks-api/internal/middleware/auth"
	"github.com/openinfradev/tks-api/internal/middleware/auth/authenticator"
	authKeycloak "github.com/openinfradev/tks-api/internal/middleware/auth/authenticator/keycloak"
	"github.com/openinfradev/tks-api/internal/middleware/auth/authorizer"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/internal/usecase"
	argowf "github.com/openinfradev/tks-api/pkg/argo-client"
	"github.com/openinfradev/tks-api/pkg/log"
	gcache "github.com/patrickmn/go-cache"
	"github.com/swaggo/http-swagger"
	"gorm.io/gorm"
)

var (
	API_PREFIX      = internal.API_PREFIX
	API_VERSION     = internal.API_VERSION
	ADMINAPI_PREFIX = internal.ADMINAPI_PREFIX

	SYSTEM_API_VERSION = internal.SYSTEM_API_VERSION
	SYSTEM_API_PREFIX  = internal.SYSTEM_API_PREFIX
)

type StatusRecorder struct {
	http.ResponseWriter
	Status int
}

func (r *StatusRecorder) WriteHeader(status int) {
	r.Status = status
	r.ResponseWriter.WriteHeader(status)
}

func SetupRouter(db *gorm.DB, argoClient argowf.ArgoClient, kc keycloak.IKeycloak, asset http.Handler) http.Handler {
	r := mux.NewRouter()

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
	}
	authMiddleware := auth.NewAuthMiddleware(
		authenticator.NewAuthenticator(authKeycloak.NewKeycloakAuthenticator(kc)),
		authorizer.NewDefaultAuthorization(repoFactory))

	cache := gcache.New(time.Hour, time.Hour)

	r.Use(loggingMiddleware)

	// [TODO] Transaction
	//r.Use(transactionMiddleware(db))

	authHandler := delivery.NewAuthHandler(usecase.NewAuthUsecase(repoFactory, kc))
	r.HandleFunc(API_PREFIX+API_VERSION+"/auth/login", authHandler.Login).Methods(http.MethodPost)
	r.Handle(API_PREFIX+API_VERSION+"/auth/logout", authMiddleware.Handle(http.HandlerFunc(authHandler.Logout))).Methods(http.MethodPost)
	r.Handle(API_PREFIX+API_VERSION+"/auth/refresh", authMiddleware.Handle(http.HandlerFunc(authHandler.RefreshToken))).Methods(http.MethodPost)
	r.HandleFunc(API_PREFIX+API_VERSION+"/auth/find-id/verification", authHandler.FindId).Methods(http.MethodPost)
	r.HandleFunc(API_PREFIX+API_VERSION+"/auth/find-password/verification", authHandler.FindPassword).Methods(http.MethodPost)
	r.HandleFunc(API_PREFIX+API_VERSION+"/auth/find-id/code", authHandler.VerifyIdentityForLostId).Methods(http.MethodPost)
	r.HandleFunc(API_PREFIX+API_VERSION+"/auth/find-password/code", authHandler.VerifyIdentityForLostPassword).Methods(http.MethodPost)
	//r.HandleFunc(API_PREFIX+API_VERSION+"/cookie-test", authHandler.CookieTest).Methods(http.MethodPost)
	//r.HandleFunc(API_PREFIX+API_VERSION+"/auth/callback", authHandler.CookieTestCallback).Methods(http.MethodGet)

	userHandler := delivery.NewUserHandler(usecase.NewUserUsecase(repoFactory, kc))
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/users", authMiddleware.Handle(http.HandlerFunc(userHandler.Create))).Methods(http.MethodPost)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/users", authMiddleware.Handle(http.HandlerFunc(userHandler.List))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/users/{accountId}", authMiddleware.Handle(http.HandlerFunc(userHandler.Get))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/users/{accountId}", authMiddleware.Handle(http.HandlerFunc(userHandler.Update))).Methods(http.MethodPut)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/users/{accountId}/reset-password", authMiddleware.Handle(http.HandlerFunc(userHandler.ResetPassword))).Methods(http.MethodPut)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/users/{accountId}", authMiddleware.Handle(http.HandlerFunc(userHandler.Delete))).Methods(http.MethodDelete)

	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/my-profile", authMiddleware.Handle(http.HandlerFunc(userHandler.GetMyProfile))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/my-profile", authMiddleware.Handle(http.HandlerFunc(userHandler.UpdateMyProfile))).Methods(http.MethodPut)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/my-profile/password", authMiddleware.Handle(http.HandlerFunc(userHandler.UpdateMyPassword))).Methods(http.MethodPut)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/my-profile/next-password-change", authMiddleware.Handle(http.HandlerFunc(userHandler.RenewPasswordExpiredDate))).Methods(http.MethodPut)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/my-profile", authMiddleware.Handle(http.HandlerFunc(userHandler.DeleteMyProfile))).Methods(http.MethodDelete)

	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/users/account-id/{accountId}/existence", authMiddleware.Handle(http.HandlerFunc(userHandler.CheckId))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/users/email/{email}/existence", authMiddleware.Handle(http.HandlerFunc(userHandler.CheckEmail))).Methods(http.MethodGet)

	organizationHandler := delivery.NewOrganizationHandler(usecase.NewOrganizationUsecase(repoFactory, argoClient, kc), usecase.NewUserUsecase(repoFactory, kc))
	r.Handle(API_PREFIX+API_VERSION+"/organizations", authMiddleware.Handle(http.HandlerFunc(organizationHandler.CreateOrganization))).Methods(http.MethodPost)
	r.Handle(API_PREFIX+API_VERSION+"/organizations", authMiddleware.Handle(http.HandlerFunc(organizationHandler.GetOrganizations))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}", authMiddleware.Handle(http.HandlerFunc(organizationHandler.GetOrganization))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}", authMiddleware.Handle(http.HandlerFunc(organizationHandler.DeleteOrganization))).Methods(http.MethodDelete)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}", authMiddleware.Handle(http.HandlerFunc(organizationHandler.UpdateOrganization))).Methods(http.MethodPut)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/primary-cluster", authMiddleware.Handle(http.HandlerFunc(organizationHandler.UpdatePrimaryCluster))).Methods(http.MethodPatch)

	clusterHandler := delivery.NewClusterHandler(usecase.NewClusterUsecase(repoFactory, argoClient, cache))
	r.Handle(API_PREFIX+API_VERSION+"/clusters", authMiddleware.Handle(http.HandlerFunc(clusterHandler.CreateCluster))).Methods(http.MethodPost)
	r.Handle(API_PREFIX+API_VERSION+"/clusters", authMiddleware.Handle(http.HandlerFunc(clusterHandler.GetClusters))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/clusters/{clusterId}", authMiddleware.Handle(http.HandlerFunc(clusterHandler.GetCluster))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/clusters/{clusterId}/site-values", authMiddleware.Handle(http.HandlerFunc(clusterHandler.GetClusterSiteValues))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/clusters/{clusterId}", authMiddleware.Handle(http.HandlerFunc(clusterHandler.DeleteCluster))).Methods(http.MethodDelete)

	appGroupHandler := delivery.NewAppGroupHandler(usecase.NewAppGroupUsecase(repoFactory, argoClient))
	r.Handle(API_PREFIX+API_VERSION+"/app-groups", authMiddleware.Handle(http.HandlerFunc(appGroupHandler.CreateAppGroup))).Methods(http.MethodPost)
	r.Handle(API_PREFIX+API_VERSION+"/app-groups", authMiddleware.Handle(http.HandlerFunc(appGroupHandler.GetAppGroups))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/app-groups/{appGroupId}", authMiddleware.Handle(http.HandlerFunc(appGroupHandler.GetAppGroup))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/app-groups/{appGroupId}", authMiddleware.Handle(http.HandlerFunc(appGroupHandler.DeleteAppGroup))).Methods(http.MethodDelete)
	r.Handle(API_PREFIX+API_VERSION+"/app-groups/{appGroupId}/applications", authMiddleware.Handle(http.HandlerFunc(appGroupHandler.GetApplications))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/app-groups/{appGroupId}/applications", authMiddleware.Handle(http.HandlerFunc(appGroupHandler.CreateApplication))).Methods(http.MethodPost)

	appServeAppHandler := delivery.NewAppServeAppHandler(usecase.NewAppServeAppUsecase(repoFactory, argoClient))
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/app-serve-apps", authMiddleware.Handle(http.HandlerFunc(appServeAppHandler.CreateAppServeApp))).Methods(http.MethodPost)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/app-serve-apps", authMiddleware.Handle(http.HandlerFunc(appServeAppHandler.GetAppServeApps))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/app-serve-apps/count", authMiddleware.Handle(http.HandlerFunc(appServeAppHandler.GetNumOfAppsOnStack))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/app-serve-apps/{appId}", authMiddleware.Handle(http.HandlerFunc(appServeAppHandler.GetAppServeApp))).Methods(http.MethodGet)
	// TODO: To be implemented
	//	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/app-serve-apps/{appId}/tasks/{taskId}", authMiddleware.Handle(http.HandlerFunc(appServeAppHandler.GetAppServeAppTask))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/app-serve-apps/{appId}/latest-task", authMiddleware.Handle(http.HandlerFunc(appServeAppHandler.GetAppServeAppLatestTask))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/app-serve-apps/{appId}/exist", authMiddleware.Handle(http.HandlerFunc(appServeAppHandler.IsAppServeAppExist))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/app-serve-apps/name/{name}/existence", authMiddleware.Handle(http.HandlerFunc(appServeAppHandler.IsAppServeAppNameExist))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/app-serve-apps/{appId}", authMiddleware.Handle(http.HandlerFunc(appServeAppHandler.DeleteAppServeApp))).Methods(http.MethodDelete)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/app-serve-apps/{appId}", authMiddleware.Handle(http.HandlerFunc(appServeAppHandler.UpdateAppServeApp))).Methods(http.MethodPut)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/app-serve-apps/{appId}/status", authMiddleware.Handle(http.HandlerFunc(appServeAppHandler.UpdateAppServeAppStatus))).Methods(http.MethodPatch)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/app-serve-apps/{appId}/endpoint", authMiddleware.Handle(http.HandlerFunc(appServeAppHandler.UpdateAppServeAppEndpoint))).Methods(http.MethodPatch)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/app-serve-apps/{appId}/rollback", authMiddleware.Handle(http.HandlerFunc(appServeAppHandler.RollbackAppServeApp))).Methods(http.MethodPost)

	cloudAccountHandler := delivery.NewCloudAccountHandler(usecase.NewCloudAccountUsecase(repoFactory, argoClient))
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/cloud-accounts", authMiddleware.Handle(http.HandlerFunc(cloudAccountHandler.GetCloudAccounts))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/cloud-accounts", authMiddleware.Handle(http.HandlerFunc(cloudAccountHandler.CreateCloudAccount))).Methods(http.MethodPost)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/cloud-accounts/name/{name}/existence", authMiddleware.Handle(http.HandlerFunc(cloudAccountHandler.CheckCloudAccountName))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/cloud-accounts/aws-account-id/{awsAccountId}/existence", authMiddleware.Handle(http.HandlerFunc(cloudAccountHandler.CheckAwsAccountId))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/cloud-accounts/{cloudAccountId}", authMiddleware.Handle(http.HandlerFunc(cloudAccountHandler.GetCloudAccount))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/cloud-accounts/{cloudAccountId}", authMiddleware.Handle(http.HandlerFunc(cloudAccountHandler.UpdateCloudAccount))).Methods(http.MethodPut)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/cloud-accounts/{cloudAccountId}", authMiddleware.Handle(http.HandlerFunc(cloudAccountHandler.DeleteCloudAccount))).Methods(http.MethodDelete)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/cloud-accounts/{cloudAccountId}/error", authMiddleware.Handle(http.HandlerFunc(cloudAccountHandler.DeleteForceCloudAccount))).Methods(http.MethodDelete)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/cloud-accounts/{cloudAccountId}/quotas", authMiddleware.Handle(http.HandlerFunc(cloudAccountHandler.GetResourceQuota))).Methods(http.MethodGet)

	stackTemplateHandler := delivery.NewStackTemplateHandler(usecase.NewStackTemplateUsecase(repoFactory))
	r.Handle(API_PREFIX+API_VERSION+"/stack-templates", authMiddleware.Handle(http.HandlerFunc(stackTemplateHandler.GetStackTemplates))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/stack-templates", authMiddleware.Handle(http.HandlerFunc(stackTemplateHandler.CreateStackTemplate))).Methods(http.MethodPost)
	r.Handle(API_PREFIX+API_VERSION+"/stack-templates/{stackTemplateId}", authMiddleware.Handle(http.HandlerFunc(stackTemplateHandler.GetStackTemplate))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/stack-templates/{stackTemplateId}", authMiddleware.Handle(http.HandlerFunc(stackTemplateHandler.UpdateStackTemplate))).Methods(http.MethodPut)
	r.Handle(API_PREFIX+API_VERSION+"/stack-templates/{stackTemplateId}", authMiddleware.Handle(http.HandlerFunc(stackTemplateHandler.DeleteStackTemplate))).Methods(http.MethodDelete)

	stackHandler := delivery.NewStackHandler(usecase.NewStackUsecase(repoFactory, argoClient))
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/stacks", authMiddleware.Handle(http.HandlerFunc(stackHandler.GetStacks))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/stacks", authMiddleware.Handle(http.HandlerFunc(stackHandler.CreateStack))).Methods(http.MethodPost)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/stacks/name/{name}/existence", authMiddleware.Handle(http.HandlerFunc(stackHandler.CheckStackName))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/stacks/{stackId}", authMiddleware.Handle(http.HandlerFunc(stackHandler.GetStack))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/stacks/{stackId}", authMiddleware.Handle(http.HandlerFunc(stackHandler.UpdateStack))).Methods(http.MethodPut)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/stacks/{stackId}", authMiddleware.Handle(http.HandlerFunc(stackHandler.DeleteStack))).Methods(http.MethodDelete)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/stacks/{stackId}/kube-config", authMiddleware.Handle(http.HandlerFunc(stackHandler.GetStackKubeConfig))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/stacks/{stackId}/status", authMiddleware.Handle(http.HandlerFunc(stackHandler.GetStackStatus))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/stacks/{stackId}/favorite", authMiddleware.Handle(http.HandlerFunc(stackHandler.SetFavorite))).Methods(http.MethodPost)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/stacks/{stackId}/favorite", authMiddleware.Handle(http.HandlerFunc(stackHandler.DeleteFavorite))).Methods(http.MethodDelete)

	dashboardHandler := delivery.NewDashboardHandler(usecase.NewDashboardUsecase(repoFactory, cache))
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/dashboard/charts", authMiddleware.Handle(http.HandlerFunc(dashboardHandler.GetCharts))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/dashboard/charts/{chartType}", authMiddleware.Handle(http.HandlerFunc(dashboardHandler.GetChart))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/dashboard/stacks", authMiddleware.Handle(http.HandlerFunc(dashboardHandler.GetStacks))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/dashboard/resources", authMiddleware.Handle(http.HandlerFunc(dashboardHandler.GetResources))).Methods(http.MethodGet)

	alertHandler := delivery.NewAlertHandler(usecase.NewAlertUsecase(repoFactory))
	r.HandleFunc(SYSTEM_API_PREFIX+SYSTEM_API_VERSION+"/alerts", alertHandler.CreateAlert).Methods(http.MethodPost)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/alerts", authMiddleware.Handle(http.HandlerFunc(alertHandler.GetAlerts))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/alerts/{alertId}", authMiddleware.Handle(http.HandlerFunc(alertHandler.GetAlert))).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/alerts/{alertId}", authMiddleware.Handle(http.HandlerFunc(alertHandler.DeleteAlert))).Methods(http.MethodDelete)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/alerts/{alertId}", authMiddleware.Handle(http.HandlerFunc(alertHandler.UpdateAlert))).Methods(http.MethodPut)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/alerts/{alertId}/actions", authMiddleware.Handle(http.HandlerFunc(alertHandler.CreateAlertAction))).Methods(http.MethodPost)
	//r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/alerts/{alertId}/actions/status", authMiddleware.Handle(http.HandlerFunc(alertHandler.UpdateAlertActionStatus))).Methods(http.MethodPatch)

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

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		r = r.WithContext(context.WithValue(ctx, internal.ContextKeyRequestID, uuid.New().String()))

		log.InfoWithContext(r.Context(), fmt.Sprintf("***** START [%s %s] ***** ", r.Method, r.RequestURI))

		body, err := io.ReadAll(r.Body)
		if err == nil {
			log.InfoWithContext(r.Context(), fmt.Sprintf("REQUEST BODY : %s", bytes.NewBuffer(body).String()))
		}
		r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

		next.ServeHTTP(w, r)

		log.InfofWithContext(r.Context(), "***** END [%s %s] *****", r.Method, r.RequestURI)
	})
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
