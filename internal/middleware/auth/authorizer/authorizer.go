package authorizer

import (
	"github.com/openinfradev/tks-api/pkg/log"
	"net/http"
)

type Interface interface {
	WithAuthorization(handler http.Handler) http.Handler
}

type defaultAuthorization struct {
}

func NewDefaultAuthorization() *defaultAuthorization {
	return &defaultAuthorization{}
}

func (a *defaultAuthorization) WithAuthorization(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info("called Authorization")

		handler.ServeHTTP(w, r)
	})
}
