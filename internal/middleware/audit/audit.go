package audit

import (
	"fmt"
	"net"
	"net/http"

	internalApi "github.com/openinfradev/tks-api/internal/delivery/api"
	"github.com/openinfradev/tks-api/internal/middleware/logging"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/pkg/log"
)

type Interface interface {
	WithAudit(endpoint internalApi.Endpoint, handler http.Handler) http.Handler
}

type defaultAudit struct {
	repo repository.IAuditRepository
}

func NewDefaultAudit(repo repository.Repository) *defaultAudit {
	return &defaultAudit{
		repo: repo.Audit,
	}
}

func (a *defaultAudit) WithAudit(endpoint internalApi.Endpoint, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		log.InfoWithContext(ctx, endpoint)

		GetIpAddress(w, r)

		lrw := logging.NewLoggingResponseWriter(w)
		handler.ServeHTTP(lrw, r)

		statusCode := lrw.GetStatusCode()

		log.Infof("%v", endpoint)
		log.Infof("%v", internalApi.CreateStack)

		// check & matching
		if statusCode >= 200 && statusCode < 300 {
			if endpoint == internalApi.CreateStack {
				log.Info("스택을 생성하였습니다.")
			}
		}
	})
}

func GetIpAddress(w http.ResponseWriter, r *http.Request) {
	clientAddr, _, err := net.SplitHostPort(r.RemoteAddr)
	log.Info(err)
	log.Info(clientAddr)

	xforward := r.Header.Get("X-Forwarded-For")
	fmt.Println("X-Forwarded-For : ", xforward)
}
