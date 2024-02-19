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

type fnAudit = func(out *bytes.Buffer, in []byte, statusCode int) (message string, description string)

func NewDefaultAudit(repo repository.Repository) *defaultAudit {
	return &defaultAudit{
		repo: repo.Audit,
	}
}

// WRITE LOGIC HERE
var auditMap = map[internalApi.Endpoint]fnAudit{
	internalApi.CreateStack: func(out *bytes.Buffer, in []byte, statusCode int) (message string, description string) {
		input := domain.CreateStackRequest{}
		if err := json.Unmarshal(in, &input); err != nil {
			log.Error(err)
		}

		if statusCode >= 200 && statusCode < 300 {
			return fmt.Sprintf("스택 [%s]을 생성하였습니다.", input.Name), ""
		} else {
			var e httpErrors.RestError
			if err := json.NewDecoder(out).Decode(&e); err != nil {
				log.Error(err)
			}
			return fmt.Sprintf("스택 [%s]을 생성하는데 실패하였습니다.", input.Name), e.Text()
		}
	}, internalApi.CreateProject: func(out *bytes.Buffer, in []byte, statusCode int) (message string, description string) {
		input := domain.CreateProjectRequest{}
		_ = json.Unmarshal(in, &input)

		if statusCode >= 200 && statusCode < 300 {
			return fmt.Sprintf("프로젝트 [%s]를 생성하였습니다.", input.Name), ""
		} else {
			var e httpErrors.RestError
			if err := json.NewDecoder(out).Decode(&e); err != nil {
				log.Error(err)
			}
			return "프로젝트 [%s]를 생성하는데 실패하였습니다. ", e.Text()
		}
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

		requestBody := &bytes.Buffer{}
		_, _ = io.Copy(requestBody, r.Body)

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
			body, err := io.ReadAll(requestBody)
			if err != nil {
				log.Error(err)
			}
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
