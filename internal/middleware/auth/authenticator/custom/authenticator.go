package custom

import (
	"fmt"
	"github.com/openinfradev/tks-api/internal/helper"
	"github.com/openinfradev/tks-api/internal/middleware/auth/authenticator"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
	"net/http"
	"strings"
)

type CustomAuthenticator struct {
	repo repository.Repository
}

func NewCustomAuthenticator(repo repository.Repository) *CustomAuthenticator {
	return &CustomAuthenticator{
		repo: repo,
	}
}
func (a *CustomAuthenticator) AuthenticateRequest(r *http.Request) (*authenticator.Response, bool, error) {
	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if authHeader == "" {
		return nil, false, fmt.Errorf("authorizer header is invalid")
	}
	parts := strings.SplitN(authHeader, " ", 3)
	if len(parts) < 2 || strings.ToLower(parts[0]) != "bearer" {
		return nil, false, fmt.Errorf("authorizer header is invalid")
	}
	token := parts[1]

	if len(token) == 0 {
		// The space before the token case
		if len(parts) == 3 {
			log.Warn(r.Context(), "the provided Authorization header contains extra space before the bearer token, and is ignored")
		}
		return nil, false, fmt.Errorf("token is empty")
	}

	parsedToken, err := helper.StringToTokenWithoutVerification(token)
	if err != nil {
		return nil, false, httpErrors.NewUnauthorizedError(err, "A_INVALID_TOKEN", "토큰이 유효하지 않습니다.")
	}

	claims, err := helper.RetrieveClaims(parsedToken)
	if err != nil {
		return nil, false, httpErrors.NewUnauthorizedError(err, "A_INVALID_TOKEN", "토큰이 유효하지 않습니다.")
	}

	organizationId, ok := claims["organization"].(string)
	if !ok {
		return nil, false, httpErrors.NewUnauthorizedError(fmt.Errorf("organizationId is not found"), "A_INVALID_TOKEN", "토큰이 유효하지 않습니다.")
	}

	userId, ok := claims["sub"].(string)
	if !ok {
		return nil, false, httpErrors.NewUnauthorizedError(fmt.Errorf("userId is not found"), "A_INVALID_TOKEN", "토큰이 유효하지 않습니다.")
	}

	expiredTime, err := a.repo.Auth.GetExpiredTimeOnToken(r.Context(), organizationId, userId)
	if expiredTime == nil {
		return nil, true, nil
	}
	if err != nil {
		return nil, false, httpErrors.NewUnauthorizedError(err, "A_INVALID_TOKEN", "토큰이 유효하지 않습니다.")
	}

	iat, ok := claims["iat"].(float64)
	if !ok {
		return nil, false, httpErrors.NewUnauthorizedError(fmt.Errorf("iat is not found"), "A_INVALID_TOKEN", "토큰이 유효하지 않습니다.")
	}

	if int64(iat) < expiredTime.ExpiredTime.Unix() {
		return nil, false, httpErrors.NewUnauthorizedError(fmt.Errorf("token is changed"), "A_UNUSABLE_TOKEN", "토큰이 변경되었습니다.")
	}

	return nil, true, nil
}
