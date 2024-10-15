package requestRecoder

import (
	"fmt"
	internalApi "github.com/openinfradev/tks-api/internal/delivery/api"
	internalHttp "github.com/openinfradev/tks-api/internal/delivery/http"
	"github.com/openinfradev/tks-api/internal/middleware/auth/request"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"net/http"
)

type Interface interface {
	WithRequestRecoder(endpoint internalApi.Endpoint, handler http.Handler) http.Handler
}

type defaultRequestRecoder struct {
}

func NewDefaultRequestRecoder() *defaultRequestRecoder {
	return &defaultRequestRecoder{}
}

func (a *defaultRequestRecoder) WithRequestRecoder(endpoint internalApi.Endpoint, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(request.WithEndpoint(r.Context(), endpoint))
		_, ok := request.EndpointFrom(r.Context())
		if !ok {
			internalHttp.ErrorJSON(w, r, httpErrors.NewInternalServerError(fmt.Errorf("endpoint not found"), "", ""))
			return
		}
		handler.ServeHTTP(w, r)
	})
}
