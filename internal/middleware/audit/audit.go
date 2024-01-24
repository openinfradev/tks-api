package audit

import (
	"net/http"

	internalApi "github.com/openinfradev/tks-api/internal/delivery/api"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/pkg/log"
)

type Interface interface {
	WithAudit(endpoint internalApi.Endpoint, handler http.Handler) http.Handler
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
func (a *defaultAudit) WithAudit(endpoint internalApi.Endpoint, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: implement audit logic

		log.Info(endpoint)
		log.Info("KLKK")

		handler.ServeHTTP(w, r)
	})
}
