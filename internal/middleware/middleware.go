package middleware

import (
	"net/http"

	internalApi "github.com/openinfradev/tks-api/internal/delivery/api"
	"github.com/openinfradev/tks-api/internal/middleware/audit"
	"github.com/openinfradev/tks-api/internal/middleware/auth/authenticator"
	"github.com/openinfradev/tks-api/internal/middleware/auth/authorizer"
	"github.com/openinfradev/tks-api/internal/middleware/auth/requestRecoder"
)

type Middleware struct {
	authenticator  authenticator.Interface
	authorizer     authorizer.Interface
	requestRecoder requestRecoder.Interface
	audit          audit.Interface
}

func NewMiddleware(authenticator authenticator.Interface,
	authorizer authorizer.Interface,
	requestRecoder requestRecoder.Interface,
	audit audit.Interface) *Middleware {
	ret := &Middleware{
		authenticator:  authenticator,
		authorizer:     authorizer,
		requestRecoder: requestRecoder,
		audit:          audit,
	}
	return ret
}

func (m *Middleware) Handle(endpoint internalApi.Endpoint, handle http.Handler) http.Handler {

	// pre-handler
	preHandler := m.authorizer.WithAuthorization(handle)
	preHandler = m.requestRecoder.WithRequestRecoder(endpoint, preHandler)
	preHandler = m.authenticator.WithAuthentication(preHandler)
	preHandler = m.audit.WithAudit(endpoint, preHandler)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		preHandler.ServeHTTP(w, r)
	})
}
