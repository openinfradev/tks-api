package authenticator

import (
	"fmt"
	"github.com/openinfradev/tks-api/internal/repository"
	"net/http"

	internalHttp "github.com/openinfradev/tks-api/internal/delivery/http"
	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
	"github.com/openinfradev/tks-api/internal/middleware/auth/user"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
)

type Interface interface {
	WithAuthentication(handler http.Handler) http.Handler
}
type Request interface {
	AuthenticateRequest(req *http.Request) (*Response, bool, error)
}

type defaultAuthenticator struct {
	kcAuth     Request
	customAuth Request
	repo       repository.Repository
}

func NewAuthenticator(kc Request, repo repository.Repository, c Request) *defaultAuthenticator {
	return &defaultAuthenticator{
		kcAuth:     kc,
		repo:       repo,
		customAuth: c,
	}
}

func (a *defaultAuthenticator) WithAuthentication(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp, ok, err := a.kcAuth.AuthenticateRequest(r)
		if !ok {
			log.Error(r.Context(), err)
			internalHttp.ErrorJSON(w, r, err)
			return
		}
		if err != nil {
			internalHttp.ErrorJSON(w, r, err)
			return
		}

		_, ok, err = a.customAuth.AuthenticateRequest(r)
		if !ok {
			internalHttp.ErrorJSON(w, r, err)
			return
		}

		r = r.WithContext(request.WithUser(r.Context(), resp.User))

		_, ok = request.UserFrom(r.Context())
		if !ok {
			internalHttp.ErrorJSON(w, r, httpErrors.NewInternalServerError(fmt.Errorf("user not found"), "", ""))
			return
		}
		_, ok = request.TokenFrom(r.Context())
		if !ok {
			internalHttp.ErrorJSON(w, r, httpErrors.NewInternalServerError(fmt.Errorf("token not found"), "", ""))
			return
		}
		_, ok = request.SessionFrom(r.Context())
		if !ok {
			internalHttp.ErrorJSON(w, r, httpErrors.NewInternalServerError(fmt.Errorf("session not found"), "", ""))
			return
		}
		handler.ServeHTTP(w, r)
	})
}

type Response struct {
	User user.Info
}
