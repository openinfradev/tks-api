package authorizer

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/internal"
	internalApi "github.com/openinfradev/tks-api/internal/delivery/api"
	internalHttp "github.com/openinfradev/tks-api/internal/delivery/http"
	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
	internalRole "github.com/openinfradev/tks-api/internal/middleware/auth/role"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
)

func RBACFilter(handler http.Handler, repo repository.Repository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestUserInfo, ok := request.UserFrom(r.Context())
		if !ok {
			internalHttp.ErrorJSON(w, r, httpErrors.NewInternalServerError(fmt.Errorf("user not found"), "", ""))
			return
		}
		organizationRole := requestUserInfo.GetRoleOrganizationMapping()[requestUserInfo.GetOrganizationId()]

		// TODO: 추후 tks-admin role 수정 필요
		if organizationRole == "tks-admin" {
			handler.ServeHTTP(w, r)
			return
		}

		vars := mux.Vars(r)
		// Organization Filter
		if organizationRole == "admin" || organizationRole == "user" {
			if orgId, ok := vars["organizationId"]; ok {
				if orgId != requestUserInfo.GetOrganizationId() {
					internalHttp.ErrorJSON(w, r, httpErrors.NewForbiddenError(fmt.Errorf("permission denied"), "", ""))
					return
				}
			} else {
				log.Warn("RBACFilter: organizationId not found. Passing through unsafely.")
			}
		}

		// User Resource Filter
		if strings.HasPrefix(r.URL.Path, internal.API_PREFIX+internal.API_VERSION+"/organizations/"+requestUserInfo.GetOrganizationId()+"/user") {
			switch r.Method {
			case http.MethodPost, http.MethodPut, http.MethodDelete:
				if organizationRole != "admin" {
					internalHttp.ErrorJSON(w, r, httpErrors.NewForbiddenError(fmt.Errorf("permission denied"), "", ""))
					return
				}
			}
		}

		handler.ServeHTTP(w, r)
	})
}

func RBACFilterWithEndpoint(handler http.Handler, repo repository.Repository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestEndpointInfo, ok := request.EndpointFrom(r.Context())
		if !ok {
			internalHttp.ErrorJSON(w, r, httpErrors.NewInternalServerError(fmt.Errorf("endpoint not found"), "", ""))
			return
		}

		requestUserInfo, ok := request.UserFrom(r.Context())
		if !ok {
			internalHttp.ErrorJSON(w, r, httpErrors.NewInternalServerError(fmt.Errorf("user not found"), "", ""))
			return
		}

		// if requestEndpointInfo.String() is one of ProjectEndpoints, print true
		if internalApi.ApiMap[requestEndpointInfo].Group == "Project" && requestEndpointInfo != internalApi.CreateProject {
			log.Infof("RBACFilterWithEndpoint: %s is ProjectEndpoint", requestEndpointInfo.String())
			vars := mux.Vars(r)
			var projectRole string
			if projectId, ok := vars["projectId"]; ok {
				projectRole = requestUserInfo.GetRoleProjectMapping()[projectId]
			} else {
				log.Warn("RBACFilterWithEndpoint: projectId not found. Passing through unsafely.")
			}
			if !internalRole.IsRoleAllowed(requestEndpointInfo, internalRole.StrToRole(projectRole)) {
				internalHttp.ErrorJSON(w, r, httpErrors.NewForbiddenError(fmt.Errorf("permission denied"), "", ""))
				return

			}
		}

		handler.ServeHTTP(w, r)
	})
}

//type pair struct {
//	regexp string
//	method string
//}
//
//var LeaderPair = []pair{
//	{`/organizations/o[A-Za-z0-9]{8}/projects(?:\?.*)?$`, http.MethodPost},
//	{`/organizations/o[A-Za-z0-9]{8}/projects(?:\?.*)?$`, http.MethodGet},
//	{`/organizations/o[A-Za-z0-9]{8}/projects/p[A-Za-z0-9]{8}(?:\?.*)?$`, http.MethodGet},
//	{`/organizations/o[A-Za-z0-9]{8}/projects/p[A-Za-z0-9]{8}(?:\?.*)?$`, http.MethodPut},
//	{`/organizations/o[A-Za-z0-9]{8}/projects/p[A-Za-z0-9]{8}(?:\?.*)?$`, http.MethodDelete},
//	{`/organizations/o[A-Za-z0-9]{8}/projects/p[A-Za-z0-9]{8}/members(?:\?.*)?$`, http.MethodPost},
//	{`/organizations/o[A-Za-z0-9]{8}/projects/p[A-Za-z0-9]{8}/members(?:\?.*)?$`, http.MethodGet},
//	{`/organizations/o[A-Za-z0-9]{8}/projects/p[A-Za-z0-9]{8}/members/[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}(?:\?.*)?$`, http.MethodDelete},
//	{`/organizations/o[A-Za-z0-9]{8}/projects/p[A-Za-z0-9]{8}/members/[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}/role(?:\?.*)?$`, http.MethodPut},
//	{`/organizations/o[A-Za-z0-9]{8}/projects/p[A-Za-z0-9]{8}/namespace(?:\?.*)?$`, http.MethodPost},
//	{`/organizations/o[A-Za-z0-9]{8}/projects/p[A-Za-z0-9]{8}/namespace(?:\?.*)?$`, http.MethodGet},
//	{`/organizations/o[A-Za-z0-9]{8}/projects/p[A-Za-z0-9]{8}/namespace/n[A-Za-z0-9]{8}(?:\?.*)?$`, http.MethodGet},
//	{`/organizations/o[A-Za-z0-9]{8}/projects/p[A-Za-z0-9]{8}/namespace/n[A-Za-z0-9]{8}(?:\?.*)?$`, http.MethodDelete},
//}
//var roleApiMapper = make(map[string][]pair)
//
//func projectFilter(url string, method string, userInfo user.Info) bool {
//
//	return true
//}
