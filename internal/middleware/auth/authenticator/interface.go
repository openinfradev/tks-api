package authenticator

import (
	internalHttp "github.com/openinfradev/tks-api/internal/delivery/http"
	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
	"github.com/openinfradev/tks-api/internal/middleware/auth/user"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
	"net/http"
)

type Interface interface {
	WithAuthentication(handler http.Handler) http.Handler
}

type defaultAuthenticator struct {
	auth Request
}

func NewDefaultAuthenticator(auth Request) *defaultAuthenticator {
	return &defaultAuthenticator{
		auth: auth,
	}
}

func (a *defaultAuthenticator) WithAuthentication(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		log.Info("called Authentication")

		resp, ok, err := a.auth.AuthenticateRequest(r)
		if !ok {
			internalHttp.ErrorJSON(w, httpErrors.NewUnauthorizedError(err))
			return
		}
		log.Info(request.TokenFrom(r.Context()))
		r = r.WithContext(request.WithUser(r.Context(), resp.User))
		handler.ServeHTTP(w, r)
	})
}

type Request interface {
	AuthenticateRequest(req *http.Request) (*Response, bool, error)
}

type Response struct {
	User user.Info
}