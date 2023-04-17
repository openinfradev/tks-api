package middleware

import (
	"github.com/openinfradev/tks-api/internal/middleware/auth/authenticator"
	"github.com/openinfradev/tks-api/internal/middleware/auth/authorization"
	"net/http"
)

type defaultMiddleware struct {
	authenticator authenticator.Interface
	authorizer    authorization.Interface
}

func NewDefaultMiddleware(authenticator authenticator.Interface,
	authorizer authorization.Interface) *defaultMiddleware {
	ret := &defaultMiddleware{
		authenticator: authenticator,
		authorizer:    authorizer,
	}
	return ret
}

func (m *defaultMiddleware) Handle(handle http.Handler) http.Handler {
	handler := m.authorizer.WithAuthorization(handle)
	handler = m.authenticator.WithAuthentication(handler)
	return handler
}
