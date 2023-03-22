package request

import (
	"context"
	"github.com/openinfradev/tks-api/internal/auth/user"
)

type key int

const (
	userKey key = iota
	userToken
)

func WithValue(parent context.Context, key, val interface{}) context.Context {
	return context.WithValue(parent, key, val)
}
func WithUser(parent context.Context, user user.Info) context.Context {
	return WithValue(parent, userKey, user)
}

func WithToken(parent context.Context, token string) context.Context {
	return WithValue(parent, userToken, token)
}

func TokenFrom(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(userToken).(string)
	return token, ok
}

func UserFrom(ctx context.Context) (user.Info, bool) {
	user, ok := ctx.Value(userKey).(user.Info)
	return user, ok
}
