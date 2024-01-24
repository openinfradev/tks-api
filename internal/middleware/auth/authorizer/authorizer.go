package authorizer

import (
	"github.com/openinfradev/tks-api/internal/repository"
	"net/http"
)

type Interface interface {
	WithAuthorization(handler http.Handler) http.Handler
}
type filterFunc func(http.Handler, repository.Repository) http.Handler
type defaultAuthorization struct {
	repo    repository.Repository
	filters []filterFunc
}

func NewDefaultAuthorization(repo repository.Repository) *defaultAuthorization {
	d := &defaultAuthorization{
		repo: repo,
	}
	d.addFilters(PasswordFilter)
	d.addFilters(RBACFilter)
	d.addFilters(RBACFilterWithEndpoint)

	return d
}

func (a *defaultAuthorization) WithAuthorization(handler http.Handler) http.Handler {
	compositeFilter := combineFilters(a.filters...)
	return compositeFilter(handler, a.repo)
}

func (a *defaultAuthorization) addFilters(filters ...filterFunc) {
	a.filters = append(a.filters, filters...)
}

func combineFilters(filters ...filterFunc) filterFunc {
	return func(handler http.Handler, repo repository.Repository) http.Handler {
		for _, filter := range filters {
			handler = filter(handler, repo)
		}
		return handler
	}
}
