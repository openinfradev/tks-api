package route

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/auth/request"
	user "github.com/openinfradev/tks-api/internal/auth/user"
	"github.com/openinfradev/tks-api/internal/keycloak"

	jwtWithouKey "github.com/dgrijalva/jwt-go"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"gorm.io/gorm"

	"github.com/swaggo/http-swagger"

	delivery "github.com/openinfradev/tks-api/internal/delivery/http"
	"github.com/openinfradev/tks-api/internal/helper"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/internal/usecase"
	argowf "github.com/openinfradev/tks-api/pkg/argo-client"
	"github.com/openinfradev/tks-api/pkg/log"
)

const (
	API_VERSION = "/1.0"
	API_PREFIX  = "/api"
)

type StatusRecorder struct {
	http.ResponseWriter
	Status int
}

func (r *StatusRecorder) WriteHeader(status int) {
	r.Status = status
	r.ResponseWriter.WriteHeader(status)
}

func SetupRouter(db *gorm.DB, argoClient argowf.ArgoClient, asset http.Handler, kc keycloak.IKeycloak) http.Handler {
	r := mux.NewRouter()

	r.Use(loggingMiddleware)

	// [TODO] Transaction
	//r.Use(transactionMiddleware(db))

	r.PathPrefix("/swagger").Handler(httpSwagger.WrapHandler)

	repoFactory := repository.Repository{
		User:          repository.NewUserRepository(db),
		Cluster:       repository.NewClusterRepository(db),
		Organization:  repository.NewOrganizationRepository(db),
		AppGroup:      repository.NewAppGroupRepository(db),
		AppServeApp:   repository.NewAppServeAppRepository(db),
		CloudSetting:  repository.NewCloudSettingRepository(db),
		StackTemplate: repository.NewStackTemplateRepository(db),
		History:       repository.NewHistoryRepository(db),
	}

	authHandler := delivery.NewAuthHandler(usecase.NewAuthUsecase(repoFactory, kc))
	r.HandleFunc(API_PREFIX+API_VERSION+"/auth/login", authHandler.Login).Methods(http.MethodPost)
	r.HandleFunc(API_PREFIX+API_VERSION+"/auth/logout", authHandler.Logout).Methods(http.MethodPost)
	r.HandleFunc(API_PREFIX+API_VERSION+"/auth/find-id", authHandler.FindId).Methods(http.MethodPost)
	r.HandleFunc(API_PREFIX+API_VERSION+"/auth/find-password", authHandler.FindPassword).Methods(http.MethodPost)

	userHandler := delivery.NewUserHandler(usecase.NewUserUsecase(repoFactory, kc))
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/users", authMiddleware(http.HandlerFunc(userHandler.Create), kc)).Methods(http.MethodPost)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/users", authMiddleware(http.HandlerFunc(userHandler.List), kc)).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/users/{accountId}", authMiddleware(http.HandlerFunc(userHandler.Get), kc)).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/users/{accountId}", authMiddleware(http.HandlerFunc(userHandler.Delete), kc)).Methods(http.MethodDelete)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/users/{accountId}", authMiddleware(http.HandlerFunc(userHandler.Update), kc)).Methods(http.MethodPut)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/users/{accountId}/password", authMiddleware(http.HandlerFunc(userHandler.UpdatePassword), kc)).Methods(http.MethodPut)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/users/{accountId}/existence", authMiddleware(http.HandlerFunc(userHandler.CheckId), kc)).Methods(http.MethodGet)

	organizationHandler := delivery.NewOrganizationHandler(usecase.NewOrganizationUsecase(repoFactory, argoClient, kc), usecase.NewUserUsecase(repoFactory, kc))
	r.Handle(API_PREFIX+API_VERSION+"/organizations", authMiddleware(http.HandlerFunc(organizationHandler.CreateOrganization), kc)).Methods(http.MethodPost)
	r.Handle(API_PREFIX+API_VERSION+"/organizations", authMiddleware(http.HandlerFunc(organizationHandler.GetOrganizations), kc)).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}", authMiddleware(http.HandlerFunc(organizationHandler.GetOrganization), kc)).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}", authMiddleware(http.HandlerFunc(organizationHandler.DeleteOrganization), kc)).Methods(http.MethodDelete)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}", authMiddleware(http.HandlerFunc(organizationHandler.UpdateOrganization), kc)).Methods(http.MethodPut)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/primary-cluster", authMiddleware(http.HandlerFunc(organizationHandler.UpdatePrimaryCluster), kc)).Methods(http.MethodPatch)

	clusterHandler := delivery.NewClusterHandler(usecase.NewClusterUsecase(repoFactory, argoClient))
	r.Handle(API_PREFIX+API_VERSION+"/clusters", authMiddleware(http.HandlerFunc(clusterHandler.CreateCluster), kc)).Methods(http.MethodPost)
	r.Handle(API_PREFIX+API_VERSION+"/clusters", authMiddleware(http.HandlerFunc(clusterHandler.GetClusters), kc)).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/clusters/{clusterId}", authMiddleware(http.HandlerFunc(clusterHandler.GetCluster), kc)).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/clusters/{clusterId}", authMiddleware(http.HandlerFunc(clusterHandler.DeleteCluster), kc)).Methods(http.MethodDelete)

	appGroupHandler := delivery.NewAppGroupHandler(usecase.NewAppGroupUsecase(repoFactory, argoClient))
	r.Handle(API_PREFIX+API_VERSION+"/app-groups", authMiddleware(http.HandlerFunc(appGroupHandler.CreateAppGroup), kc)).Methods(http.MethodPost)
	r.Handle(API_PREFIX+API_VERSION+"/app-groups", authMiddleware(http.HandlerFunc(appGroupHandler.GetAppGroups), kc)).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/app-groups/{appGroupId}", authMiddleware(http.HandlerFunc(appGroupHandler.GetAppGroup), kc)).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/app-groups/{appGroupId}", authMiddleware(http.HandlerFunc(appGroupHandler.DeleteAppGroup), kc)).Methods(http.MethodDelete)
	r.Handle(API_PREFIX+API_VERSION+"/app-groups/{appGroupId}/applications", authMiddleware(http.HandlerFunc(appGroupHandler.GetApplications), kc)).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/app-groups/{appGroupId}/applications", authMiddleware(http.HandlerFunc(appGroupHandler.UpdateApplication), kc)).Methods(http.MethodPost)

	appServeAppHandler := delivery.NewAppServeAppHandler(usecase.NewAppServeAppUsecase(repoFactory, argoClient))
	r.Handle(API_PREFIX+API_VERSION+"/app-serve-apps", authMiddleware(http.HandlerFunc(appServeAppHandler.CreateAppServeApp), kc)).Methods(http.MethodPost)
	r.Handle(API_PREFIX+API_VERSION+"/app-serve-apps", authMiddleware(http.HandlerFunc(appServeAppHandler.GetAppServeApps), kc)).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/app-serve-apps/{appId}", authMiddleware(http.HandlerFunc(appServeAppHandler.GetAppServeApp), kc)).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/app-serve-apps/{appId}", authMiddleware(http.HandlerFunc(appServeAppHandler.DeleteAppServeApp), kc)).Methods(http.MethodDelete)
	r.Handle(API_PREFIX+API_VERSION+"/app-serve-apps/{appId}", authMiddleware(http.HandlerFunc(appServeAppHandler.UpdateAppServeApp), kc)).Methods(http.MethodPut)

	cloudSettingHandler := delivery.NewCloudSettingHandler(usecase.NewCloudSettingUsecase(repoFactory, argoClient))
	r.Handle(API_PREFIX+API_VERSION+"/cloud-settings", authMiddleware(http.HandlerFunc(cloudSettingHandler.GetCloudSettings), kc)).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/cloud-settings", authMiddleware(http.HandlerFunc(cloudSettingHandler.CreateCloudSetting), kc)).Methods(http.MethodPost)
	r.Handle(API_PREFIX+API_VERSION+"/cloud-settings/name/{name}/existence", authMiddleware(http.HandlerFunc(cloudSettingHandler.CheckCloudSettingName), kc)).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/cloud-settings/{cloudSettingId}", authMiddleware(http.HandlerFunc(cloudSettingHandler.GetCloudSetting), kc)).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/cloud-settings/{cloudSettingId}", authMiddleware(http.HandlerFunc(cloudSettingHandler.UpdateCloudSetting), kc)).Methods(http.MethodPut)
	r.Handle(API_PREFIX+API_VERSION+"/cloud-settings/{cloudSettingId}", authMiddleware(http.HandlerFunc(cloudSettingHandler.DeleteCloudSetting), kc)).Methods(http.MethodDelete)

	stackTemplateHandler := delivery.NewStackTemplateHandler(usecase.NewStackTemplateUsecase(repoFactory))
	r.Handle(API_PREFIX+API_VERSION+"/stack-templates", authMiddleware(http.HandlerFunc(stackTemplateHandler.GetStackTemplates), kc)).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/stack-templates", authMiddleware(http.HandlerFunc(stackTemplateHandler.CreateStackTemplate), kc)).Methods(http.MethodPost)
	r.Handle(API_PREFIX+API_VERSION+"/stack-templates/{stackTemplateId}", authMiddleware(http.HandlerFunc(stackTemplateHandler.GetStackTemplate), kc)).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/stack-templates/{stackTemplateId}", authMiddleware(http.HandlerFunc(stackTemplateHandler.UpdateStackTemplate), kc)).Methods(http.MethodPut)
	r.Handle(API_PREFIX+API_VERSION+"/stack-templates/{stackTemplateId}", authMiddleware(http.HandlerFunc(stackTemplateHandler.DeleteStackTemplate), kc)).Methods(http.MethodDelete)

	stackHandler := delivery.NewStackHandler(usecase.NewStackUsecase(repoFactory, argoClient))
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/stacks", authMiddleware(http.HandlerFunc(stackHandler.GetStacks), kc)).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/stacks", authMiddleware(http.HandlerFunc(stackHandler.CreateStack), kc)).Methods(http.MethodPost)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/stacks/name/{name}/existence", authMiddleware(http.HandlerFunc(stackHandler.CheckStackName), kc)).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/stacks/{stackId}", authMiddleware(http.HandlerFunc(stackHandler.GetStack), kc)).Methods(http.MethodGet)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/stacks/{stackId}", authMiddleware(http.HandlerFunc(stackHandler.UpdateStack), kc)).Methods(http.MethodPut)
	r.Handle(API_PREFIX+API_VERSION+"/organizations/{organizationId}/stacks/{stackId}", authMiddleware(http.HandlerFunc(stackHandler.DeleteStack), kc)).Methods(http.MethodDelete)

	// assets
	r.PathPrefix("/api/").HandlerFunc(http.NotFound)
	r.PathPrefix("/").Handler(asset).Methods(http.MethodGet)

	//withLog := handlers.LoggingHandler(os.Stdout, r)

	credentials := handlers.AllowCredentials()
	headersOk := handlers.AllowedHeaders([]string{"content-type", "Authorization", "Authorization-Type"})
	originsOk := handlers.AllowedOrigins([]string{"http://localhost:3000"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"})

	return handlers.CORS(credentials, headersOk, originsOk, methodsOk)(r)
}

func authMiddleware(next http.Handler, kc keycloak.IKeycloak) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Possible values : "basic", "keycloak"
		authType := r.Header.Get("Authorization-Type")

		switch authType {
		case "basic":
			tokenString := r.Header.Get("Authorization")
			if len(tokenString) == 0 {
				w.WriteHeader(http.StatusUnauthorized)
				if _, err := w.Write([]byte("Missing Authorization Header")); err != nil {
					log.Error(err)
				}

				return
			}
			tokenString = strings.Replace(tokenString, "Bearer ", "", 1)
			token, err := helper.VerifyToken(tokenString)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				if _, err := w.Write([]byte("Error verifying JWT token: " + err.Error())); err != nil {
					log.Error(err)
				}
				return
			}

			accountId := token.Claims.(jwt.MapClaims)["AccountId"]
			organizationId := token.Claims.(jwt.MapClaims)["OrganizationId"]
			id := token.Claims.(jwt.MapClaims)["ID"]

			log.Debug("[authMiddleware] accountId : ", accountId)
			log.Debug("[authMiddleware] Id : ", id)
			log.Debug("[authMiddleware] organizationId : ", organizationId)

			r.Header.Set("OrganizationId", fmt.Sprint(organizationId))
			r.Header.Set("AccountId", fmt.Sprint(accountId))
			r.Header.Set("ID", fmt.Sprint(id))

			next.ServeHTTP(w, r)
			return

		case "keycloak":
		default:
			auth := strings.TrimSpace(r.Header.Get("Authorization"))
			if auth == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			parts := strings.SplitN(auth, " ", 3)
			if len(parts) < 2 || strings.ToLower(parts[0]) != "bearer" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			token := parts[1]

			// Empty bearer tokens aren't valid
			if len(token) == 0 {
				// The space before the token case
				if len(parts) == 3 {
					log.Warn("the provided Authorization header contains extra space before the bearer token, and is ignored")
				}
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			parsedToken, _, err := new(jwtWithouKey.Parser).ParseUnverified(token, jwtWithouKey.MapClaims{})

			if err != nil {
				log.Error("failed to parse access token: ", err)
				w.WriteHeader(http.StatusUnauthorized)
				if _, err := w.Write([]byte(err.Error())); err != nil {
					log.Error(err)
				}
				return
			}
			organization := parsedToken.Claims.(jwtWithouKey.MapClaims)["organization"].(string)
			if err := kc.VerifyAccessToken(token, organization); err != nil {
				log.Error("failed to verify access token: ", err)
				w.WriteHeader(http.StatusUnauthorized)
				if _, err := w.Write([]byte("failed to verify access token: " + err.Error())); err != nil {
					log.Error(err)
				}
				return
			}
			jwtToken, mapClaims, err := kc.ParseAccessToken(token, organization)
			if err != nil {
				log.Error("failed to parse access token: ", err)
				w.WriteHeader(http.StatusInternalServerError)
				if _, err := w.Write([]byte(err.Error())); err != nil {
					log.Error(err)
				}
				return
			}

			if jwtToken == nil || mapClaims == nil || mapClaims.Valid() != nil {
				w.WriteHeader(http.StatusUnauthorized)
				if _, err := w.Write([]byte("Error message TODO")); err != nil {
					log.Error(err)
				}
				return
			}
			roleProjectMapping := make(map[string]string)
			for _, role := range jwtToken.Claims.(jwt.MapClaims)["tks-role"].([]interface{}) {
				slice := strings.Split(role.(string), "@")
				if len(slice) != 2 {
					log.Error("invalid role format: ", role)
					w.WriteHeader(http.StatusInternalServerError)
					if _, err := w.Write([]byte(fmt.Sprintf("invalid role format: %s", role))); err != nil {
						log.Error(err)
					}
					return
				}
				// key is projectName and value is roleName
				roleProjectMapping[slice[1]] = slice[0]
			}
			userId, err := uuid.Parse(jwtToken.Claims.(jwt.MapClaims)["sub"].(string))
			if err != nil {
				userId = uuid.Nil
			}

			userInfo := &user.DefaultInfo{
				OrganizationId:     jwtToken.Claims.(jwt.MapClaims)["organization"].(string),
				UserId:             userId,
				RoleProjectMapping: roleProjectMapping,
			}

			r = r.WithContext(request.WithToken(r.Context(), token))
			r = r.WithContext(request.WithUser(r.Context(), userInfo))

			next.ServeHTTP(w, r)
		}

	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info(fmt.Sprintf("***** START [%s %s] ***** ", r.Method, r.RequestURI))

		body, err := io.ReadAll(r.Body)
		if err == nil {
			log.Info(fmt.Sprintf("body : %s", bytes.NewBuffer(body).String()))
		}
		r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

		next.ServeHTTP(w, r)

		log.Info("***** END *****")
	})
}

/*
func transactionMiddleware(db *gorm.DB) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			txHandle := db.Begin()
			log.Debug("beginning database transaction")

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
				log.Debug("committing transactions")
				if err := txHandle.Commit().Error; err != nil {
					log.Debug("trx commit error: ", err)
				}
			} else {
				log.Debug("rolling back transaction due to status code: ", recorder.Status)
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
