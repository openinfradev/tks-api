package request

import (
	"context"
	"github.com/openinfradev/tks-api/internal/auth/user"
)

type key int

const (
	userKey key = iota
	userToken
	organinzation
)

func WithValue(parent context.Context, key, val interface{}) context.Context {
	return context.WithValue(parent, key, val)
}
func WithUser(parent context.Context, user user.Info) context.Context {
	return WithValue(parent, userKey, user)
}

// UserFrom function to retrieve user from context. If there is no user in context, it returns false
func UserFrom(ctx context.Context) (user.Info, bool) {
	user, ok := ctx.Value(userKey).(user.Info)
	return user, ok
}

func WithToken(parent context.Context, token string) context.Context {
	return WithValue(parent, userToken, token)
}

func TokenFrom(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(userToken).(string)
	return token, ok
}
