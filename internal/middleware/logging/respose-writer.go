package logging

import (
	"bytes"
	"net/http"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
}

func NewLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	var buf bytes.Buffer
	return &loggingResponseWriter{w, http.StatusOK, buf}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Write(buf []byte) (int, error) {
	lrw.body.Write(buf)
	return lrw.ResponseWriter.Write(buf)
}

func (lrw *loggingResponseWriter) GetBody() *bytes.Buffer {
	return &lrw.body
}

func (lrw *loggingResponseWriter) GetStatusCode() int {
	return lrw.statusCode
}
