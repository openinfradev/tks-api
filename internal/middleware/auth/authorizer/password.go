package authorizer

import (
	"fmt"
	"net/http"

	"github.com/openinfradev/tks-api/internal"
	internalHttp "github.com/openinfradev/tks-api/internal/delivery/http"
	"github.com/openinfradev/tks-api/internal/helper"
	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
)

func PasswordFilter(handler http.Handler, repo repository.Repository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestUserInfo, ok := request.UserFrom(r.Context())
		if !ok {
			internalHttp.ErrorJSON(w, httpErrors.NewInternalServerError(fmt.Errorf("user not found"), "", ""))
			return
		}

		storedUser, err := repo.User.GetByUuid(requestUserInfo.GetUserId())
		if err != nil {
			internalHttp.ErrorJSON(w, err)
			return
		}
		//TODO: 임시로 admin 계정은 비밀번호 변경 기간을 무시하도록 함. 추후 설계 필요
		if storedUser.AccountId == "admin" {
			handler.ServeHTTP(w, r)
			return
		}
		if helper.IsDurationExpired(storedUser.PasswordUpdatedAt, internal.PasswordExpiredDuration) {
			allowedUrl := []string{
				internal.API_PREFIX + internal.API_VERSION + "/organizations/" + requestUserInfo.GetOrganizationId() + "/my-profile" + "/password",
				internal.API_PREFIX + internal.API_VERSION + "/organizations/" + requestUserInfo.GetOrganizationId() + "/my-profile" + "/next-password-change",
				internal.API_PREFIX + internal.API_VERSION + "/auth/logout",
			}
			if !(urlContains(allowedUrl, r.URL.Path) && r.Method == http.MethodPut) {
				internalHttp.ErrorJSON(w, httpErrors.NewForbiddenError(fmt.Errorf("password expired"), "", ""))
				return
			}
		}
		handler.ServeHTTP(w, r)
	})
}

func urlContains(urls []string, url string) bool {
	for _, u := range urls {
		if u == url {
			return true
		}
	}
	return false
}
