package authorizer

import (
	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/log"
	"net/http"
	"strings"
)

var (
	userBasedAccessControl []func(r *http.Request) bool
)

func init() {
	userBasedAccessControl = append(userBasedAccessControl, OpaGatekeeper)
}

func ABACFilter(handler http.Handler, repo repository.Repository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, f := range userBasedAccessControl {
			if !f(r) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
		}

		handler.ServeHTTP(w, r)
	})
}

func OpaGatekeeper(r *http.Request) bool {
	requestUserInfo, ok := request.UserFrom(r.Context())
	if !ok {
		log.Errorf(r.Context(), "user not found")
		return false
	}

	if strings.HasSuffix(requestUserInfo.GetAccountId(), string(usecase.OPAGatekeeperReservedAccountIdSuffix)) {
		// Allow restricted API from OPA Gatekeeper
	}

	return true
}
