package authorizer

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/internal"
	internalHttp "github.com/openinfradev/tks-api/internal/delivery/http"
	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
	"net/http"
	"strings"
)

func RBACFilter(handler http.Handler, repo repository.Repository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestUserInfo, ok := request.UserFrom(r.Context())
		if !ok {
			internalHttp.ErrorJSON(w, httpErrors.NewInternalServerError(fmt.Errorf("user not found")))
			return
		}
		role := requestUserInfo.GetRoleProjectMapping()[requestUserInfo.GetOrganizationId()]

		// TODO: 추후 tks-admin role 수정 필요
		if role == "tks-admin" {
			handler.ServeHTTP(w, r)
			return
		}

		vars := mux.Vars(r)
		// Organization Filter
		if role == "admin" || role == "user" {
			if orgId, ok := vars["organizationId"]; ok {
				if orgId != requestUserInfo.GetOrganizationId() {
					internalHttp.ErrorJSON(w, httpErrors.NewForbiddenError(fmt.Errorf("permission denied")))
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
				if role != "admin" {
					internalHttp.ErrorJSON(w, httpErrors.NewForbiddenError(fmt.Errorf("permission denied")))
					return
				}
			}
		}

		handler.ServeHTTP(w, r)
	})
}
