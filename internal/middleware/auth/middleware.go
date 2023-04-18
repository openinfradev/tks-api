package auth

import (
	"github.com/openinfradev/tks-api/internal/middleware/auth/authenticator"
	"github.com/openinfradev/tks-api/internal/middleware/auth/authorizer"
	"net/http"
)

type authMiddleware struct {
	authenticator authenticator.Interface
	authorizer    authorizer.Interface
}

func NewAuthMiddleware(authenticator authenticator.Interface,
	authorizer authorizer.Interface) *authMiddleware {
	ret := &authMiddleware{
		authenticator: authenticator,
		authorizer:    authorizer,
	}
	return ret
}

func (m *authMiddleware) Handle(handle http.Handler) http.Handler {
	handler := m.authorizer.WithAuthorization(handle)
	handler = m.authenticator.WithAuthentication(handler)
	return handler
}
