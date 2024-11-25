package logging

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal"
	"github.com/openinfradev/tks-api/pkg/log"
)

const MAX_LOG_LEN = 1000

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		r = r.WithContext(context.WithValue(ctx, internal.ContextKeyRequestID, uuid.New().String()))

		log.Infof(r.Context(), fmt.Sprintf("***** START [%s %s] ***** ", r.Method, r.RequestURI))

		body, err := io.ReadAll(r.Body)
		if err == nil {
			log.Infof(r.Context(), fmt.Sprintf("REQUEST BODY : %v", bytes.NewBuffer(body)))
			log.Infof(r.Context(), fmt.Sprintf("REQUEST BODY : %v", bytes.NewBuffer(body).String()))
		}
		r.Body = io.NopCloser(bytes.NewBuffer(body))
		lrw := NewLoggingResponseWriter(w)

		next.ServeHTTP(lrw, r)

		statusCode := lrw.GetStatusCode()

		if len(lrw.GetBody().String()) > MAX_LOG_LEN {
			log.Infof(r.Context(), "[API_RESPONSE] [%d][%s][%s]", statusCode, http.StatusText(statusCode), lrw.GetBody().String()[:MAX_LOG_LEN-1])
		} else {
			log.Infof(r.Context(), "[API_RESPONSE] [%d][%s][%s]", statusCode, http.StatusText(statusCode), lrw.GetBody().String())
		}
		log.Infof(r.Context(), "***** END [%s %s] *****", r.Method, r.RequestURI)
	})
}
