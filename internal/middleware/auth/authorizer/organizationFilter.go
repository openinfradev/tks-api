package authorizer

import (
	"fmt"
	"github.com/gorilla/mux"
	internalHttp "github.com/openinfradev/tks-api/internal/delivery/http"
	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
	"net/http"
)

func OrganizationFilter(handler http.Handler, repo repository.Repository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestUserInfo, ok := request.UserFrom(r.Context())
		if !ok {
			internalHttp.ErrorJSON(w, r, httpErrors.NewInternalServerError(fmt.Errorf("user not found"), "", ""))
			return
		}

		if requestUserInfo.GetOrganizationId() != "" && requestUserInfo.GetOrganizationId() == "master" {
			handler.ServeHTTP(w, r)
			return
		}

		vars := mux.Vars(r)
		requestedOrganization, ok := vars["organizationId"]
		if !ok {
			log.Warn(r.Context(), "OrganizationFilter: organizationId not found. Passing through unsafely.")
			handler.ServeHTTP(w, r)
		}

		if requestedOrganization != requestUserInfo.GetOrganizationId() {
			log.Debugf(r.Context(), "OrganizationFilter: requestedOrganization: %s, userOrganization: %s", requestedOrganization, requestUserInfo.GetOrganizationId())
			internalHttp.ErrorJSON(w, r, httpErrors.NewForbiddenError(fmt.Errorf("permission denied"), "", ""))
		}

		handler.ServeHTTP(w, r)
	})
}
