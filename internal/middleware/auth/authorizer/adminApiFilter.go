package authorizer

import (
	"fmt"
	internalHttp "github.com/openinfradev/tks-api/internal/delivery/http"
	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"net/http"
	"strings"
)

func AdminApiFilter(handler http.Handler, repo repository.Repository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestUserInfo, ok := request.UserFrom(r.Context())
		if !ok {
			internalHttp.ErrorJSON(w, r, httpErrors.NewInternalServerError(fmt.Errorf("user not found"), "", ""))
			return
		}

		endpointInfo, ok := request.EndpointFrom(r.Context())
		if !ok {
			internalHttp.ErrorJSON(w, r, httpErrors.NewInternalServerError(fmt.Errorf("endpoint not found"), "", ""))
			return
		}

		if strings.HasPrefix(endpointInfo.String(), "Admin_") {
			if requestUserInfo.GetOrganizationId() != "master" {
				internalHttp.ErrorJSON(w, r, httpErrors.NewForbiddenError(fmt.Errorf("permission denied"), "", ""))
				return
			}
		}

		handler.ServeHTTP(w, r)
	})
}
