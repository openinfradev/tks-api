package audit

import (
	"bytes"
	"io"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	internalApi "github.com/openinfradev/tks-api/internal/delivery/api"
	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
	"github.com/openinfradev/tks-api/internal/middleware/logging"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/pkg/log"
)

type Interface interface {
	WithAudit(endpoint internalApi.Endpoint, handler http.Handler) http.Handler
}

type defaultAudit struct {
	repo     repository.IAuditRepository
	userRepo repository.IUserRepository
}

func NewDefaultAudit(repo repository.Repository) *defaultAudit {
	return &defaultAudit{
		repo:     repo.Audit,
		userRepo: repo.User,
	}
}

func (a *defaultAudit) WithAudit(endpoint internalApi.Endpoint, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := request.UserFrom(r.Context())
		if !ok {
			log.Error(r.Context(), "Invalid user token")
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
			// workarround pingtoken
			if endpoint != internalApi.VerifyToken {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					log.Error(r.Context(), err)
				}
				message, description = fn(r.Context(), lrw.GetBody().Bytes(), body, statusCode)
				r.Body = io.NopCloser(bytes.NewBuffer(body))

				u, err := a.userRepo.GetByUuid(r.Context(), userId)
				if err != nil {
					log.Error(r.Context(), err)
					return
				}

				userRoles := ""
				for i, role := range u.Roles {
					if i > 0 {
						userRoles = userRoles + ","
					}
					userRoles = userRoles + role.Name
				}

				dto := model.Audit{
					OrganizationId:   organizationId,
					OrganizationName: u.Organization.Name,
					Group:            internalApi.ApiMap[endpoint].Group,
					Message:          message,
					Description:      description,
					ClientIP:         GetClientIpAddress(w, r),
					UserId:           &u.ID,
					UserAccountId:    u.AccountId,
					UserName:         u.Name,
					UserRoles:        userRoles,
				}
				if _, err := a.repo.Create(r.Context(), dto); err != nil {
					log.Error(r.Context(), err)
				}
			}
		}

	})
}

var X_FORWARDED_FOR = "X-Forwarded-For"

func GetClientIpAddress(w http.ResponseWriter, r *http.Request) string {
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
