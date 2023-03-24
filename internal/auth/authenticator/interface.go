package authenticator

import (
	"context"
	"github.com/openinfradev/tks-api/internal/auth/user"
	"net/http"
)

type Token interface {
	AuthenticateToken(ctx context.Context, token string) (*Response, bool, error)
}

type Request interface {
	AuthenticateRequest(req *http.Request) (*Response, bool, error)
}

type Response struct {
	User user.Info
}
