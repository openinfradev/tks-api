package middleware

import (
	internalApi "github.com/openinfradev/tks-api/internal/delivery/api"
	"github.com/openinfradev/tks-api/internal/middleware/auth/authenticator"
	"github.com/openinfradev/tks-api/internal/middleware/auth/authorizer"
	"github.com/openinfradev/tks-api/internal/middleware/auth/requestRecoder"
	"net/http"
)

type Middleware struct {
	authenticator  authenticator.Interface
	authorizer     authorizer.Interface
	requestRecoder requestRecoder.Interface
}

func NewMiddleware(authenticator authenticator.Interface,
	authorizer authorizer.Interface,
	requestRecoder requestRecoder.Interface) *Middleware {
	ret := &Middleware{
		authenticator:  authenticator,
		authorizer:     authorizer,
		requestRecoder: requestRecoder,
	}
	return ret
}

func (m *Middleware) Handle(endpoint internalApi.Endpoint, handle http.Handler) http.Handler {
	handler := m.authorizer.WithAuthorization(handle)
	handler = m.requestRecoder.WithRequestRecoder(endpoint, handler)
	handler = m.authenticator.WithAuthentication(handler)
	return handler
}
