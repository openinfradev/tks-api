package audit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	internalApi "github.com/openinfradev/tks-api/internal/delivery/api"
	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
	"github.com/openinfradev/tks-api/internal/middleware/logging"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
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

type fnAudit = func(out *bytes.Buffer, in []byte, statusCode int) (message string, description string)

// WRITE LOGIC HERE
var auditMap = map[internalApi.Endpoint]fnAudit{
	internalApi.CreateStack: func(out *bytes.Buffer, in []byte, statusCode int) (message string, description string) {
		input := domain.CreateStackRequest{}
		_ = json.Unmarshal(in, &input)

		if statusCode >= 200 && statusCode < 300 {
			return fmt.Sprintf("스택 [%s]을 생성하였습니다.", input.Name), ""
		}

		var e httpErrors.RestError
		_ = json.NewDecoder(out).Decode(&e)
		return fmt.Sprintf("스택 [%s]을 생성하는데 실패하였습니다.", input.Name), e.Text()
	}, internalApi.CreateProject: func(out *bytes.Buffer, in []byte, statusCode int) (message string, description string) {
		input := domain.CreateProjectRequest{}
		_ = json.Unmarshal(in, &input)

		if statusCode >= 200 && statusCode < 300 {
			return fmt.Sprintf("프로젝트 [%s]를 생성하였습니다.", input.Name), ""
		}

		var e httpErrors.RestError
		_ = json.NewDecoder(out).Decode(&e)
		return "프로젝트 [%s]를 생성하는데 실패하였습니다. ", e.Text()
	},
}

func (a *defaultAudit) WithAudit(endpoint internalApi.Endpoint, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := request.UserFrom(r.Context())
		if !ok {
			log.Error("Invalid user token")
			return
		}
		userId := user.GetUserId()

		lrw := logging.NewLoggingResponseWriter(w)
		handler.ServeHTTP(lrw, r)
		statusCode := lrw.GetStatusCode()

		vars := mux.Vars(r)
		organizationId, ok := vars["organizationId"]
		if !ok {
			organizationId = user.GetOrganizationId()
		}

		message, description := "", ""
		if fn, ok := auditMap[endpoint]; ok {
			body, _ := io.ReadAll(r.Body)
			message, description = fn(lrw.GetBody(), body, statusCode)
		}

		dto := domain.Audit{
			OrganizationId: organizationId,
			Group:          internalApi.ApiMap[endpoint].Group,
			Message:        message,
			Description:    description,
			ClientIP:       getClientIpAddress(w, r),
			UserId:         &userId,
		}
		if _, err := a.repo.Create(dto); err != nil {
			log.Error(err)
		}
	})
}

var X_FORWARDED_FOR = "X-Forwarded-For"

func getClientIpAddress(w http.ResponseWriter, r *http.Request) string {
	xforward := r.Header.Get(X_FORWARDED_FOR)
	if xforward != "" {
		return xforward
	}

	clientAddr, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return clientAddr
	}
	return ""
}
