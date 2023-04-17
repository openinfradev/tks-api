package auth

import (
	"net/http"
)

type Interface interface {
	WithAuthentication(handler http.Handler) http.Handler
	WithAuthorization(handler http.Handler) http.Handler
}
