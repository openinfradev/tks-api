package audit

import (
	"github.com/openinfradev/tks-api/internal/repository"
	"net/http"
)

type Interface interface {
	WithAudit(handler http.Handler) http.Handler
}

type defaultAudit struct {
	repo repository.Repository
}

func NewDefaultAudit(repo repository.Repository) *defaultAudit {
	return &defaultAudit{
		repo: repo,
	}
}

// TODO: implement audit logic
func (a *defaultAudit) WithAudit(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: implement audit logic

		handler.ServeHTTP(w, r)
	})
}
